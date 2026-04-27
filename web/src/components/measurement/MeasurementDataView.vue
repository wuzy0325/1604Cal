<template>
  <div class="data-table-wrapper">
    <!-- 压力点表 -->
    <div v-if="points.length > 0" class="points-section">
      <div class="table-header">
        <div class="table-title">
          <el-icon><Aim /></el-icon>
          <h3>目标压力表与数据采集</h3>
          <span class="record-count">{{ points.length }} 个测点</span>
        </div>
      </div>
      <div class="table-scroll">
        <table class="data-table">
          <thead>
            <tr>
              <th class="col-index">序号</th>
              <th class="col-status">状态</th>
              <th class="col-target">目标值</th>
              <th v-for="ch in channelCount" :key="ch" class="col-channel">{{ ch }}</th>
              <th class="col-time">时间</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="pt in points"
              :key="pt.id"
              :class="getRowClass(pt)"
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
                <span :class="['status-tag', getStatusType(pt.status)]">
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
                <div v-if="pt.collectedData && pt.collectedData[ch - 1] !== undefined" class="channel-value">
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
      <div class="table-header">
        <div class="table-title">
          <el-icon><DataLine /></el-icon>
          <h3>实时采样数据</h3>
          <span class="record-count">{{ rows.length }} 条采样</span>
        </div>
        <div class="table-actions">
          <button
            type="button"
            class="action-btn"
            :disabled="rows.length === 0"
            @click="$emit('export-csv')"
          >
            <el-icon><Download /></el-icon>
            导出CSV
          </button>
        </div>
      </div>
      <div class="table-scroll">
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
            <tr v-for="row in tableRows" :key="row.index">
              <td class="cell-index">{{ row.index }}</td>
              <td class="cell-pressure">{{ row.actualPressure }}</td>
              <td v-for="ch in visibleChannels" :key="`${row.index}-${ch}`" class="cell-channel">
                {{ row.channelValues[ch] ?? '--' }}
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
import { DataLine, Download, SetUp, Aim } from '@element-plus/icons-vue'
import { useMeasurementStore } from '@/stores/measurement'
import type { CollectedRow } from '@/stores/measurement'
import type { MeasurementPoint } from '@/api/measurement'

const props = defineProps<{
  rows?: CollectedRow[]
  channels?: number[]
}>()

const emit = defineEmits<{ 'export-csv': [] }>()

const measurementStore = useMeasurementStore()

const rows = computed(() => props.rows ?? measurementStore.rows)
const channels = computed(() => props.channels ?? measurementStore.channels)
const points = computed<MeasurementPoint[]>(() => measurementStore.points)

const isRoundTrip = computed(() => measurementStore.measurementParams.pressureMode === 'roundTrip')
const precisionForDisplay = computed(() => measurementStore.measurementParams.precision)
const channelCount = 16

const precisionStep = computed(() => Math.pow(10, -(measurementStore.measurementParams.precision || 2)))

function getRowClass(pt: MeasurementPoint): string {
  return pt.status === 'completed' ? 'row-completed' : ''
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
  const pt = points.value.find(p => p.id === pointId)
  if (pt) {
    pt.targetPressure = value
  }
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
  display: flex;
  flex-direction: column;
  overflow: hidden;
  gap: var(--spacing-sm);
}

.points-section,
.sample-section {
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.points-section {
  flex-shrink: 0;
}

.sample-section {
  flex: 1;
}

.table-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: var(--spacing-xs);
  border-bottom: 1px solid var(--border-color);
  margin-bottom: var(--spacing-xs);
}

.table-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);

  h3 {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
  }
}

.record-count {
  color: var(--text-muted);
  font-size: 12px;
}

.table-actions {
  display: flex;
  gap: var(--spacing-xs);
}

.action-btn {
  height: 26px;
  padding: 0 10px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color-strong);
  background: var(--bg-secondary);
  color: var(--text-secondary);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;

  &:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
}

.table-scroll {
  overflow-x: auto;
  overflow-y: auto;
  flex: 1;

  &::-webkit-scrollbar {
    width: 8px;
    height: 8px;
  }

  &::-webkit-scrollbar-thumb {
    background: var(--border-color-strong);
    border-radius: 4px;
  }
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;

  th {
    white-space: nowrap;
    padding: 6px 8px;
    text-align: center;
    background: var(--bg-tertiary);
    color: var(--text-muted);
    font-weight: 500;
    border-bottom: 1px solid var(--border-color);
    position: sticky;
    top: 0;
    z-index: 1;
  }

  td {
    padding: 4px 8px;
    text-align: center;
    border-bottom: 1px solid var(--border-color);
    color: var(--text-secondary);
  }
}

.col-index { width: 44px; min-width: 44px; }
.col-status { width: 72px; min-width: 72px; }
.col-target { width: 80px; min-width: 80px; }
.col-channel { width: 60px; min-width: 60px; }
.col-time { width: 80px; min-width: 80px; }

.index-cell-wrap {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.trip-badge {
  font-size: 10px;
  padding: 0 4px;
  border-radius: 3px;
  font-weight: 600;

  &.forward {
    background: var(--status-info-bg, #e3f2fd);
    color: var(--status-info, #1976d2);
  }

  &.backward {
    background: var(--status-warning-bg, #fff3e0);
    color: var(--status-warning, #f57c00);
  }
}

.status-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  white-space: nowrap;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;

  &.pending { background: var(--text-muted, #9e9e9e); }
  &.processing { background: var(--status-info, #1976d2); }
  &.completed { background: var(--status-success, #388e3c); }
  &.error { background: var(--status-error, #d32f2f); }
}

.target-input {
  width: 70px;
  text-align: center;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  background: var(--bg-primary);
  color: var(--text-primary);
  padding: 2px 4px;
  font-size: 12px;

  &:focus {
    outline: none;
    border-color: var(--accent-primary);
  }
}

.channel-value {
  font-variant-numeric: tabular-nums;
  font-size: 12px;

  &.empty {
    color: var(--text-muted);
  }
}

.time-display {
  font-size: 12px;

  &.empty {
    color: var(--text-muted);
  }
}

.row-completed {
  background: rgba(56, 142, 60, 0.05);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-xl) 0;
  color: var(--text-muted);

  p {
    margin: var(--spacing-xs) 0;
    font-size: 14px;
  }
}

.empty-hint {
  font-size: 12px;
  color: var(--text-muted);
}

.cell-pressure {
  font-variant-numeric: tabular-nums;
}

.col-pressure {
  width: 72px;
  min-width: 72px;
}

@media (max-width: 768px) {
  .table-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-xs);
  }
}
</style>
