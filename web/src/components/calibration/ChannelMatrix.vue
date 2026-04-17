<template>
  <div class="channel-matrix">
    <div class="matrix-header">
      <h4>通道选择</h4>
      <div class="actions">
        <span class="count">已选: {{ selectedCount }}/16</span>
        <el-button
          type="primary"
          link
          size="small"
          @click="selectAll"
        >
          全选
        </el-button>
        <el-button
          type="danger"
          link
          size="small"
          @click="clearAll"
        >
          清空
        </el-button>
      </div>
    </div>
    
    <div class="matrix-grid">
      <div 
        v-for="(selected, index) in localChannels" 
        :key="index"
        class="channel-item"
        :class="{ selected }"
        role="checkbox"
        :aria-checked="isSelected(index)"
        tabindex="0"
        @click="toggleChannel(index)"
        @keydown.enter="toggleChannel(index)"
        @keydown.space.prevent="toggleChannel(index)"
      >
        <el-checkbox
          v-model="localChannels[index]"
          @click.stop
          @change="emitUpdate"
        >
          CH{{ index + 1 }}
        </el-checkbox>
      </div>
    </div>
    
    <div
      v-if="selectedCount === 0"
      class="warning"
      role="alert"
    >
      <el-icon><Warning /></el-icon>
      <span>请至少选择一个通道</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Warning } from '@element-plus/icons-vue'

const props = defineProps<{
  selectedChannels?: number[]
}>()

const emit = defineEmits<{
  'update:selectedChannels': [channels: number[]]
}>()

// 本地状态 - 16个通道的选中状态
const localChannels = ref<boolean[]>(new Array(16).fill(false))

// 从props同步初始值
watch(() => props.selectedChannels, (newVal) => {
  if (newVal) {
    localChannels.value = new Array(16).fill(false).map((_, i) => newVal.includes(i + 1))
  }
}, { immediate: true })

// 计算选中的数量
const selectedCount = computed(() => localChannels.value.filter(Boolean).length)

// 计算选中的通道编号（1-16）
const selectedChannelNumbers = computed(() => {
  return localChannels.value
    .map((selected, index) => selected ? index + 1 : null)
    .filter((num): num is number => num !== null)
})

// 判断通道是否选中
const isSelected = (index: number): boolean => {
  return localChannels.value[index]
}

// 切换通道
const toggleChannel = (index: number) => {
  localChannels.value[index] = !localChannels.value[index]
  emitUpdate()
}

// 全选
const selectAll = () => {
  localChannels.value.fill(true)
  emitUpdate()
}

// 清空
const clearAll = () => {
  localChannels.value.fill(false)
  emitUpdate()
}

// 触发更新事件
const emitUpdate = () => {
  emit('update:selectedChannels', selectedChannelNumbers.value)
}
</script>

<style scoped lang="scss">
.channel-matrix {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);

  .matrix-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    h4 {
      color: var(--text-primary);
      margin: 0;
      font-size: 12px;
      font-weight: 500;
    }

    .actions {
      display: flex;
      align-items: center;
      gap: var(--spacing-xs);

      .count {
        color: var(--text-muted);
        font-size: 11px;
      }
    }
  }

  .matrix-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 4px;

    .channel-item {
      background: var(--bg-tertiary);
      border: 1px solid transparent;
      border-radius: var(--radius-sm);
      padding: 5px 6px;
      cursor: pointer;
      transition: all 0.15s;

      &:hover {
        border-color: var(--accent-primary);
        background: var(--bg-quaternary);
      }

      &.selected {
        background: var(--status-success-bg);
        border-color: var(--status-success);
      }

      .el-checkbox {
        color: var(--text-primary);
        height: auto;

        :deep(.el-checkbox__label) {
          font-size: 11px;
          padding-left: 4px;
        }

        :deep(.el-checkbox__input.is-checked + .el-checkbox__label) {
          color: var(--status-success);
          font-weight: 500;
        }
      }
    }
  }

  .warning {
    display: flex;
    align-items: center;
    gap: var(--spacing-xs);
    color: var(--status-warning);
    font-size: 11px;
    padding: 6px var(--spacing-sm);
    background: var(--status-warning-bg-subtle);
    border-radius: var(--radius-sm);
  }
}
</style>
