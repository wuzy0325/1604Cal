<template>
  <div class="data-table-wrapper">
    <!-- 压力点表 -->
    <div v-if="points.length > 0" class="points-section">
      <div class="table-toolbar">
        <div class="toolbar-title">
          <el-icon class="toolbar-icon"><Aim /></el-icon>
          <h3>目标压力表数据清单</h3>
          <span class="record-badge">{{ points.length }} 个测点</span>
        </div>
        <div class="toolbar-legend">
          <span class="legend-item">
            <span class="legend-dot pending" />
            待采集
          </span>
          <span class="legend-item">
            <span class="legend-dot completed" />
            已采集
          </span>
        </div>
      </div>
      <div class="table-scroll custom-scroll">
        <table class="data-table">
          <thead>
            <tr>
              <th class="col-index">序号</th>
              <th class="col-status">状态</th>
              <th class="col-target">目标值</th>
              <th v-for="ch in channelCount" :key="ch" class="col-channel">{{ ch }}</th>
              <th class="col-time">采集时间</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="pt in points"
              :key="pt.id"
              :class="getRowClass(pt)"
              class="data-row"
            >
              <td class="cell-index">
                <div class="index-cell-wrap">
                  <span>{{ pt.index }}</span>
                  <span
                    v-if="isRoundTrip"
                    :class="['trip-badge', pt.direction === 'forward' ? 'forward' : 'backward']"
                  >
                    {{ pt.direction === 'forward' ? '正' : '回' }}
                  </span>
                </div>
              </td>
              <td class="cell-status">
                <button
                  v-if="controlMode === ControlMode.Manual"
                  type="button"
                  class="row-collect-btn"
                  @click="$emit('collect-point', pt.index)"
                >
                  采集
                </button>
                <span v-else :class="['status-tag', getStatusType(pt.status)]">
                  <span :class="['status-dot', getStatusType(pt.status)]" />
                  {{ getStatusText(pt.status) }}
                </span>
              </td>
              <td class="cell-target">
                <input
                  :value="pt.targetPressure"
                  type="number"
                  class="target-input"
                  :step="precisionStep"
                  @change="onTargetChange(pt.id, ($event.target as HTMLInputElement).valueAsNumber)"
                />
              </td>
              <td v-for="ch in channelCount" :key="ch" class="cell-channel">
                <div
                  v-if="pt.collectedData && pt.collectedData[ch - 1] !== undefined"
                  :class="['channel-value', { 'channel-over-limit': isChannelOverLimit(pt, ch) }]"
                >
                  {{ pt.collectedData[ch - 1].toFixed(precisionForDisplay) }}
                </div>
                <div v-else class="channel-value empty">--</div>
              </td>
              <td class="cell-time">
                <span v-if="pt.collectTime" class="time-display">{{ formatTime(pt.collectTime) }}</span>
                <span v-else class="time-display empty">--:--:--</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-else class="empty-table-state">
      <el-empty description="请配置参数并生成压力表" :image-size="80">
        <p class="empty-hint">设置最小值、最大值和点数后点击"生成压力表"</p>
      </el-empty>
    </div>

    <!-- 实时采样数据 -->
    <div v-if="tableRows.length > 0" class="sample-section">
      <div class="table-toolbar">
        <div class="toolbar-title">
          <el-icon class="toolbar-icon"><DataLine /></el-icon>
          <h3>实时采样数据</h3>
          <span class="record-badge">{{ rows.length }} 条采样</span>
        </div>
        <div class="toolbar-actions">

        </div>
      </div>
      <div class="table-scroll custom-scroll">
        <table class="data-table">
          <thead>
            <tr>
              <th class="col-index">序号</th>
              <th class="col-pressure">平均压力</th>
              <th v-for="ch in visibleChannels" :key="ch" class="col-channel">CH{{ ch }}</th>
              <th class="col-time">时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in tableRows" :key="row.index" class="data-row">
              <td class="cell-index">{{ row.index }}</td>
              <td class="cell-pressure">{{ row.actualPressure }}</td>
              <td v-for="ch in visibleChannels" :key="`${row.index}-${ch}`" class="cell-channel">
                <div :class="['channel-value', { 'channel-over-limit': isSampleOverLimit(row, ch) }]">
                  {{ row.channelValues[ch] ?? '--' }}
                </div>
              </td>
              <td class="cell-time">{{ row.collectTime }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { DataLine, Download, Aim } from '@element-plus/icons-vue'
import { useMeasurementStore } from '@/stores/measurement'
import type { CollectedRow } from '@/stores/measurement'
import type { MeasurementPoint } from '@/api/measurement'
import { ControlMode, PressureMode } from '@/types/calibration'

const props = defineProps<{
  rows?: CollectedRow[]
  channels?: number[]
  controlMode?: string
}>()

defineEmits<{
  'collect-point': [pointIndex: number]
}>()

const measurementStore = useMeasurementStore()

const rows = computed(() => props.rows ?? measurementStore.rows)
const channels = computed(() => props.channels ?? measurementStore.channels)
const points = computed<MeasurementPoint[]>(() => measurementStore.points)

const isRoundTrip = computed(() => measurementStore.measurementParams.pressureMode === PressureMode.RoundTrip)
const precisionForDisplay = computed(() => measurementStore.measurementParams.precision)
const channelCount = 16
const currentPointIndex = computed(() => measurementStore.currentPointIndex)

const precisionStep = computed(() => Math.pow(10, -(measurementStore.measurementParams.precision || 2)))

const alarmEnabled = computed(() => measurementStore.alarmConfig.enabled)
const precisionLevel = computed(() => measurementStore.measurementParams.precisionLevel)
const minPressure = computed(() => measurementStore.measurementParams.minPressure)
const maxPressure = computed(() => measurementStore.measurementParams.maxPressure)

function getPointOverLimitChannels(pt: MeasurementPoint): number[] {
  if (!alarmEnabled.value) return []
  if (!pt.collectedData || pt.collectedData.length === 0) return []

  const span = Math.abs(maxPressure.value - minPressure.value)
  let allowance: number
  
  if (span > 1e-10) {
    allowance = span * precisionLevel.value
  } else {
    allowance = Math.abs(pt.targetPressure) * precisionLevel.value
  }

  const overLimit: number[] = []
  for (let ch = 1; ch <= pt.collectedData.length; ch++) {
    const collectedVal = pt.collectedData[ch - 1]
    const deviation = Math.abs(collectedVal - pt.targetPressure)

    if (deviation > allowance) {
      overLimit.push(ch)
    }
  }
  return overLimit
}

function isChannelOverLimit(pt: MeasurementPoint, ch: number): boolean {
  const overLimit = getPointOverLimitChannels(pt)
  return overLimit.includes(ch)
}

function isSampleOverLimit(row: DisplayRow, ch: number): boolean {
  if (!alarmEnabled.value) return false
  const cv = row.channelValues[ch]
  if (!cv || cv === '--') return false
  const raw = parseFloat(cv)
  if (isNaN(raw)) return false
  const target = currentTargetPressure.value
  if (target === undefined) return false

  const span = Math.abs(maxPressure.value - minPressure.value)
  let allowance: number
  if (span > 1e-10) {
    allowance = span * precisionLevel.value
  } else {
    allowance = Math.abs(target) * precisionLevel.value
  }

  return Math.abs(raw - target) > allowance
}

const currentTargetPressure = computed(() => {
  const idx = measurementStore.currentPointIndex
  if (idx <= 0 || idx > measurementStore.points.length) return undefined
  return measurementStore.points[idx - 1]?.targetPressure
})

function getRowClass(pt: MeasurementPoint): string {
  const classes: string[] = []
  if (pt.status === 'completed') classes.push('row-completed')
  if (currentPointIndex.value === pt.index) classes.push('row-current')
  return classes.join(' ')
}

function getStatusType(status: string): string {
  const map: Record<string, string> = {
    pending: 'pending',
    pressuring: 'processing',
    pressurizing: 'processing',
    stabilizing: 'processing',
    collecting: 'processing',
    completed: 'completed',
    error: 'error'
  }
  return map[status] || 'pending'
}

function getStatusText(status: string): string {
  const map: Record<string, string> = {
    pending: '待采集',
    pressuring: '打压中',
    pressurizing: '打压中',
    stabilizing: '稳定中',
    collecting: '采集中',
    completed: '已完成',
    error: '出错'
  }
  return map[status] || status
}

function onTargetChange(pointId: string, value: number | null) {
  if (value === null || isNaN(value)) return
  measurementStore.updatePointTarget(pointId, value)
}

function formatTime(timestamp: string): string {
  try {
    return new Date(timestamp).toLocaleTimeString('zh-CN', { hour12: false })
  } catch {
    return '--:--:--'
  }
}

function estimateActualPressure(row: CollectedRow | undefined): string {
  if (!row) return '--'
  const sourceChannels = channels.value.length > 0 ? channels.value : visibleChannels.value
  const values = sourceChannels
    .map(ch => row.channels[String(ch)])
    .filter((value): value is number => typeof value === 'number' && Number.isFinite(value))
  if (values.length === 0) return '--'
  const average = values.reduce((sum, val) => sum + val, 0) / values.length
  return average.toFixed(3)
}

const visibleChannels = computed(() => {
  if (channels.value.length > 0) return channels.value
  const channelSet = new Set<number>()
  for (const row of rows.value) {
    for (const key of Object.keys(row.channels)) {
      const ch = Number(key)
      if (Number.isInteger(ch)) channelSet.add(ch)
    }
  }
  return Array.from(channelSet).sort((a, b) => a - b)
})

interface DisplayRow {
  index: number
  actualPressure: string
  channelValues: Record<number, string>
  collectTime: string
}

const tableRows = computed<DisplayRow[]>(() => {
  return rows.value.map((row, index) => {
    const channelValues: Record<number, string> = {}
    for (const ch of visibleChannels.value) {
      const raw = row.channels[String(ch)]
      channelValues[ch] = (typeof raw === 'number' && !isNaN(raw)) ? raw.toFixed(3) : '--'
    }
    return {
      index: index + 1,
      actualPressure: estimateActualPressure(row),
      channelValues,
      collectTime: formatTime(row.timestamp)
    }
  })
})
</script>

<style scoped lang="scss">
.data-table-wrapper {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  gap: 12px;
  font-family: $font-sans;
}

.points-section,
.sample-section {
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  border: 1px solid $slate-200;
  transition: box-shadow 0.2s ease, border-color 0.2s ease;

  &:hover {
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.07), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
    border-color: rgba(16, 185, 129, 0.3);
  }
}

