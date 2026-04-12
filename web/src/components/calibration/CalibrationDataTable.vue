<template>
  <div class="data-table-panel">
    <div class="panel-header">
      <h4>采集数据</h4>
      <div class="actions">
        <span class="record-count">记录数: {{ points.length }}</span>
        <el-button type="primary" @click="exportData">
          <el-icon><Download /></el-icon>
          导出CSV
        </el-button>
      </div>
    </div>
    
    <el-table :data="tableData" border stripe class="data-table">
      <el-table-column prop="point" label="压力点" width="80" />
      <el-table-column prop="targetPressure" label="目标压力" width="120" />
      <el-table-column 
        v-for="ch in selectedChannels" 
        :key="ch"
        :label="`CH${ch}`" 
        width="100"
      >
        <template #default="{ row }">
          {{ row.channelData[ch - 1]?.toFixed(4) || '--' }}
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'completed' ? 'success' : 'info'" size="small">
            {{ row.status === 'completed' ? '已采集' : '待采集' }}
          </el-tag>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Download } from '@element-plus/icons-vue'
import type { PressurePoint } from '@/stores/calibration'

interface TableData {
  point: number
  targetPressure: number
  channelData: number[]
  status: string
}

const props = defineProps<{
  points: PressurePoint[]
  selectedChannels: number[]
}>()

// 将PressurePoint转换为表格数据
const tableData = computed<TableData[]>(() => {
  return props.points.map((point, index) => ({
    point: index + 1,
    targetPressure: point.targetPressure,
    channelData: point.collectedData || [],
    status: point.status
  }))
})

const exportData = () => {
  console.log('导出数据')
  // 生成CSV并下载
  const headers = ['压力点', '目标压力', ...props.selectedChannels.map(ch => `CH${ch}`), '状态']
  const rows = tableData.value.map(row => [
    row.point,
    row.targetPressure,
    ...props.selectedChannels.map(ch => row.channelData[ch - 1]?.toFixed(4) || '--'),
    row.status === 'completed' ? '已采集' : '待采集'
  ])
  
  const csvContent = [headers.join(','), ...rows.map(r => r.join(','))].join('\n')
  const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `calibration_data_${new Date().toISOString().split('T')[0]}.csv`
  link.click()
}
</script>

<style scoped lang="scss">
.data-table-panel {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  
  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--spacing-md);
    
    h4 {
      color: var(--text-primary);
      margin: 0;
    }
    
    .actions {
      display: flex;
      align-items: center;
      gap: var(--spacing-md);
      
      .record-count {
        color: var(--text-secondary);
        font-size: 13px;
      }
    }
  }
  
  .data-table {
    :deep(th) {
      background: var(--bg-tertiary);
      color: var(--text-secondary);
    }
    
    :deep(td) {
      color: var(--text-primary);
    }
  }
}
</style>
