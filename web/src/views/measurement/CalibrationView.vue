<template>
  <div class="calibration-view">
    <!-- 可折叠侧边栏 -->
    <aside class="sidebar" :class="{ collapsed: sidebarCollapsed }">
      <div class="sidebar-toggle" @click="toggleSidebar">
        <el-icon><ArrowLeft v-if="!sidebarCollapsed" /><ArrowRight v-else /></el-icon>
      </div>
      <div v-show="!sidebarCollapsed" class="sidebar-content">
        <DevicePanel title="打压设备">
          <PressureDeviceCard
            v-for="device in deviceStore.pressureDevices"
            :key="device.id"
            :device="device"
            @connect="deviceStore.connectPressureDevice"
            @disconnect="deviceStore.disconnectPressureDevice"
          />
          <el-button type="primary" plain size="small" class="add-btn">
            <el-icon><Plus /></el-icon>添加设备
          </el-button>
        </DevicePanel>
        
        <DevicePanel title="计量设备">
          <MeasureDeviceCard
            v-for="device in deviceStore.measureDevices"
            :key="device.id"
            :device="device"
            @connect="deviceStore.connectMeasureDevice"
            @disconnect="deviceStore.disconnectMeasureDevice"
          />
          <el-button type="primary" plain size="small" class="add-btn">
            <el-icon><Plus /></el-icon>添加设备
          </el-button>
        </DevicePanel>
      </div>
    </aside>
    
    <!-- 主工作区 -->
    <main class="workbench">
      <!-- 第一行控制条 -->
      <div class="control-bar">
        <div class="control-group">
          <label>最小值</label>
          <el-input-number v-model="params.minValue" :precision="2" :step="0.1" />
        </div>
        <div class="control-group">
          <label>最大值</label>
          <el-input-number v-model="params.maxValue" :precision="2" :step="0.1" />
        </div>
        <div class="control-group">
          <label>点数</label>
          <el-input-number v-model="params.points" :min="2" :max="50" />
        </div>
        <div class="control-group">
          <label>精度</label>
          <el-input-number v-model="params.precision" :min="0" :max="4" />
        </div>
        <div class="control-group">
          <label>平均数</label>
          <el-input-number v-model="params.averageCount" :min="1" :max="100" />
        </div>
        <div class="control-group">
          <label>稳定时间</label>
          <el-select v-model="params.stableTime">
            <el-option label="1秒" :value="1" />
            <el-option label="3秒" :value="3" />
            <el-option label="5秒" :value="5" />
            <el-option label="10秒" :value="10" />
          </el-select>
        </div>
        <div class="control-group">
          <label>精度Level</label>
          <el-select v-model="params.precisionLevel">
            <el-option label="0.01%" value="0.01" />
            <el-option label="0.05%" value="0.05" />
            <el-option label="0.1%" value="0.1" />
            <el-option label="0.2%" value="0.2" />
          </el-select>
        </div>
        <el-button type="primary" class="generate-btn">
          生成压力表
        </el-button>
      </div>
      
      <!-- 第二行控制条 -->
      <div class="control-bar secondary">
        <div class="mode-switches">
          <div class="switch-group">
            <span>控制模式</span>
            <el-radio-group v-model="controlMode">
              <el-radio-button label="auto">自动</el-radio-button>
              <el-radio-button label="manual">手动</el-radio-button>
            </el-radio-group>
          </div>
          <div class="switch-group">
            <span>打压模式</span>
            <el-radio-group v-model="pressureMode">
              <el-radio-button label="single">单程</el-radio-button>
              <el-radio-button label="round">回程</el-radio-button>
            </el-radio-group>
          </div>
        </div>
        
        <div class="progress-section">
          <div class="progress-info">
            <span>进度: {{ currentPoint }}/{{ totalPoints }}</span>
            <el-progress :percentage="progressPercent" :stroke-width="8" />
          </div>
          <div class="stable-status">
            <span>稳定状态: {{ isStable ? '已稳定' : '稳定中' }}</span>
            <span v-if="!isStable" class="countdown">剩余: {{ stableCountdown }}s</span>
          </div>
        </div>
        
        <div class="action-buttons">
          <el-button type="success" @click="startCollection">开始采集</el-button>
          <el-button @click="pauseCollection">暂停</el-button>
          <el-button type="danger" @click="stopCollection">停止</el-button>
          <el-button @click="resetCollection">重置</el-button>
          <el-button type="primary" plain @click="exportReport">导出报告</el-button>
        </div>
      </div>
      
      <!-- 数据表格 -->
      <div class="data-table-wrapper">
        <el-table :data="tableData" border stripe class="data-table">
          <el-table-column prop="index" label="序号" width="60" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)">{{ getStatusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="targetValue" label="目标值" width="120" />
          <el-table-column 
            v-for="ch in 16" 
            :key="ch"
            :label="`CH${ch}`" 
            width="80"
          >
            <template #default="{ row }">
              <span :class="getChannelClass(row, ch - 1)">{{ row.channelValues[ch - 1]?.toFixed(2) || '--' }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="timestamp" label="时间" width="160" />
        </el-table>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ArrowLeft, ArrowRight, Plus } from '@element-plus/icons-vue'
import { useMeasurementDeviceStore } from '@/stores/measurement/deviceStore'
import DevicePanel from '@/components/measurement/DevicePanel.vue'
import PressureDeviceCard from '@/components/measurement/PressureDeviceCard.vue'
import MeasureDeviceCard from '@/components/measurement/MeasureDeviceCard.vue'

// 侧边栏状态
const sidebarCollapsed = ref(false)
const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