.points-section {
  flex: 1;
  min-height: 0;
}

.sample-section {
  flex: 1;
  min-height: 0;
}

.table-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  border-bottom: 1px solid $slate-100;
  background: rgba(249, 250, 251, 0.5);
  flex-shrink: 0;
}

.toolbar-title {
  display: flex;
  align-items: center;
  gap: 8px;

  h3 {
    font-size: 14px;
    font-weight: 600;
    color: $slate-700;
    margin: 0;
    font-family: $font-sans;
  }
}

.toolbar-icon {
  font-size: 16px;
  color: $mint;
}

/* 记录徽章：按 Tags 规范 */
.record-badge {
  padding: 2px 8px;
  background: rgba(107, 114, 128, 0.12);
  border: 1px solid rgba(107, 114, 128, 0.25);
  color: $slate-500;
  font-size: 11px;
  font-weight: 500;
  border-radius: 4px;
  margin-left: 4px;
}

.toolbar-legend {
  display: flex;
  align-items: center;
  gap: 16px;
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: $slate-500;
  font-family: $font-sans;
}

.legend-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;

  &.pending { background: $slate-300; }
  &.completed { background: $mint; }
}

.toolbar-actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  height: 28px;
  padding: 0 12px;
  border-radius: 8px;
  border: 1px solid $slate-200;
  background: #fff;
  color: $slate-600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.15s ease;
  font-family: $font-sans;

  &:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  &:hover:not(:disabled) {
    background: $slate-50;
    border-color: $slate-300;
  }
}

