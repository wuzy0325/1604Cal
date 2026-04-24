<template>
  <div class="data-table-wrapper">
    <div class="table-header">
      <div class="table-title">
        <el-icon><DataLine /></el-icon>
        <h3>实时采样数据</h3>
        <span
          v-if="tableRows.length > 0"
          class="record-count"
        >{{ rows.length }} 条采样</span>
      </div>
      <div class="table-actions">
        <button
          type="button"
          class="action-btn"
          disabled
        >
          报告模板
        </button>
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

    <div
      v-if="tableRows.length === 0"
      class="empty-state"
    >
      <el-icon class="empty-icon">
        <SetUp />
      </el-icon>
      <p>请先选择采集通道并开始采样</p>
      <p class="empty-hint">
        当前页面仅展示实时采样结果，不生成测点压力表。
      </p>
    </div>

    <div
      v-else
      class="table-scroll"
    >
      <table class="data-table">
        <thead>
          <tr>
            <th class="col-index">
              序号
            </th>
            <th class="col-pressure">
              平均压力
            </th>
            <th
              v-for="ch in visibleChannels"
              :key="ch"
              class="col-channel"
            >
              CH{{ ch }}
            </th>
            <th class="col-time">
              时间
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in tableRows"
            :key="row.index"
          >
            <td class="cell-index">
              {{ row.index }}
            </td>
            <td class="cell-pressure">
              {{ row.actualPressure }}
            </td>
            <td
              v-for="ch in visibleChannels"
              :key="`${row.index}-${ch}`"
              class="cell-channel"
            >
              {{ row.channelValues[ch] ?? '--' }}
            </td>
            <td class="cell-time">
              {{ row.collectTime }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { DataLine, Download, SetUp } from '@element-plus/icons-vue'
import type { CollectedRow } from '@/stores/measurement'

const props = defineProps<{
  rows: CollectedRow[]
  channels: number[]
}>()

defineEmits<{ 'export-csv': [] }>()

interface DisplayRow {
  index: number
  actualPressure: string
  channelValues: Record<number, string>
  collectTime: string
}

function formatTime(timestamp: string): string {
  try {
    return new Date(timestamp).toLocaleTimeString('zh-CN', { hour12: false })
  } catch {
    return '--:--:--'
    }
}

function estimateActualPressure(row: CollectedRow | undefined): string {
  if (!row) {
    return '--'
  }

  const sourceChannels = props.channels.length > 0 ? props.channels : visibleChannels.value
  const values = sourceChannels
    .map(channel => row.channels[String(channel)])
    .filter((value): value is number => typeof value === 'number' && Number.isFinite(value))

  if (values.length === 0) {
    return '--'
  }

  const average = values.reduce((sum, value) => sum + value, 0) / values.length
  return average.toFixed(3)
}

const visibleChannels = computed(() => {
  if (props.channels.length > 0) {
    return props.channels
  }

  const channelSet = new Set<number>()
  for (const row of props.rows) {
    for (const key of Object.keys(row.channels)) {
      const channel = Number(key)
      if (Number.isInteger(channel)) {
        channelSet.add(channel)
      }
    }
  }

  return Array.from(channelSet).sort((left, right) => left - right)
})

const tableRows = computed<DisplayRow[]>(() => {
  return props.rows.map((row, index) => {
    const channelValues: Record<number, string> = {}

    for (const channel of visibleChannels.value) {
      const raw = row.channels[String(channel)]
      if (typeof raw !== 'number' || Number.isNaN(raw)) {
        channelValues[channel] = '--'
      } else {
        channelValues[channel] = raw.toFixed(3)
      }
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
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-xs) var(--spacing-sm);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-xs);
  flex-shrink: 0;
  padding-bottom: var(--spacing-xs);
  border-bottom: 1px solid var(--border-color);
}

.table-title {
  display: flex;
  align-items: center;
  gap: 6px;

  .el-icon {
    color: var(--accent-primary);
    font-size: 14px;
  }

  h3 {
    color: var(--text-primary);
    margin: 0;
    font-size: 14px;
    font-weight: 600;
  }
}

.record-count {
  color: var(--text-muted);
  font-size: 11px;
  margin-left: 4px;
}

.table-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.action-btn {
  height: 28px;
  padding: 0 10px;
  border: 1px solid var(--border-color-strong);
  border-radius: var(--radius-sm);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;

  &:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
}

.table-scroll {
  flex: 1;
  min-height: 0;
  overflow: auto;
}

.data-table {
  width: max-content;
  min-width: 100%;
  border-collapse: collapse;

  th,
  td {
    border: 1px solid var(--border-color-strong);
    text-align: center;
    white-space: nowrap;
  }

  th {
    background: var(--bg-quaternary);
    color: var(--text-secondary);
    font-size: 11px;
    font-weight: 600;
    padding: 4px 6px;
  }

  td {
    background: var(--bg-tertiary);
    color: var(--text-primary);
    font-size: 11px;
    padding: 3px 4px;
    font-family: 'Consolas', monospace;
    line-height: 1.1;
  }
}

.col-index,
.col-status,
.col-target,
.col-pressure {
  width: 72px;
}

.col-time {
  width: 110px;
}

.col-channel {
  width: 58px;
}

.status-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 52px;
  height: 20px;
  border-radius: 999px;
  font-size: 10px;
  font-weight: 600;
}

.status-tag.pending {
  background: color-mix(in srgb, var(--text-muted) 18%, transparent);
  color: var(--text-muted);
}

.status-tag.running {
  background: var(--status-info-bg);
  color: var(--status-info);
}

.status-tag.paused {
  background: var(--status-warning-bg);
  color: var(--status-warning);
}

.status-tag.completed {
  background: var(--status-success-bg);
  color: var(--status-success);
}

.status-tag.error {
  background: var(--status-error-bg);
  color: var(--status-error);
}

.channel-value.type-good {
  color: var(--status-success);
}

.channel-value.type-warn {
  color: var(--status-warning);
}

.channel-value.type-error {
  color: var(--status-error);
}

.row-running td {
  background: color-mix(in srgb, var(--status-info) 10%, var(--bg-tertiary));
}

.row-completed td {
  background: color-mix(in srgb, var(--status-success) 10%, var(--bg-tertiary));
}

.row-error td {
  background: color-mix(in srgb, var(--status-error) 10%, var(--bg-tertiary));
}

.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-xl);
  color: var(--text-muted);

  .empty-icon {
    font-size: 48px;
    color: var(--bg-quaternary);
  }

  p {
    font-size: 13px;
    margin: 0;
  }
}

.empty-hint {
  color: var(--text-muted);
  opacity: 0.85;
}
</style>