// Store
const deviceStore = useMeasurementDeviceStore()

// 参数
const params = ref({
  minValue: 0,
  maxValue: 100,
  points: 10,
  precision: 2,
  averageCount: 5,
  stableTime: 3,
  precisionLevel: '0.05'
})

// 模式
const controlMode = ref('auto')
const pressureMode = ref('single')

// 进度
const currentPoint = ref(0)
const totalPoints = ref(10)
const progressPercent = computed(() => (currentPoint.value / totalPoints.value) * 100)

// 稳定状态
const isStable = ref(false)
const stableCountdown = ref(0)

// 表格数据
const tableData = ref([
  { index: 1, status: 'completed', targetValue: 10.00, channelValues: [10.01, 10.02, 10.00, 10.01, 9.99, 10.00, 10.01, 10.02, 10.00, 10.01, 9.99, 10.00, 10.01, 10.02, 10.00, 10.01], timestamp: '2024-01-15 10:30:00' },
  { index: 2, status: 'collecting', targetValue: 20.00, channelValues: [20.01, 20.02, 20.00, 20.01, 19.99, 20.00, 20.01, 20.02, null, null, null, null, null, null, null, null], timestamp: '2024-01-15 10:31:00' }
])

// 状态处理
const getStatusType = (status: string) => {
  const map: Record<string, string> = {
    pending: 'info',
    pressurizing: 'warning',
    stabilizing: '',
    collecting: 'primary',
    completed: 'success'
  }
  return map[status] || 'info'
}

const getStatusText = (status: string) => {
  const map: Record<string, string> = {
    pending: '待执行',
    pressurizing: '打压中',
    stabilizing: '稳定中',
    collecting: '采集中',
    completed: '已完成'
  }
  return map[status] || status
}

interface TableRow {
  index: number
  status: string
  targetValue: number
  channelValues: (number | null)[]
  timestamp: string
}

const getChannelClass = (row: TableRow, index: number) => {
  const value = row.channelValues[index]
  if (!value) return ''
  const diff = Math.abs(value - row.targetValue)
  if (diff < 0.1) return 'channel-good'
  if (diff < 0.5) return 'channel-warning'
  return 'channel-error'
}

// 操作
const startCollection = () => console.log('开始采集')
const pauseCollection = () => console.log('暂停采集')
const stopCollection = () => console.log('停止采集')
const resetCollection = () => console.log('重置')
const exportReport = () => console.log('导出报告')
</script>

<style scoped lang="scss">
.calibration-view {
  display: flex;
  height: calc(100vh - 60px);
  background: var(--bg-primary);
  
  .sidebar {
    width: 300px;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border-color);
    position: relative;
    transition: width 0.3s;
    
    &.collapsed {
      width: 40px;
    }
    
    .sidebar-toggle {
      position: absolute;
      right: -20px;
      top: 50%;
      transform: translateY(-50%);
      width: 20px;
      height: 60px;
      background: var(--bg-tertiary);
      border-radius: 0 4px 4px 0;
      display: flex;
      align-items: center;
      justify-content: center;
      cursor: pointer;
      z-index: 10;
      
      .el-icon {
        color: var(--text-secondary);
        font-size: 12px;
      }
    }
    
    .sidebar-content {
      padding: var(--spacing-md);
      overflow-y: auto;
      height: 100%;
      
      .add-btn {
        width: 100%;
        margin-top: var(--spacing-sm);
      }
    }
  }
  
  .workbench {
    flex: 1;
    padding: var(--spacing-md);
    overflow: auto;
    
    .control-bar {
      background: var(--bg-secondary);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-md);
      padding: var(--spacing-md);
      margin-bottom: var(--spacing-md);
      display: flex;
      align-items: center;
      gap: var(--spacing-md);
      flex-wrap: wrap;
      
      .control-group {
        display: flex;
        flex-direction: column;
        gap: var(--spacing-xs);
        
        label {
          color: var(--text-secondary);
          font-size: 12px;
        }
        
        .el-input-number,
        .el-select {
          width: 100px;
        }
      }
      
      .generate-btn {
        margin-left: auto;
      }
      
      &.secondary {
        justify-content: space-between;
        
        .mode-switches {
          display: flex;
          gap: var(--spacing-lg);
          
          .switch-group {
            display: flex;
            flex-direction: column;
            gap: var(--spacing-xs);
            
            span {
              color: var(--text-secondary);
              font-size: 12px;
            }
          }
        }
        
        .progress-section {
          display: flex;
          flex-direction: column;
          gap: var(--spacing-xs);
          min-width: 200px;
          
          .progress-info {
            display: flex;
            align-items: center;
            gap: var(--spacing-sm);
            color: var(--text-primary);
            font-size: 14px;
            
            .el-progress {
              flex: 1;
            }
          }
          
          .stable-status {
            display: flex;
            gap: var(--spacing-md);
            color: var(--text-secondary);
            font-size: 12px;
            
            .countdown {
              color: var(--accent-primary);
            }
          }
        }
        
        .action-buttons {
          display: flex;
          gap: var(--spacing-sm);
        }
      }
    }
    
    .data-table-wrapper {
      background: var(--bg-secondary);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-md);
      padding: var(--spacing-md);
      
      .data-table {
        width: 100%;
        
        :deep(th) {
          background: var(--bg-tertiary);
          color: var(--text-secondary);
        }
        
        :deep(td) {
          color: var(--text-primary);
        }
        
        .channel-good {
          color: var(--status-success);
        }
        
        .channel-warning {
          color: var(--status-warning);
        }
        
        .channel-error {
          color: var(--status-error);
        }
      }
    }
  }
}
</style>