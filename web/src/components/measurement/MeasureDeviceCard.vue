<template>
  <div class="device-card">
    <div class="device-header">
      <span class="device-name">{{ device.name }}</span>
      <DeviceStatusBadge :status="device.status" />
    </div>
    <div class="device-info">
      <div class="info-row">
        <span class="label">型号:</span>
        <span class="value">{{ device.model }}</span>
      </div>
      <div class="info-row">
        <span class="label">通道数:</span>
        <span class="value">{{ device.channels }}</span>
      </div>
    </div>
    <div class="device-actions">
      <el-button 
        :type="device.status === 'connected' ? 'danger' : 'primary'"
        size="small"
        @click="toggleConnection"
      >
        {{ device.status === 'connected' ? '断开' : '连接' }}
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import DeviceStatusBadge from '@/components/common/DeviceStatusBadge.vue'
import type { MeasureDevice } from '@/stores/measurement/deviceStore'

const props = defineProps<{
  device: MeasureDevice
}>()

const emit = defineEmits<{
  connect: [id: string]
  disconnect: [id: string]
}>()

const toggleConnection = () => {
  if (props.device.status === 'connected') {
    emit('disconnect', props.device.id)
  } else {
    emit('connect', props.device.id)
  }
}
</script>

<style scoped lang="scss">
.device-card {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 3px;
  padding: var(--spacing-sm) var(--spacing-md);
  margin-bottom: var(--spacing-xs);

  .device-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--spacing-sm);

    .device-name {
      color: var(--text-primary);
      font-weight: 500;
      font-size: 13px;
    }
  }

  .device-info {
    margin-bottom: var(--spacing-sm);

    .info-row {
      display: flex;
      justify-content: space-between;
      margin-bottom: 2px;

      .label {
        color: var(--text-secondary);
        font-size: 11px;
      }

      .value {
        color: var(--text-primary);
        font-size: 11px;
      }
    }
  }

  .device-actions {
    display: flex;
    justify-content: flex-end;
  }
}
</style>
