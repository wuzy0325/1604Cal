<template>
  <el-dialog
    :model-value="visible"
    title="报警通道选择"
    width="420px"
    :close-on-click-modal="false"
    class="channel-select-dialog"
    @close="emit('close')"
  >
    <div class="channel-select-content">
      <div class="channel-select-tip">
        选择需要参与报警检测的通道，未选择的通道将不进行报警确认
      </div>
      <div class="channel-grid">
        <div
          v-for="ch in allChannels"
          :key="ch"
          class="channel-item"
          :class="{ selected: selected.includes(ch) }"
          @click="toggle(ch)"
        >
          <span class="channel-label">通道{{ ch + 1 }}</span>
          <el-icon v-if="selected.includes(ch)" class="check-icon"><Check /></el-icon>
        </div>
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="selectAll">全选</el-button>
        <el-button @click="deselectAll">全不选</el-button>
        <div class="footer-spacer" />
        <el-button @click="emit('close')">取消</el-button>
        <el-button type="primary" @click="emit('confirm', [...selected])">确定</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Check } from '@element-plus/icons-vue'

defineProps<{ visible: boolean }>()

const emit = defineEmits<{
  close: []
  confirm: [channels: number[]]
}>()

const allChannels = Array.from({ length: 16 }, (_, i) => i)
const selected = ref<number[]>([...allChannels])

function toggle(ch: number) {
  const idx = selected.value.indexOf(ch)
  if (idx >= 0) selected.value.splice(idx, 1)
  else selected.value.push(ch)
}

function selectAll() { selected.value = [...allChannels] }
function deselectAll() { selected.value = [] }
</script>

<style scoped lang="scss">
.channel-select-content {
  .channel-select-tip {
    color: var(--text-muted);
    font-size: 12px;
    margin-bottom: var(--spacing-md);
    line-height: 1.5;
  }
}

.channel-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--spacing-sm);
}

.channel-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.15s;
  user-select: none;

  &:hover {
    border-color: var(--accent-primary);
  }

  &.selected {
    border-color: var(--accent-primary);
    background: color-mix(in srgb, var(--accent-primary) 10%, transparent);
  }

  .channel-label {
    font-size: 13px;
    color: var(--text-primary);
  }

  .check-icon {
    color: var(--accent-primary);
    font-size: 16px;
  }
}

.dialog-footer {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);

  .footer-spacer {
    flex: 1;
  }
}
</style>
