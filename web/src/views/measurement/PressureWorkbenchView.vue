<template>
  <div class="pressure-workbench">
    <!-- 顶部工具栏 -->
    <header class="toolbar">
      <div class="left">
        <el-button @click="$router.push('/')">
          <el-icon><ArrowLeft /></el-icon>返回首页
        </el-button>
        <h2>多设备打压控制</h2>
      </div>
      <div class="center">
        <StatCard
          label="设备总数"
          :value="devices.length"
        />
        <StatCard
          label="在线设备"
          :value="onlineCount"
          color="#10b981"
        />
      </div>
      <div class="right">
        <el-button @click="refreshStatus">
          <el-icon><Refresh /></el-icon>刷新状态
        </el-button>
        <el-button
          type="primary"
          @click="showAddDialog = true"
        >
          <el-icon><Plus /></el-icon>添加设备
        </el-button>
      </div>
    </header>
    
    <!-- 设备卡片网格 -->
    <div class="device-grid">
      <div 
        v-for="device in devices" 
        :key="device.id"
        class="device-card"
        :class="{ 'is-connected': device.status === 'connected' }"
      >
        <div class="card-header">
          <h4>{{ device.name }}</h4>
          <div
            class="status-indicator"
            :class="device.status"
          />
        </div>
        
        <div class="pressure-display">
          <div class="current-pressure">
            <span class="value">{{ device.currentPressure?.toFixed(2) || '--' }}</span>
            <span class="unit">{{ device.unit }}</span>
          </div>
          <div class="label">
            当前压力
          </div>
        </div>
        
        <div class="control-section">
          <div class="input-group">
            <label>设定压力</label>
            <el-input-number 
              v-model="device.targetPressure" 
              :precision="2" 
              :step="0.1"
              :min="0"
            />
          </div>
          <div class="input-group">
            <label>单位</label>
            <el-select
              v-model="device.unit"
              size="small"
            >
              <el-option
                label="kPa"
                value="kPa"
              />
              <el-option
                label="MPa"
                value="MPa"
              />
              <el-option
                label="bar"
                value="bar"
              />
              <el-option
                label="psi"
                value="psi"
              />
            </el-select>
          </div>
        </div>
        
        <div class="card-actions">
          <el-button 
            :type="device.status === 'connected' ? 'danger' : 'success'"
            @click="toggleConnection(device)"
          >
            {{ device.status === 'connected' ? '断开' : '连接' }}
          </el-button>
          <el-button 
            type="primary" 
            :disabled="device.status !== 'connected'"
            @click="setPressure(device)"
          >
            设定压力
          </el-button>
        </div>
      </div>
    </div>
    
    <!-- 添加设备对话框 -->
    <el-dialog
      v-model="showAddDialog"
      title="添加打压设备"
      width="400px"
    >
      <el-form
        :model="newDevice"
        label-width="80px"
      >
        <el-form-item label="设备名称">
          <el-input
            v-model="newDevice.name"
            placeholder="请输入设备名称"
          />
        </el-form-item>
        <el-form-item label="IP地址">
          <el-input
            v-model="newDevice.ip"
            placeholder="192.168.1.xxx"
          />
        </el-form-item>
        <el-form-item label="端口">
          <el-input-number
            v-model="newDevice.port"
            :min="1"
            :max="65535"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">
          取消
        </el-button>
        <el-button
          type="primary"
          @click="addDevice"
        >
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ArrowLeft, Refresh, Plus } from '@element-plus/icons-vue'
import StatCard from '@/components/common/StatCard.vue'

interface PressureDevice {
  id: string
  name: string
  ip: string
  port: number
  status: 'connected' | 'disconnected'
  currentPressure?: number
  targetPressure: number
  unit: string
}

// 设备列表
const devices = ref<PressureDevice[]>([
  { id: '1', name: '打压设备-1', ip: '192.168.1.101', port: 502, status: 'connected', currentPressure: 50.25, targetPressure: 100, unit: 'kPa' },
  { id: '2', name: '打压设备-2', ip: '192.168.1.102', port: 502, status: 'disconnected', targetPressure: 200, unit: 'kPa' },
  { id: '3', name: '打压设备-3', ip: '192.168.1.103', port: 502, status: 'connected', currentPressure: 150.00, targetPressure: 300, unit: 'kPa' }
])

