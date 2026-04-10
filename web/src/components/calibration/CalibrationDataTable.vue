<template>
  <div class="data-table-panel">
    <div class="panel-header">
      <h4>采集数据</h4>
      <div class="actions">
        <span class="record-count">记录数: {{ data.length }}</span>
        <el-button type="primary" @click="exportData">
          <el-icon><Download /></el-icon>
          导出CSV
        </el-button>
      </div>
    </div>
    
    <el-table :data="data" border stripe class="data-table">
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
          <el-tag :type="row.status === 'collected' ? 'success' : 'info'" size="small">
            {{ row.status === 'collected' ? '已采集' : '待采集' }}
          </el-tag>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { Download } from '@element-plus/icons-vue'

interface CalibrationData {
  point: number
  targetPressure: number
  channelData: number[]
  status: 'collected' | 'pending'
}

const props = defineProps<{
  data: CalibrationData[]
  selectedChannels: number[]
}>()

const exportData = () => {
  console.log('导出数据')
  // 生成CSV并下载
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
