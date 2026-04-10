<template>
  <div class="main-view">
    <div class="calibration-container">
      <!-- 第一行：进度指示器 + 设备面板 -->
      <div class="row row-3">
        <div class="col col-progress">
          <ProgressIndicator :current-step="currentStep" />
        </div>
        <div class="col col-device-1604">
          <Device1604Panel />
        </div>
        <div class="col col-device-press">
          <PressDevicePanel />
        </div>
      </div>
      
      <!-- 第二行：通道选择 + 校准控制 -->
      <div class="row row-2">
        <div class="col col-channels">
          <ChannelMatrix />
        </div>
        <div class="col col-control">
          <CalibrationControlPanel
            :device1604-connected="device1604Connected"
            :press-device-connected="pressDeviceConnected"
            :channels-selected="channelsSelected"
            :has-collected-data="hasCollectedData"
            @start="startCalibration"
            @fit="fitData"
            @end="endCalibration"
          />
        </div>
      </div>
      
      <!-- 第三行：压力点列表 -->
      <div class="row row-full">
        <PressurePointList />
      </div>
      
      <!-- 第四行：数据表格 -->
      <div class="row row-full">
        <CalibrationDataTable :data="calibrationData" :selected-channels="[1, 2, 3, 4, 5]" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import ProgressIndicator from '@/components/calibration/ProgressIndicator.vue'
import Device1604Panel from '@/components/calibration/Device1604Panel.vue'
import PressDevicePanel from '@/components/calibration/PressDevicePanel.vue'
import ChannelMatrix from '@/components/calibration/ChannelMatrix.vue'
import CalibrationControlPanel from '@/components/calibration/CalibrationControlPanel.vue'
import PressurePointList from '@/components/calibration/PressurePointList.vue'
import CalibrationDataTable from '@/components/calibration/CalibrationDataTable.vue'

// 标定数据类型
type CalibrationData = {
  point: number
  targetPressure: number
  channelData: number[]
  status: 'collected' | 'pending'
}

// 当前步骤
const currentStep = ref(0)

// 设备连接状态（应从store获取）
const device1604Connected = ref(false)
const pressDeviceConnected = ref(false)
const channelsSelected = ref(false)
const hasCollectedData = ref(false)

// 模拟数据
const calibrationData = ref<CalibrationData[]>([
  { point: 1, targetPressure: 10, channelData: [10.001, 10.002, 10.000, 10.001, 9.999], status: 'collected' },
  { point: 2, targetPressure: 20, channelData: [20.001, 20.002, 20.000, 20.001, 19.999], status: 'collected' },
  { point: 3, targetPressure: 30, channelData: [30.001, 30.002, 30.000, 30.001, 29.999], status: 'collected' }
])

// 操作
const startCalibration = () => {
  console.log('开始校准')
  currentStep.value = 2
}

const fitData = () => {
  console.log('数据拟合')
  currentStep.value = 4
}

const endCalibration = () => {
  console.log('结束校准')
  currentStep.value = 5
}
</script>

<style scoped lang="scss">
.main-view {
  padding: var(--spacing-lg);
  
  .calibration-container {
    max-width: 1600px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: var(--spacing-lg);
    
    .row {
      display: flex;
      gap: var(--spacing-lg);
      
      &.row-3 {
        .col-progress {
          flex: 5;
        }
        .col-device-1604 {
          flex: 3;
        }
        .col-device-press {
          flex: 4;
        }
      }
      
      &.row-2 {
        .col-channels {
          flex: 7;
        }
        .col-control {
          flex: 5;
        }
      }
      
      &.row-full {
        width: 100%;
      }
      
      .col {
        min-width: 0;
      }
    }
  }
}
</style>
