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
          <span class="state-badge" :class="stateClass">{{ stateLabel }}</span>
        </div>
      </div>
      <div class="header-right">
        <div class="status-info">
          <span class="info-label">稳定:</span>
          <span
            class="info-value"
            :class="calibrationStore.isStable ? 'stable' : 'unstable'"
          >
            {{ calibrationStore.isStable ? '是' : '否' }}
          </span>
        </div>
        <div class="status-info">
          <span class="info-label">实时时间:</span>
          <span class="time-badge">
            {{ stabilityStatus ? (stabilityStatus.stableDurationMs / 1000).toFixed(1) : '0.0' }}<small>s</small>
          </span>
        </div>
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
import { ref, computed, provide } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowLeft, DocumentChecked } from '@element-plus/icons-vue'
import { useCalibrationStore } from '@/stores/calibration'
import { useCalibrationSync, stabilityStatusKey } from '@/composables/useCalibrationSync'
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

const { stabilityStatus } = useCalibrationSync()
provide(stabilityStatusKey, stabilityStatus)
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
/* ── 设计系统令牌 ── */
$font-sans: 'Inter', 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
$font-mono: 'JetBrains Mono', 'Fira Code', 'SF Mono', Consolas, monospace;
$mint: #10b981;
$mint-dark: #059669;
$slate-50: #f9fafb;
$slate-100: #f3f4f6;
$slate-200: #e5e7eb;
$slate-300: #d1d5db;
$slate-400: #9ca3af;
$slate-500: #6b7280;
$slate-600: #4b5563;
$slate-700: #374151;
$slate-800: #1f2937;
$green: #22c55e;
$red: #ef4444;
$blue: #3b82f6;
$amber: #f59e0b;

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
  height: 48px;
  padding: 0 24px;
  background: #ffffff;
  border-bottom: 1px solid $slate-200;
  font-family: $font-sans;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-btn {
  width: 28px;
  height: 28px;
  background: transparent;
  border: 1px solid $slate-200;
  border-radius: 8px;
  color: $slate-400;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;

  &:hover {
    background: $slate-50;
    color: $slate-600;
    border-color: $slate-300;
  }
}

.header-title {
  display: flex;
  align-items: center;
  gap: 10px;

  h1 {
    font-size: 18px;
    font-weight: 700;
    color: $slate-800;
    margin: 0;
    letter-spacing: -0.01em;
    font-family: $font-sans;
  }
}

/* 状态徽章：按 Tags 规范 */
.state-badge {
  padding: 3px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  line-height: 1.5;
  letter-spacing: 0.02em;
}

.state-idle {
  background: rgba(107, 114, 128, 0.12);
  border: 1px solid rgba(107, 114, 128, 0.25);
  color: $slate-500;
}
.state-preparing, .state-running {
  background: rgba(59, 130, 246, 0.12);
  border: 1px solid rgba(59, 130, 246, 0.25);
  color: #2563eb;
}
.state-paused {
  background: rgba(245, 158, 11, 0.12);
  border: 1px solid rgba(245, 158, 11, 0.25);
  color: #d97706;
}
.state-completed {
  background: rgba(16, 185, 129, 0.12);
  border: 1px solid rgba(16, 185, 129, 0.25);
  color: #059669;
}
.state-error {
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.25);
  color: #dc2626;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 24px;
}

.status-info {
  display: flex;
  align-items: center;
  gap: 6px;
}

.info-label {
  color: $slate-400;
  font-size: 12px;
  font-weight: 500;
  letter-spacing: 0.05em;
}

.info-value {
  font-size: 13px;
  font-weight: 600;

  &.stable { color: $green; }
  &.unstable { color: $red; }
}

.time-badge {
  font-family: $font-mono;
  background: $slate-100;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 13px;
  font-weight: 500;
  color: $slate-700;

  small {
    font-size: 10px;
    margin-left: 2px;
    color: $slate-400;
  }
}

.workbench-content {
  flex: 1;
  min-height: 0;
  display: flex;
  gap: 16px;
  overflow: hidden;
}

.workbench-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow-y: auto;
  padding-right: 4px;

  &::-webkit-scrollbar {
    width: 6px;
  }

  &::-webkit-scrollbar-thumb {
    background: $slate-300;
    border-radius: 3px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }
}

.template-result-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: rgba(16, 185, 129, 0.08);
  border: 1px solid rgba(16, 185, 129, 0.2);
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  color: $slate-600;
  flex-shrink: 0;
  font-family: $font-sans;

  .el-icon {
    color: $mint;
  }
}

// 响应式适配
@media (max-width: 768px) {
  .page-header { padding: 0 16px; }
  .header-right { display: none; }
  .workbench-content { flex-direction: column; }
}
</style>
