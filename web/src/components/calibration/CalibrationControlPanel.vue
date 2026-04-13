<template>
  <div class="control-panel">
    <h4 class="title">
      校准控制
    </h4>
    
    <div class="main-action">
      <el-button 
        type="primary" 
        size="large" 
        class="start-btn"
        :disabled="!canStart"
        @click="startCalibration"
      >
        <el-icon><VideoPlay /></el-icon>
        开始校准
      </el-button>
      
      <div class="prerequisites">
        <div 
          v-for="(item, index) in prerequisites" 
          :key="index"
          class="prereq-item"
          :class="{ satisfied: item.satisfied }"
        >
          <el-icon v-if="item.satisfied">
            <CircleCheckFilled />
          </el-icon>
          <el-icon v-else>
            <CircleClose />
          </el-icon>
          <span>{{ item.label }}</span>
        </div>
      </div>
    </div>
    
    <el-divider />
    
    <div class="secondary-actions">
      <el-button 
        type="success" 
        :disabled="!canFit"
        @click="fitData"
      >
        <el-icon><DataAnalysis /></el-icon>
        数据拟合
      </el-button>
      <el-button 
        type="danger" 
        plain
        @click="endCalibration"
      >
        <el-icon><CircleClose /></el-icon>
        结束校准
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { VideoPlay, CircleCheckFilled, CircleClose, DataAnalysis } from '@element-plus/icons-vue'

const props = defineProps<{
  device1604Connected: boolean
  pressDeviceConnected: boolean
  channelsSelected: boolean
  hasCollectedData: boolean
}>()

const prerequisites = computed(() => [
  { label: '1604设备已连接', satisfied: props.device1604Connected },
  { label: '打压设备已连接', satisfied: props.pressDeviceConnected },
  { label: '已选择通道', satisfied: props.channelsSelected }
])

const canStart = computed(() => 
  props.device1604Connected && 
  props.pressDeviceConnected && 
  props.channelsSelected
)

const canFit = computed(() => props.hasCollectedData)

const emit = defineEmits<{
  start: []
  fit: []
  end: []
}>()

const startCalibration = () => emit('start')
const fitData = () => emit('fit')
const endCalibration = () => emit('end')
</script>

<style scoped lang="scss">
.control-panel {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  
  .title {
    color: var(--text-primary);
    margin: 0 0 var(--spacing-md) 0;
  }
  
  .main-action {
    text-align: center;
    
    .start-btn {
      width: 100%;
      height: 50px;
      font-size: 16px;
      margin-bottom: var(--spacing-md);
      
      .el-icon {
        margin-right: var(--spacing-xs);
      }
    }
    
    .prerequisites {
      text-align: left;
      
      .prereq-item {
        display: flex;
        align-items: center;
        gap: var(--spacing-xs);
        padding: var(--spacing-xs) 0;
        color: var(--text-muted);
        font-size: 13px;
        
        .el-icon {
          font-size: 14px;
        }
        
        &.satisfied {
          color: var(--status-success);
        }
      }
    }
  }
  
  :deep(.el-divider) {
    margin: var(--spacing-md) 0;
    border-color: var(--border-color);
  }
  
  .secondary-actions {
    display: flex;
    gap: var(--spacing-sm);
    
    .el-button {
      flex: 1;
    }
  }
}
</style>
