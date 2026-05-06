<template>
  <PageLayout>
    <!-- ═══ 仪表盘头部 — 深色仪器面板 ═══ -->
    <header class="instrument-header">
      <div class="header-nav">
        <button class="back-btn" @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
        </button>
      </div>

      <div class="header-identity">
        <h1 class="header-title">标定工作台</h1>
        <span class="state-chip" :class="stateClass">
          <span class="chip-dot" />
          {{ stateLabel }}
        </span>
      </div>

      <div class="header-telemetry">
        <div class="telem-cell">
          <span class="telem-label">当前压力</span>
          <span class="telem-value mono">{{ displayPressure }}</span>
          <span class="telem-unit">{{ unitLabel }}</span>
        </div>
        <span class="telem-divider" />
        <div class="telem-cell">
          <span class="telem-label">稳定性</span>
          <span class="telem-indicator" :class="stabilityStatus?.isStable ? 'on' : 'off'">
            <span class="telem-dot" />
            {{ stabilityStatus?.isStable ? '已稳定' : '稳定中' }}
          </span>
        </div>
        <span class="telem-divider" />
        <div class="telem-cell">
          <span class="telem-label">稳定计时</span>
          <span class="telem-value mono">{{ stableSeconds }}<small>s</small></span>
        </div>
        <span class="telem-divider" />
        <div class="telem-cell">
          <span class="telem-label">偏差</span>
          <span class="telem-value mono" :class="deviationClass">{{ deviationDisplay }}</span>
        </div>
      </div>
    </header>

    <!-- ═══ 工作台主体 ═══ -->
    <div class="workbench">
      <!-- 左侧：设备面板 -->
      <CalibrationSidebar :collapsed="sidebarCollapsed" @toggle="sidebarCollapsed = !sidebarCollapsed" />

      <!-- 中央：主工作区 -->
      <main class="workbench-main">
        <!-- 报警横幅 -->
        <div v-if="alarmEvent" class="alarm-banner">
          <span class="alarm-dot" />
          <span>通道 {{ alarmEvent.overLimitChannels?.join(', ') }} 超限报警</span>
        </div>

        <div class="scroll-container">
          <ProgressIndicator :current-step="calibrationStore.currentStep" />

          <section class="card-block card-block-control">
            <div class="card-accent" />
            <CalibrationParams />
            <div class="control-divider" />
            <CalibrationControl />
          </section>

          <div class="section-gap" />

          <section class="card-block card-block-data">
            <div class="card-accent" />
            <CalibrationDataView @select-template="dialogsRef?.openTemplateDialog()" />
            <div v-if="dialogsRef?.templateFilename" class="template-bar">
              <el-icon><DocumentChecked /></el-icon>
              <span>当前报告模板：{{ dialogsRef.templateFilename }}</span>
            </div>
          </section>
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

const { stabilityStatus, alarmEvent } = useCalibrationSync()
provide(stabilityStatusKey, stabilityStatus)
useConfigPersistence()

function goBack(): void {
  router.push('/')
}

/* ── 会话状态 ── */
const stateLabel = computed(() => {
  const m: Record<string, string> = {
    idle: '空闲', preparing: '准备中', running: '运行中',
    paused: '已暂停', completed: '已完成', error: '错误',
  }
  return m[calibrationStore.sessionState] || calibrationStore.sessionState
})

const stateClass = computed(() => {
  const m: Record<string, string> = {
    idle: 'chip-idle', preparing: 'chip-preparing', running: 'chip-running',
    paused: 'chip-paused', completed: 'chip-completed', error: 'chip-error',
  }
  return m[calibrationStore.sessionState] || ''
})

/* ── 压力数据 ── */
const displayPressure = computed(() => {
  const v = calibrationStore.currentPressure
  if (v === null || v === undefined) return '—'
  const n = typeof v === 'number' ? v : Number(v)
  return isNaN(n) ? String(v) : n.toFixed(3)
})

const unitLabel = computed(() => calibrationStore.measureUnit || 'MPa')

/* ── 稳定性 ── */
const stableSeconds = computed(() => {
  if (!stabilityStatus.value) return '0.0'
  return (stabilityStatus.value.stableDurationMs / 1000).toFixed(1)
})

const deviationDisplay = computed(() => {
  if (!stabilityStatus.value) return '—'
  return stabilityStatus.value.deviation.toFixed(4)
})

const deviationClass = computed(() => {
  if (!stabilityStatus.value) return ''
  const d = Math.abs(stabilityStatus.value.deviation)
  if (d < 0.01) return 'dev-good'
  if (d < 0.1) return 'dev-warn'
  return 'dev-bad'
})
</script>

<style scoped lang="scss">
/* ── 设计令牌 ── */
$font-sans: 'Inter', 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
$font-mono: 'JetBrains Mono', 'Fira Code', 'SF Mono', Consolas, monospace;
$mint: #10b981;
$mint-light: #34d399;
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
$slate-900: #111827;
$green: #22c55e;
$red: #ef4444;
$amber: #f59e0b;

/* ════════════════════════════════════════
   仪表盘头部
   ════════════════════════════════════════ */
.instrument-header {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-shrink: 0;
  height: 56px;
  padding: 0 24px;
  background: linear-gradient(135deg, $slate-800 0%, $slate-900 100%);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  font-family: $font-sans;
}

.header-nav {
  display: flex;
  align-items: center;
}

.back-btn {
  width: 30px; height: 30px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 8px;
  color: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  transition: all 0.2s ease;

  &:hover {
    background: rgba(255, 255, 255, 0.12);
    color: #fff;
    border-color: rgba(255, 255, 255, 0.25);
  }
}