.table-scroll {
  overflow-x: auto;
  overflow-y: auto;
  flex: 1;
}

.custom-scroll::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.custom-scroll::-webkit-scrollbar-track {
  background: $slate-100;
}

.custom-scroll::-webkit-scrollbar-thumb {
  background: $slate-300;
  border-radius: 10px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
  white-space: nowrap;

  thead {
    position: sticky;
    top: 0;
    z-index: 10;
    background: #fff;
    border-bottom: 1px solid $slate-200;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
    color: $slate-400;
    text-transform: uppercase;
  }

  th {
    padding: 8px 12px;
    text-align: center;
    font-weight: 600;
    font-size: 12px;
    letter-spacing: 0.05em;
    font-family: $font-sans;
  }

  td {
    padding: 6px 12px;
    text-align: center;
    color: $slate-600;
  }

  tbody {
    color: $slate-600;
  }
}

.data-row {
  height: 40px;
  transition: background 0.15s ease;
  border-bottom: 1px solid $slate-50;

  &:hover {
    background: $slate-50;
  }
}

.col-index {
  width: 72px;
  min-width: 72px;
  text-align: left;
}

.col-status {
  width: 100px;
  min-width: 100px;
  text-align: left;
}

.col-target {
  width: 80px;
  min-width: 80px;
}

