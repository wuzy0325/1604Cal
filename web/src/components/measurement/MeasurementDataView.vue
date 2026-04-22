<template>
  <div class="data-table-wrapper">
    <div class="table-header">
      <div class="table-title">
        <el-icon><DataLine /></el-icon>
        <h3>采集数据</h3>
        <span v-if="rows.length > 0" class="record-count">{{ rows.length }} 行数据</span>
      </div>
      <div class="table-actions">
        <el-button size="small" :disabled="rows.length === 0" @click="$emit('export-csv')"><el-icon><Download /></el-icon>导出CSV</el-button>
      </div>
    </div>
    <el-table v-if="rows.length > 0" :data="rows" border stripe class="data-table" max-height="400">
      <el-table-column label="时间" width="200">
        <template #default="{ row }">{{ formatTime(row.timestamp) }}</template>
      </el-table-column>
      <el-table-column v-for="ch in channels" :key="ch" :label="`CH${ch}`" width="100">
        <template #default="{ row }">{{ row.channels[String(ch)]?.toFixed(4) || '--' }}</template>
      </el-table-column>
    </el-table>
    <div v-else class="empty-state"><el-icon class="empty-icon"><SetUp /></el-icon><p>点击「开始采集」后数据将实时显示在此处</p></div>
  </div>
</template>

<script setup lang="ts">
import { DataLine, Download, SetUp } from '@element-plus/icons-vue'
import type { CollectedRow } from '@/stores/measurement'

defineProps<{
  rows: CollectedRow[]
  channels: number[]
}>()
defineEmits<{ 'export-csv': [] }>()

const formatTime = (ts: string) => {
  try { return new Date(ts).toLocaleTimeString('zh-CN', { hour12: false }) }
  catch { return ts }
}
</script>

<style scoped lang="scss">
.data-table-wrapper { flex: 1; min-height: 0; background: var(--bg-secondary); border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: var(--spacing-md); display: flex; flex-direction: column; overflow: hidden; }
.table-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--spacing-sm); flex-shrink: 0; padding-bottom: var(--spacing-sm); border-bottom: 1px solid var(--border-color); }
.table-title { display: flex; align-items: center; gap: var(--spacing-sm); .el-icon { color: var(--accent-primary); font-size: 16px; } h3 { color: var(--text-primary); margin: 0; font-size: 14px; font-weight: 600; } }
.record-count { color: var(--text-muted); font-size: 12px; margin-left: var(--spacing-xs); }
.table-actions { display: flex; gap: var(--spacing-xs); }
.data-table {
  width: 100%; flex: 1; min-height: 0;
  :deep(.el-table) { --el-table-bg-color: var(--bg-tertiary); --el-table-tr-bg-color: var(--bg-tertiary); --el-table-header-bg-color: var(--bg-quaternary); --el-table-row-hover-bg-color: rgba(255, 215, 0, 0.06); --el-table-current-row-bg-color: rgba(255, 215, 0, 0.1); --el-table-border-color: var(--border-color-strong); --el-table-text-color: var(--text-primary); --el-table-header-text-color: var(--text-secondary); background-color: var(--bg-tertiary); }
  :deep(th.el-table__cell) { background: var(--bg-quaternary) !important; color: var(--text-secondary) !important; font-size: 12px; font-weight: 600; padding: 8px 0; }
  :deep(td.el-table__cell) { background: var(--bg-tertiary); color: var(--text-primary); font-size: 13px; padding: 6px 0; font-family: 'Consolas', monospace; }
}
.empty-state { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: var(--spacing-sm); padding: var(--spacing-xl); color: var(--text-muted); .empty-icon { font-size: 48px; color: var(--bg-quaternary); } p { font-size: 13px; margin: 0; } }
</style>