.header-identity {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.header-title {
  font-size: 16px;
  font-weight: 700;
  color: #fff;
  margin: 0;
  letter-spacing: 0.02em;
  font-family: $font-sans;
}

.header-telemetry {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-left: auto;
}

.telem-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.telem-label {
  font-size: 11px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.35);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  white-space: nowrap;
}

.telem-value {
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  letter-spacing: 0.01em;

  &.mono { font-family: $font-mono; }
  small { font-size: 10px; color: rgba(255, 255, 255, 0.4); margin-left: 1px; }
}

.telem-unit {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.4);
  font-weight: 500;
  font-family: $font-sans;
}

.telem-divider {
  width: 1px;
  height: 24px;
  background: rgba(255, 255, 255, 0.08);
}

.telem-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  padding: 2px 10px;
  border-radius: 4px;

  &.on {
    color: $mint-light;
    background: rgba(16, 185, 129, 0.12);
  }
  &.off {
    color: $amber;
    background: rgba(245, 158, 11, 0.12);
  }
}

.telem-dot {
  width: 6px; height: 6px;
  border-radius: 50%;

  .on & {
    background: $mint-light;
    box-shadow: 0 0 6px rgba(16, 185, 129, 0.6);
    animation: pulse-dot 2s ease-in-out infinite;
  }
  .off & {
    background: $amber;
    box-shadow: 0 0 6px rgba(245, 158, 11, 0.6);
    animation: pulse-dot 1.2s ease-in-out infinite;
  }
}

@keyframes pulse-dot {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(0.85); }
}

/* ── 状态芯片 ── */
.state-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 2px 10px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.03em;

  .chip-dot {
    width: 5px; height: 5px;
    border-radius: 50%;
  }

  &.chip-idle {
    background: rgba(156, 163, 175, 0.15); color: $slate-400;
    .chip-dot { background: $slate-400; }
  }
  &.chip-preparing, &.chip-running {
    background: rgba(59, 130, 246, 0.15); color: #60a5fa;
    .chip-dot { background: #60a5fa; box-shadow: 0 0 5px rgba(96, 165, 250, 0.5); }
  }
  &.chip-paused {
    background: rgba(245, 158, 11, 0.15); color: $amber;
    .chip-dot { background: $amber; }
  }
  &.chip-completed {
    background: rgba(16, 185, 129, 0.15); color: $mint-light;
    .chip-dot { background: $mint-light; }
  }
  &.chip-error {
    background: rgba(239, 68, 68, 0.15); color: #f87171;
    .chip-dot { background: #f87171; }
  }
}

/* ════════════════════════════════════════
   工作台主体
   ════════════════════════════════════════ */
.workbench {
  flex: 1;
  min-height: 0;
  display: flex;
  gap: 16px;
  overflow: hidden;
  position: relative;
  padding: 4px 24px 24px;

  // 坐标纸背景
  &::before {
    content: '';
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(rgba(16, 185, 129, 0.04) 1px, transparent 1px),
      linear-gradient(90deg, rgba(16, 185, 129, 0.04) 1px, transparent 1px);
    background-size: 24px 24px;
    pointer-events: none;
    z-index: 0;
  }
}

/* ── 中央工作区 ── */
.workbench-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  position: relative;
  z-index: 1;
}

.scroll-container {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 4px 4px 4px 0;

  &::-webkit-scrollbar { width: 4px; }
  &::-webkit-scrollbar-thumb {
    background: $slate-300;
    border-radius: 2px;
  }
  &::-webkit-scrollbar-track { background: transparent; }
}

.section-gap {
  flex-shrink: 0;
  height: 8px;
}

/* ── 卡片区块 ── */
.card-block {
  background: #ffffff;
  border-radius: 10px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04);
  position: relative;
  overflow: hidden;
  animation: card-enter 0.35s ease both;
}

.card-block:nth-child(1) { animation-delay: 0ms; }
.card-block:nth-child(2) { animation-delay: 60ms; }

.card-accent {
  position: absolute;
  top: 0; left: 0; right: 0;
  height: 3px;
  background: linear-gradient(90deg, $mint, $mint-light, rgba(16, 185, 129, 0.3));
}

@keyframes card-enter {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

.card-block-control {
  padding: 0 16px;
}

.control-divider {
  height: 1px;
  background: $slate-200;
  margin: 0 -16px;
}

.template-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px 14px;
  font-size: 13px;
  font-weight: 500;
  color: $slate-500;
  font-family: $font-sans;

  .el-icon { color: $mint; font-size: 16px; }
}

/* ── 偏差颜色 ── */
.dev-good { color: $mint-light !important; }
.dev-warn { color: $amber !important; }
.dev-bad { color: $red !important; }

/* ── 报警横幅 ── */
.alarm-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: rgba(239, 68, 68, 0.06);
  border-bottom: 1px solid rgba(239, 68, 68, 0.15);
  font-size: 13px;
  font-weight: 600;
  color: $red;
  flex-shrink: 0;
  animation: card-enter 0.3s ease both;
}

.alarm-dot {
  width: 7px; height: 7px;
  border-radius: 50%;
  background: $red;
  animation: pulse-dot 1s ease-in-out infinite;
}

/* ════════════════════════════════════════
   响应式
   ════════════════════════════════════════ */
@media (max-width: 1024px) {
  .instrument-header {
    gap: 12px;
    padding: 0 16px;
  }
  .header-telemetry { gap: 10px; }
  .telem-label { display: none; }
}

@media (max-width: 768px) {
  .instrument-header {
    height: auto;
    flex-wrap: wrap;
    padding: 10px 16px;
    gap: 8px;
  }
  .header-telemetry { flex-wrap: wrap; margin-left: 0; width: 100%; }
  .telem-divider { display: none; }
  .workbench { flex-direction: column; }
  .workbench-main { min-height: 400px; }
}
</style>
