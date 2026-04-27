<template>
  <PageLayout>
    <!-- 页面头部 -->
    <header class="page-header">
      <div class="header-left">
        <button class="back-btn" @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
        </button>
        <div class="header-title">
          <h1>标定工作台</h1>
          <p>设备标定校准与数据采集</p>
        </div>
      </div>
      <div class="header-actions">
        <span class="state-badge" :class="stateClass">{{ stateLabel }}</span>
      </div>
    </header>

    <!-- 标定工作台内容 -->
    <div class="workbench-content">
      <CalibrationSidebar :collapsed="sidebarCollapsed" @toggle="sidebarCollapsed = !sidebarCollapsed" />
      
      <main class="workbench-main">
        <ProgressIndicator :current-step="calibrationStore.currentStep" />
        <CalibrationParams />
        <CalibrationControl />
        <CalibrationDataView @select-template="dialogsRef?.openTemplateDialog()" />
        <div v-if="dialogsRef?.templateFilename" class="template-result-bar">
          <el-icon><DocumentChecked /></el-icon>
          <span>当前报告模板：{{ dialogsRef.templateFilename }}</span>
        </div>
      </main>
    </div>

    <CalibrationDialogs ref="dialogsRef" />
  </PageLayout>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowLeft, DocumentChecked } from '@element-plus/icons-vue'
import { useCalibrationStore } from '@/stores/calibration'
import { useCalibrationSync } from '@/composables/useCalibrationSync'
import { useConfigPersistence } from '@/composables/useConfigPersistence'
import PageLayout from '@/components/common/PageLayout.vue'
import CalibrationSidebar from '@/components/calibration/CalibrationSidebar.vue'
import CalibrationParams from '@/components/calibration/CalibrationParams.vue'
import CalibrationControl from '@/components/calibration/CalibrationControl.vue'
import CalibrationDataView from '@/components/calibration/CalibrationDataView.vue'
import CalibrationDialogs from '@/components/calibration/CalibrationDialogs.vue'
import ProgressIndicator from '@/components/calibration/ProgressIndicator.vue'

const router = useRouter()
const calibrationStore = useCalibrationStore()
const sidebarCollapsed = ref(false)
const dialogsRef = ref<InstanceType<typeof CalibrationDialogs>>()

useCalibrationSync()
useConfigPersistence()

function goBack(): void {
  router.push('/')
}

const stateLabel = computed(() => {
  const state = calibrationStore.sessionState
  const stateMap: Record<string, string> = {
    idle: '空闲',
    preparing: '准备中',
    running: '运行中',
    paused: '已暂停',
    completed: '已完成',
    error: '错误'
  }
  return stateMap[state] || state
})

const stateClass = computed(() => {
  const state = calibrationStore.sessionState
  const classMap: Record<string, string> = {
    idle: 'state-idle',
    preparing: 'state-preparing',
    running: 'state-running',
    paused: 'state-paused',
    completed: 'state-completed',
    error: 'state-error'
  }
  return classMap[state] || ''
})
</script>

<style scoped lang="scss">

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  flex-shrink: 0;
  padding-bottom: $spacing-4;
  border-bottom: 1px solid $border-color-light;
}

.header-left {
  display: flex;
  align-items: center;
  gap: $spacing-4;
}

.back-btn {
  width: 40px;
  height: 40px;
  background: rgba($neutral-800, 0.6);
  border: 1px solid $border-color;
  border-radius: $radius-md;
  color: $text-secondary;
  cursor: pointer;
  transition: all $transition-fast;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;

  &:hover {
    background: rgba($neutral-700, 0.8);
    color: $text-primary;
    border-color: $border-color-strong;
  }
}

.header-title {
  h1 {
    font-size: 28px;
    font-weight: $font-weight-bold;
    color: $text-primary;
    margin: 0 0 $spacing-1;
    letter-spacing: -0.02em;
  }

  p {
    font-size: $font-size-sm;
    color: $text-secondary;
    margin: 0;
  }
}

.header-actions {
  display: flex;
  gap: $spacing-3;
}

.state-badge {
  padding: $spacing-2 $spacing-4;
  border-radius: $radius-md;
  font-size: $font-size-sm;
  font-weight: $font-weight-medium;
  
  &.state-idle {
    background: rgba($neutral-700, 0.5);
    color: $text-secondary;
  }
  
  &.state-preparing,
  &.state-running {
    background: rgba($primary-500, 0.2);
    color: $primary-400;
  }
  
  &.state-paused {
    background: rgba($warning-500, 0.2);
    color: $warning-400;
  }
  
  &.state-completed {
    background: rgba($success-500, 0.2);
    color: $success-400;
  }
  
  &.state-error {
    background: rgba($danger-500, 0.2);
    color: $danger-400;
  }
}

.workbench-content {
  flex: 1;
  min-height: 0;
  display: flex;
  gap: $spacing-6;
  overflow: hidden;
}

.workbench-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: $spacing-6;
  overflow-y: auto;
  
  &::-webkit-scrollbar {
    width: 8px;
  }
  
  &::-webkit-scrollbar-thumb {
    background: $neutral-700;
    border-radius: 4px;
  }
  
  &::-webkit-scrollbar-track {
    background: transparent;
  }
}

.template-result-bar {
  display: flex;
  align-items: center;
  gap: $spacing-2;
  padding: $spacing-3 $spacing-4;
  background: rgba($success-500, 0.1);
  border: 1px solid rgba($success-500, 0.3);
  border-radius: $radius-md;
  font-size: $font-size-sm;
  flex-shrink: 0;

  .el-icon {
    color: $success-400;
  }

  span {
    color: $text-secondary;
  }
}

// 响应式适配
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: $spacing-4;
  }

  .header-left {
    width: 100%;
  }

  .header-title h1 {
    font-size: $font-size-xl;
  }
  
  .workbench-content {
    flex-direction: column;
  }
}
</style>