.col-channel {
  min-width: 52px;
  text-align: center;
  font-family: $font-mono;
  font-size: 11px;
}

.col-time {
  text-align: right;
  padding-right: 24px;
}

.cell-index {
  color: $slate-400;
  font-weight: 600;
  text-align: left;
  font-family: $font-mono;
}

.cell-status {
  text-align: left;
}

.cell-target {
  font-family: $font-mono;
  font-size: 13px;
  font-weight: 700;
  color: $slate-700;
}

.cell-time {
  text-align: right;
  padding-right: 24px;
  font-family: $font-mono;
  font-size: 11px;
  color: $slate-400;
}

.index-cell-wrap {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

/* 正/回标签 */
.trip-badge {
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 4px;
  font-weight: 600;

  &.forward {
    background: rgba(59, 130, 246, 0.12);
    border: 1px solid rgba(59, 130, 246, 0.25);
    color: $blue;
  }

  &.backward {
    background: rgba(245, 158, 11, 0.12);
    border: 1px solid rgba(245, 158, 11, 0.25);
    color: #d97706;
  }
}

/* 状态标签：按 Tags 规范 */
.status-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  white-space: nowrap;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
  font-family: $font-sans;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

/* pending: gray */
.status-tag.pending {
  background: rgba(107, 114, 128, 0.1);
  border: 1px solid rgba(107, 114, 128, 0.2);
  color: $slate-500;
}
.status-dot.pending { background: $slate-300; }

/* processing: blue */
.status-tag.processing {
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid rgba(59, 130, 246, 0.2);
  color: $blue;
}
.status-dot.processing { background: $blue; }

/* completed: green */
.status-tag.completed {
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.2);
  color: $mint-dark;
}
.status-dot.completed { background: $mint; }

/* error: red */
.status-tag.error {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.2);
  color: $red;
}
.status-dot.error { background: $red; }

/* 行内采集按钮 */
.row-collect-btn {
  padding: 3px 12px;
  border: 1px solid $mint;
  border-radius: 6px;
  background: linear-gradient(135deg, $mint, $mint-dark);
  color: #fff;
  font-size: 11px;
  font-weight: 600;
  font-family: $font-sans;
  cursor: pointer;
  transition: all 0.15s ease;

  &:hover:not(:disabled) {
    background: linear-gradient(135deg, $mint-light, $mint);
    transform: translateY(-1px);
    box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
  }

  &:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
}

.target-input {
  width: 52px;
  text-align: center;
  border: 1px solid $slate-200;
  border-radius: 8px;
  background: #fff;
  color: $slate-700;
  padding: 4px 6px;
  font-size: 13px;
  font-weight: 600;
  outline: none;
  font-variant-numeric: tabular-nums;
  font-family: $font-mono;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;

  &::-webkit-inner-spin-button,
  &::-webkit-outer-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }

  -moz-appearance: textfield;

  &:focus {
    border-color: $mint;
    box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.15);
  }
}

.channel-value {
  font-variant-numeric: tabular-nums;
  font-family: $font-mono;
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
  border: 1px solid transparent;

  &.empty {
    color: $slate-300;
  }
}

.channel-over-limit {
  background: rgba(239, 68, 68, 0.12) !important;
  border-color: rgba(239, 68, 68, 0.25) !important;
  color: $red;
  font-weight: 700;
}

.time-display {
  font-size: 11px;
  color: $slate-400;

  &.empty {
    color: $slate-300;
  }
}

.row-completed {
  background: rgba(16, 185, 129, 0.04);
}

.row-current {
  background: rgba(16, 185, 129, 0.06);
  border-left: 1px solid $mint;
}

.empty-table-state {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  border: 1px solid $slate-200;
  padding: 40px 0;
}

.empty-hint {
  font-size: 12px;
  color: $slate-400;
  margin-top: 8px;
  font-family: $font-sans;
}

.cell-pressure {
  font-variant-numeric: tabular-nums;
  font-family: $font-mono;
}

.col-pressure {
  width: 80px;
  min-width: 80px;
}

@media (max-width: 768px) {
  .table-toolbar {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .toolbar-legend {
    width: 100%;
  }
}
</style>
