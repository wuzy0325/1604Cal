<template>
  <div class="device-panel">
    <div class="panel-header">
      <div class="device-info">
        <el-icon class="device-icon"><Cpu /></el-icon>
        <div>
          <div class="device-name">1604设备</div>
          <div class="device-type">计量采集设备</div>
        </div>
      </div>
      <DeviceStatusBadge :status="status" />
    </div>

    <div class="connection-control">
      <div class="input-group">
        <span class="prefix">TCP://</span>
        <el-input v-model="ip" placeholder="192.168.1.100" />
      </div>
      <el-button
        :type="status === 'connected' ? 'danger' : 'primary'"
        @click="toggleConnection"
      >
        {{ status === 'connected' ? '断开' : '连接' }}
      </el-button>
    </div>

    <div v-if="status === 'connected'" class="device-status">
      <div class="status-row">
        <span class="label">阀门状态:</span>
        <el-tag :type="valveStatus === 'open' ? 'success' : 'info'" size="small">
          {{ valveStatus === 'open' ? '开启' : '关闭' }}
        </el-tag>
      </div>
      <div class="status-row">
        <span class="label">单位类型:</span>
        <span class="value">kPa</span>
      </div>
      <div v-if="needCalibration" class="warning">
        <el-icon><Warning /></el-icon>
        <span>设备需要校准</span>
      </div>
    </div>

    <div v-if="status === 'connected'" class="valve-control">
      <el-button
        :type="valveStatus === 'open' ? 'primary' : 'default'"
        @click="valveStatus = 'open'"
      >
        打开阀门
      </el-button>
      <el-button
        :type="valveStatus === 'close' ? 'primary' : 'default'"
        @click="valveStatus = 'close'"
      >
        关闭阀门
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Cpu, Warning } from '@element-plus/icons-vue'
import DeviceStatusBadge from '@/components/common/DeviceStatusBadge.vue'

const ip = ref('192.168.1.100')
const status = ref<'connected' | 'disconnected'>('disconnected')
const valveStatus = ref<'open' | 'close'>('close')
const needCalibration = ref(false)

const toggleConnection = () => {
  status.value = status.value === 'connected' ? 'disconnected' : 'connected'
}
</script>

<style scoped lang="scss">
.device-panel {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--spacing-md);

    .device-info {
      display: flex;
      align-items: center;
      gap: var(--spacing-sm);

      .device-icon {
        font-size: 24px;
        color: var(--accent-primary);
      }

      .device-name {
        color: var(--text-primary);
        font-weight: 500;
      }

      .device-type {
        color: var(--text-secondary);
        font-size: 12px;
      }
    }
  }

  .connection-control {
    display: flex;
    gap: var(--spacing-sm);
    margin-bottom: var(--spacing-md);

    .input-group {
      flex: 1;
      display: flex;
      align-items: center;
      background: var(--bg-tertiary);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-sm);
      padding: 0 var(--spacing-sm);

      .prefix {
        color: var(--text-muted);
        font-size: 12px;
        margin-right: var(--spacing-xs);
        white-space: nowrap;
      }

      .el-input {
        flex: 1;

        :deep(.el-input__wrapper) {
          background: transparent;
          box-shadow: none;
        }
      }
    }
  }

  .device-status {
    background: var(--bg-tertiary);
    border-radius: var(--radius-sm);
    padding: var(--spacing-md);
    margin-bottom: var(--spacing-md);

    .status-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: var(--spacing-sm);

      &:last-child {
        margin-bottom: 0;
      }

      .label {
        color: var(--text-secondary);
        font-size: 13px;
      }

      .value {
        color: var(--text-primary);
        font-size: 13px;
      }
    }

    .warning {
      display: flex;
      align-items: center;
      gap: var(--spacing-xs);
      color: var(--status-warning);
      font-size: 12px;
      margin-top: var(--spacing-sm);
      padding-top: var(--spacing-sm);
      border-top: 1px solid var(--border-color);
    }
  }

  .valve-control {
    display: flex;
    gap: var(--spacing-sm);

    .el-button {
      flex: 1;
    }
  }
}
</style>
