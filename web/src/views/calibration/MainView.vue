<template>
  <div class="main-view">
    <div class="calibration-container">
      <!-- 第一行：进度指示器 + 设备面板 -->
      <div class="row row-3">
        <div class="col col-progress">
          <ProgressIndicator :current-step="calibrationStore.currentStep" />
        </div>
        <div class="col col-device-1604">
          <Device1604Panel 
            @connect="calibrationStore.connectDevice1604"
            @disconnect="calibrationStore.disconnectDevice1604"
          />
        </div>
        <div class="col col-device-press">
          <PressDevicePanel 
            @connect="calibrationStore.connectPressDevice"
            @disconnect="calibrationStore.disconnectPressDevice"
          />
        </div>
      </div>
      
      <!-- 第二行：通道选择 + 校准控制 -->
      <div class="row row-2">
        <div class="col col-channels">
          <ChannelMatrix 
            :selected-channels="calibrationStore.selectedChannels"
            @update:selected-channels="calibrationStore.setSelectedChannels"
          />
        </div>
        <div class="col col-control">
          <CalibrationControlPanel
            :device1604-connected="calibrationStore.device1604Connected"
            :press-device-connected="calibrationStore.pressDeviceConnected"
            :channels-selected="calibrationStore.channelsSelected"
            :has-collected-data="calibrationStore.hasCollectedData"
            @start="calibrationStore.startCalibration"
            @fit="calibrationStore.fitData"
            @end="calibrationStore.endCalibration"
          />
        </div>
      </div>
      
      <!-- 第三行：压力点列表 -->
      <div class="row row-full">
        <PressurePointList 
          :points="calibrationStore.pressurePoints"
          :params="calibrationStore.calibrationParams"
          @generate="calibrationStore.generatePressurePoints"
          @pressurize="calibrationStore.pressurize"
          @confirm="calibrationStore.confirmPressure"
          @collect="calibrationStore.collectData"
          @remove="calibrationStore.removePressurePoint"
        />
      </div>
      
      <!-- 第四行：数据表格 -->
      <div class="row row-full">
        <CalibrationDataTable 
          :points="calibrationStore.pressurePoints"
          :selected-channels="calibrationStore.selectedChannels"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import ProgressIndicator from '@/components/calibration/ProgressIndicator.vue'
import Device1604Panel from '@/components/calibration/Device1604Panel.vue'
import PressDevicePanel from '@/components/calibration/PressDevicePanel.vue'
import ChannelMatrix from '@/components/calibration/ChannelMatrix.vue'
import CalibrationControlPanel from '@/components/calibration/CalibrationControlPanel.vue'
import PressurePointList from '@/components/calibration/PressurePointList.vue'
import CalibrationDataTable from '@/components/calibration/CalibrationDataTable.vue'
import { useCalibrationStore } from '@/stores/calibration'
import { useMeasurementDeviceStore } from '@/stores/measurement/deviceStore'

// Store
const calibrationStore = useCalibrationStore()
const deviceStore = useMeasurementDeviceStore()

// 页面加载时获取设备列表
onMounted(() => {
  deviceStore.loadDevices()
})
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
