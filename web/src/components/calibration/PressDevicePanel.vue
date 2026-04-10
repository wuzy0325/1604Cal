<template>
  <div class="device-panel">
    <div class="panel-header">
      <div class="device-info">
        <el-icon class="device-icon"><FirstAidKit /></el-icon>
        <div>
          <div class="device-name">打压设备</div>
          <div class="device-type">压力控制器</div>
        </div>
      </div>
      <DeviceStatusBadge :status="status" />
    </div>
    
    <div class="connection-control">
      <el-input v-model="ip" placeholder="IP地址" />
      <el-input-number v-model="port" :min="1" :max="65535" controls-position="right" />
      <el-button 
        :type="status === 'connected' ? 'danger' : 'primary'"
        @click="toggleConnection"
      >
        {{ status === 'connected' ? '断开' : '连接' }}
      </el-button>
    </div>
    
    <div v-if="status === 'connected'" class="pressure-control">
      <div class="current-pressure">
        <span class="label">当前压力:</span>
        <span class="value">{{ currentPressure.toFixed(2) }} kPa</span>
      </div>
      <div class="pressure-actions">
        <el-button @click="adjustPressure(-1)">
          <el-icon><ArrowDown /></el-icon>降压
        </el-button>
        <el-input-number v-model="targetPressure" :precision="2" :step="1" />
        <el-button @click="adjustPressure(1)">
          <el-icon><ArrowUp /></el-icon>升压
        </el-button>
      </div>
      <el-button type="primary" class="set-btn" @click="setPressure">
        设定压力
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { FirstAidKit, ArrowUp, ArrowDown } from '@element-plus/icons-vue'
import DeviceStatusBadge from '@/components/common/DeviceStatusBadge.vue'

const ip = ref('192.168.1.101')
const port = ref(502)
const status = ref<'connected' | 'disconnected'>('disconnected')
const currentPressure = ref(0)
const targetPressure = ref(100)

const toggleConnection = () => {
  status.value = status.value === 'connected' ? 'disconnected' : 'connected'
  if (status.value === 'connected') {
    currentPressure.value = 0
  }
}

const adjustPressure = (delta: number) => {
  targetPressure.value += delta * 10
}

const setPressure = () => {
  console.log(`设定压力: ${targetPressure.value}`)
  currentPressure.value = targetPressure.value
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
        color: var(--status-info);
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
    
    .el-input {
      flex: 2;
    }
    
    .el-input-number {
      flex: 1;
    }
  }
  
  .pressure-control {
    background: var(--bg-tertiary);
    border-radius: var(--radius-sm);
    padding: var(--spacing-md);
    
    .current-pressure {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: var(--spacing-md);
      
      .label {
        color: var(--text-secondary);
        font-size: 13px;
      }
      
      .value {
        color: var(--accent-primary);
        font-size: 20px;
        font-weight: bold;
      }
    }
    
    .pressure-actions {
      display: flex;
      align-items: center;
      gap: var(--spacing-sm);
      margin-bottom: var(--spacing-md);
      
      .el-input-number {
        flex: 1;
      }
    }
    
    .set-btn {
      width: 100%;
    }
  }
}
</style>
