<template>
  <span
    class="status-badge"
    :class="status"
  >
    {{ statusText }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  status: 'connected' | 'disconnected' | 'error'
}>()

const statusText = computed(() => {
  const map = {
    connected: '已连接',
    disconnected: '未连接',
    error: '错误'
  }
  return map[props.status]
})
</script>

<style scoped lang="scss">
.status-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 500;
  
  &.connected {
    background: rgba(16, 185, 129, 0.2);
    color: var(--status-success);
  }
  
  &.disconnected {
    background: rgba(160, 160, 160, 0.2);
    color: var(--text-secondary);
  }
  
  &.error {
    background: rgba(239, 68, 68, 0.2);
    color: var(--status-error);
  }
}
</style>