// 统计
const onlineCount = computed(() => devices.value.filter(d => d.status === 'connected').length)

// 添加设备对话框
const showAddDialog = ref(false)
const newDevice = ref({
  name: '',
  ip: '',
  port: 502
})

// 方法
const toggleConnection = (device: PressureDevice) => {
  if (device.status === 'connected') {
    device.status = 'disconnected'
    device.currentPressure = undefined
  } else {
    device.status = 'connected'
    device.currentPressure = 0
  }
}

const setPressure = (device: PressureDevice) => {
  console.log(`设置设备 ${device.name} 压力为 ${device.targetPressure} ${device.unit}`)
  if (device.status === 'connected') {
    device.currentPressure = device.targetPressure
  }
}

const refreshStatus = () => {
  console.log('刷新设备状态')
}

const addDevice = () => {
  const id = String(devices.value.length + 1)
  devices.value.push({
    id,
    name: newDevice.value.name || `打压设备-${id}`,
    ip: newDevice.value.ip || '192.168.1.100',
    port: newDevice.value.port,
    status: 'disconnected',
    targetPressure: 0,
    unit: 'kPa'
  })
  showAddDialog.value = false
  newDevice.value = { name: '', ip: '', port: 502 }
}
</script>

<style scoped lang="scss">
.pressure-workbench {
  padding: var(--spacing-lg);
  
  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    padding: var(--spacing-md);
    margin-bottom: var(--spacing-lg);
    
    .left {
      display: flex;
      align-items: center;
      gap: var(--spacing-md);
      
      h2 {
        color: var(--text-primary);
        margin: 0;
        font-size: 20px;
      }
    }
    
    .center {
      display: flex;
      gap: var(--spacing-md);
    }
    
    .right {
      display: flex;
      gap: var(--spacing-sm);
    }
  }
  
  .device-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
    gap: var(--spacing-lg);
    
    .device-card {
      background: var(--bg-secondary);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-lg);
      padding: var(--spacing-lg);
      transition: all 0.3s;
      
      &.is-connected {
        border-color: var(--status-success);
        box-shadow: 0 0 0 1px var(--status-success);
      }
      
      .card-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: var(--spacing-md);
        
        h4 {
          color: var(--text-primary);
          margin: 0;
        }
        
        .status-indicator {
          width: 12px;
          height: 12px;
          border-radius: 50%;
          
          &.connected {
            background: var(--status-success);
            box-shadow: 0 0 8px var(--status-success);
          }
          
          &.disconnected {
            background: var(--text-muted);
          }
        }
      }
      
      .pressure-display {
        text-align: center;
        padding: var(--spacing-lg);
        background: var(--bg-tertiary);
        border-radius: var(--radius-md);
        margin-bottom: var(--spacing-md);
        
        .current-pressure {
          .value {
            font-size: 48px;
            font-weight: bold;
            color: var(--accent-primary);
          }
          
          .unit {
            font-size: 24px;
            color: var(--text-secondary);
            margin-left: var(--spacing-sm);
          }
        }
        
        .label {
          color: var(--text-secondary);
          font-size: 14px;
          margin-top: var(--spacing-xs);
        }
      }
      
      .control-section {
        display: grid;
        grid-template-columns: 1fr 100px;
        gap: var(--spacing-md);
        margin-bottom: var(--spacing-md);
        
        .input-group {
          label {
            display: block;
            color: var(--text-secondary);
            font-size: 12px;
            margin-bottom: var(--spacing-xs);
          }
          
          .el-input-number,
          .el-select {
            width: 100%;
          }
        }
      }
      
      .card-actions {
        display: flex;
        gap: var(--spacing-sm);
        
        .el-button {
          flex: 1;
        }
      }
    }
  }
}
</style>