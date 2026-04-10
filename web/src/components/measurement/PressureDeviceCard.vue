<template>
  <div class="device-card">
    <div class="device-header">
      <span class="device-name">{{ device.name }}</span>
      <DeviceStatusBadge :status="device.status" />
    </div>
    <div class="device-info">
      <div class="info-row">
        <span class="label">IP:</span>
        <span class="value">{{ device.ip }}:{{ device.port }}</span>
      </div>
      <div class="info-row">
        <span class="label">当前压力:</span>
        <span class="value pressure">{{ device.currentPressure?.toFixed(2) || '--' }} {{ device.unit }}</span>
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

interface PressureDevice {
  id: string
  name: string
  ip: string
  port: number
  status: 'connected' | 'disconnected'
  currentPressure?: number
  unit: string
}

const props = defineProps<{
  device: PressureDevice
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
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  margin-bottom: var(--spacing-sm);
  
  .device-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--spacing-sm);
    
    .device-name {
      color: var(--text-primary);
      font-weight: 500;
    }
  }
  
  .device-info {
    margin-bottom: var(--spacing-sm);
    
    .info-row {
      display: flex;
      justify-content: space-between;
      margin-bottom: var(--spacing-xs);
      
      .label {
        color: var(--text-secondary);
        font-size: 12px;
      }
      
      .value {
        color: var(--text-primary);
        font-size: 12px;
        
        &.pressure {
          color: var(--accent-primary);
          font-weight: bold;
        }
      }
    }
  }
  
  .device-actions {
    display: flex;
    justify-content: flex-end;
  }
}
</style>
