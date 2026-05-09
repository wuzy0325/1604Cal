<template>
  <section class="workbench-section data-section">
    <div
      class="table-panel panel-points"
      data-testid="pressure-point-operation-table"
    >
      <div class="table-header">
        <div class="table-title">
          <el-icon><Operation /></el-icon>
          <h3>压力点设置</h3>
          <span
            v-if="calibrationStore.pressurePoints.length > 0"
            class="record-count"
          >
            {{ calibrationStore.pressurePoints.length }} 个测点
          </span>
        </div>
      </div>

      <div class="table-body point-table-body">
        <el-table
          :data="tableData"
          border
          stripe
          class="point-operation-table"
        >
          <el-table-column
            prop="index"
            label="序号"
            width="55"
          />
          <el-table-column
            label="状态"
            width="85"
          >
            <template #default="{ row }">
              <el-tag
                :type="getPointStatusType(row.status)"
                size="small"
              >
                {{ getPointStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column
            label="目标压力"
            width="110"
          >
            <template #default="{ row }">
              {{ row.targetValue.toFixed(2) }}
            </template>
          </el-table-column>
          <el-table-column
            label="确认"
            width="95"
          >
            <template #default="{ row }">
              <el-button
                class="row-btn"
                size="small"
                :type="getConfirmButtonType(row.status)"
                :disabled="!canConfirm(row.status)"
                :data-testid="`confirm-btn-point-${row.index}`"
                @click="calibrationStore.confirmPressure(row.id)"
              >
                确认
              </el-button>
            </template>
          </el-table-column>
          <el-table-column
            label="采集"
            width="95"
          >
            <template #default="{ row }">
              <el-button
                class="row-btn"
                size="small"
                type="info"
                :disabled="!canCollect(row.status)"
                :data-testid="`collect-btn-point-${row.index}`"
                @click="calibrationStore.collectData(row.id)"
              >
                采集
              </el-button>
            </template>
          </el-table-column>
          <el-table-column
            label="操作"
            min-width="120"
          >
            <template #default="{ row }">
              <div class="row-actions">
              <el-button
                  v-if="row.status === 'pending' && !manualModeWithoutPressDevice"
                  type="primary"
                  size="small"
                  :disabled="!canPressurize(row.status)"
                  :data-testid="`pressurize-btn-point-${row.index}`"
                  @click="calibrationStore.pressurize(row.id)"
                >
                  打压
                </el-button>
                <span
                  v-else-if="row.status === 'pending' && manualModeWithoutPressDevice"
                  class="idle-text"
                >{{ manualPendingActionText }}</span>
                <span
                  v-else-if="row.status === 'collecting'"
                  class="collecting-text"
                >采集中...</span>
                <span
                  v-else
                  class="idle-text"
                >--</span>
              </div>
            </template>
          </el-table-column>
        </el-table>

        <div
          v-if="calibrationStore.pressurePoints.length === 0"
          class="empty-state"
        >
          <el-icon class="empty-icon">
            <SetUp />
          </el-icon>
          <p>配置标定参数后开始标定流程</p>
        </div>
      </div>
    </div>

    <div
      class="table-panel panel-results"
      data-testid="collected-data-table"
    >
      <div class="table-header">
        <div class="table-title">
          <el-icon><DataLine /></el-icon>
          <h3>采集数据</h3>
          <span
            v-if="tableData.length > 0"
            class="record-count"
          >
            {{ collectedCount }}/{{ tableData.length }} 已采集
          </span>
        </div>
        <div class="table-actions">
          <el-button
            size="small"
            @click="$emit('selectTemplate')"
          >
            <el-icon><Document /></el-icon>
            报告模板
          </el-button>
          <el-button
            size="small"
            type="primary"
            :disabled="calibrationStore.pressurePoints.length === 0"
            @click="$emit('exportReport')"
          >
            <el-icon><Download /></el-icon>
            导出报告
          </el-button>
          <el-button
            size="small"
            :disabled="calibrationStore.pressurePoints.length === 0"
            @click="exportCSV"
          >
            <el-icon><DataAnalysis /></el-icon>
            导出CSV
          </el-button>
        </div>
      </div>

      <div class="table-body data-table-body">
        <el-table
          :data="tableData"
          border
          stripe
          class="data-table"
        >
          <el-table-column
            prop="index"
            label="压力点"
            width="70"
          />
          <el-table-column
            label="目标压力"
            width="100"
          >
            <template #default="{ row }">
              {{ row.targetValue.toFixed(2) }}
            </template>
          </el-table-column>
          <el-table-column
            label="实际压力"
            width="100"
          >
            <template #default="{ row }">
              {{ row.actualPressure?.toFixed(2) || '--' }}
            </template>
          </el-table-column>
          <el-table-column
            v-for="ch in calibrationStore.selectedChannels"
            :key="ch"
            :label="`CH${ch}`"
            width="75"
          >
            <template #default="{ row }">
              <span :class="getChannelClass(row, ch - 1)">
                {{ row.channelValues[ch - 1]?.toFixed(precision) || '--' }}
              </span>
            </template>
          </el-table-column>
          <el-table-column
            label="状态"
            width="85"
            fixed="right"
          >
            <template #default="{ row }">
              <el-tag
                :type="getPointStatusType(row.status)"
                size="small"
              >
                {{ getPointStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>

        <div
          v-if="calibrationStore.pressurePoints.length === 0"
          class="empty-state"
        >
          <el-icon class="empty-icon">
            <SetUp />
          </el-icon>
          <p>暂无采集数据</p>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  Operation,
  Document,
  DataLine,
  Download,
  DataAnalysis,
  SetUp
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useCalibrationStore } from '@/stores/calibration'
import type { SessionState } from '@/types/calibration'

defineEmits<{
  selectTemplate: []
  exportReport: []
}>()

const calibrationStore = useCalibrationStore()

const precision = computed(() => calibrationStore.calibrationParams.precision || 2)

// 测点状态
const getPointStatusType = (status: string) => {
  const map: Record<string, string> = {
    pending: 'info',
    pressurizing: 'warning',
    stabilizing: '',
    collecting: 'primary',
    completed: 'success',
    error: 'danger'
  }
  return map[status] || 'info'
}

const getPointStatusText = (status: string) => {
  if (manualModeWithoutPressDevice.value) {
    if (status === 'pending') return '待确认'
    if (status === 'stabilizing') return '待采集'
  }

  const map: Record<string, string> = {
    pending: '待执行',
    pressurizing: '打压中',
    stabilizing: '稳定中',
    collecting: '采集中',
    completed: '已完成',
    error: '错误'
  }
  return map[status] || status
}

// 表格数据
interface TableRow {
  id: string
  index: number
  status: string
  targetValue: number
  channelValues: (number | undefined)[]
  actualPressure?: number
}

const tableData = computed<TableRow[]>(() =>
  calibrationStore.pressurePoints.map(point => ({
    id: point.id,
    index: point.index,
    status: point.status,
    targetValue: point.targetPressure,
    channelValues: point.collectedData || [],
    actualPressure: point.actualPressure
  }))
)

const collectedCount = computed(() =>
  tableData.value.filter(r => r.status === 'completed').length
)

const operableSessionStates: SessionState[] = [
  'ready',
  'pressurizing',
  'stabilizing',
  'collecting',
  'point_done',
  'await_manual_collect',
  'await_alarm_resolution',
  'recovering'
]

const canOperatePointActions = computed(() =>
  operableSessionStates.includes(calibrationStore.sessionState)
)

const manualModeWithoutPressDevice = computed(() =>
  calibrationStore.controlMode === 'manual' && !calibrationStore.pressDeviceConnected
)

const manualPendingActionText = computed(() =>
  canOperatePointActions.value ? '手动确认' : '--'
)

const canPressurize = (status: string) =>
  canOperatePointActions.value && status === 'pending' && !manualModeWithoutPressDevice.value

const canConfirm = (status: string) =>
  canOperatePointActions.value && (status === 'stabilizing' || (calibrationStore.controlMode === 'manual' && status === 'pending'))

const canCollect = (status: string) =>
  canOperatePointActions.value && (
    status === 'stabilizing' ||
    status === 'completed' ||
    status === 'error'
  )

const getConfirmButtonType = (status: string) => {
  if (manualModeWithoutPressDevice.value && status === 'pending') {
    return 'primary'
  }
  return 'warning'
}

const getChannelClass = (row: TableRow, index: number) => {
  const value = row.channelValues[index]
  if (value === undefined) return ''
  const diff = Math.abs(value - row.targetValue)
  if (diff < 0.1) return 'channel-good'
  if (diff < 0.5) return 'channel-warning'
  return 'channel-error'
}

// 导出 CSV
const exportCSV = () => {
  const points = calibrationStore.pressurePoints
  if (points.length === 0) {
    ElMessage.warning('没有可导出的数据')
    return
  }

  const channels = calibrationStore.selectedChannels
  const headers = ['序号', '目标压力', '实际压力', '状态', ...channels.map(ch => `CH${ch}`)]
  const rows = points.map(p => [
    p.index,
    p.targetPressure.toFixed(2),
    p.actualPressure?.toFixed(2) || '--',
    getPointStatusText(p.status),
    ...channels.map(ch => p.collectedData?.[ch - 1]?.toFixed(precision.value) || '--')
  ])

  const csvContent = [headers.join(','), ...rows.map(r => r.join(','))].join('\n')
  const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `calibration_data_${new Date().toISOString().split('T')[0]}.csv`
  link.click()
  ElMessage.success('报告已导出')
}
</script>

<style scoped lang="scss">
/* ── 设计系统令牌 ── */
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
$green: #22c55e;
$red: #ef4444;
$blue: #3b82f6;
$amber: #f59e0b;

@use "@/styles/calibration-table" as *;

.data-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
  flex: 1;
}

.table-panel {
  background: #ffffff;
  border-radius: 12px;
  border: 1px solid $slate-200;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  padding: 12px 16px 16px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  font-family: $font-sans;
  transition: box-shadow 0.2s ease, border-color 0.2s ease;

  &:hover {
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.07), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
  }
}

.panel-points {
  flex-shrink: 0;
}

.panel-results {
  flex: 1;
  min-height: 0;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  flex-shrink: 0;
  padding-bottom: 8px;
  border-bottom: 1px solid $slate-100;
}

.table-title {
  display: flex;
  align-items: center;
  gap: 8px;

  .el-icon {
    color: $mint;
    font-size: 16px;
  }

  h3 {
    color: $slate-700;
    margin: 0;
    font-size: 14px;
    font-weight: 600;
  }
}

/* 记录徽章：按 Tags 规范 */
.record-count {
  padding: 1px 8px;
  background: rgba(107, 114, 128, 0.12);
  border: 1px solid rgba(107, 114, 128, 0.25);
  color: $slate-500;
  font-size: 11px;
  font-weight: 500;
  border-radius: 4px;
  margin-left: 4px;
}

.table-actions {
  display: flex;
  gap: 8px;

  .el-button {
    height: 28px;
    padding: 0 12px;
    font-size: 12px;
    font-weight: 500;
    border-radius: 8px;
  }
}

.table-body {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  border-radius: 8px;
}

.point-table-body {
  max-height: 280px;
}

.data-table-body {
  min-height: 220px;
}

.point-operation-table {
  width: 100%;

  @include calibration-table-deep-styles;
}

.data-table {
  width: 100%;
  height: 100%;

  @include calibration-table-deep-styles;

  .channel-good, .channel-warning, .channel-error {
    font-family: $font-mono;
  }
  .channel-good { color: $green; }
  .channel-warning { color: $amber; }
  .channel-error { color: $red; }
}

.row-actions { display: flex; gap: 6px; }

.row-btn {
  min-width: 56px;
}

.collecting-text { color: $slate-400; font-size: 12px; }

.idle-text { color: $slate-400; font-size: 12px; }

.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px 0;
  color: $slate-400;

  .empty-icon {
    font-size: 48px;
    color: $slate-200;
  }

  p { font-size: 13px; margin: 0; font-family: $font-sans; }
}
</style>
