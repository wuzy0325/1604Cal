<template>
  <div class="channel-matrix">
    <div class="matrix-header">
      <h4>通道选择</h4>
      <div class="actions">
        <span class="count">已选: {{ selectedCount }}/16</span>
        <el-button type="primary" link size="small" @click="selectAll">全选</el-button>
        <el-button type="danger" link size="small" @click="clearAll">清空</el-button>
      </div>
    </div>
    
    <div class="matrix-grid">
      <div 
        v-for="(selected, index) in localChannels" 
        :key="index"
        class="channel-item"
        :class="{ selected }"
        @click="toggleChannel(index)"
      >
        <el-checkbox v-model="localChannels[index]" @click.stop @change="emitUpdate">
          CH{{ index + 1 }}
        </el-checkbox>
      </div>
    </div>
    
    <div v-if="selectedCount === 0" class="warning">
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
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  
  .matrix-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--spacing-md);
    
    h4 {
      color: var(--text-primary);
      margin: 0;
    }
    
    .actions {
      display: flex;
      align-items: center;
      gap: var(--spacing-sm);
      
      .count {
        color: var(--text-secondary);
        font-size: 13px;
      }
    }
  }
  
  .matrix-grid {
    display: grid;
    grid-template-columns: repeat(8, 1fr);
    gap: var(--spacing-sm);
    
    @media (max-width: 1200px) {
      grid-template-columns: repeat(4, 1fr);
    }
    
    .channel-item {
      background: var(--bg-tertiary);
      border: 2px solid transparent;
      border-radius: var(--radius-sm);
      padding: var(--spacing-sm);
      cursor: pointer;
      transition: all 0.2s;
      
      &:hover {
        border-color: var(--accent-primary);
      }
      
      &.selected {
        background: rgba(16, 185, 129, 0.2);
        border-color: var(--status-success);
      }
      
      .el-checkbox {
        color: var(--text-primary);
        
        :deep(.el-checkbox__input.is-checked + .el-checkbox__label) {
          color: var(--status-success);
        }
      }
    }
  }
  
  .warning {
    display: flex;
    align-items: center;
    gap: var(--spacing-xs);
    color: var(--status-warning);
    font-size: 12px;
    margin-top: var(--spacing-md);
    padding: var(--spacing-sm);
    background: rgba(245, 158, 11, 0.1);
    border-radius: var(--radius-sm);
  }
}
</style>
