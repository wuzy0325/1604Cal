<template>
  <div class="pressure-point-list">
    <div class="list-header">
      <h4>压力点设置</h4>
      <div class="actions">
        <div class="point-count">
          <label>压力点个数:</label>
          <el-input-number v-model="pointCount" :min="1" :max="50" size="small" />
        </div>
        <div class="progress">
          <label>完成进度:</label>
          <el-progress :percentage="progressPercent" :stroke-width="8" style="width: 120px" />
        </div>
      </div>
    </div>
    
    <el-table :data="points" border stripe class="point-table">
      <el-table-column type="index" label="序号" width="60" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)" size="small">
            {{ getStatusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="目标压力" width="150">
        <template #default="{ row }">
          <el-input-number 
            v-model="row.targetPressure" 
            :precision="2" 
            :step="0.1"
            size="small"
          />
        </template>
      </el-table-column>
      <el-table-column label="打压/确认" width="120">
        <template #default="{ row }">
          <el-button 
            v-if="row.status === 'pending_press'"
            type="primary" 
            size="small"
            @click="pressurize(row)"
          >
            打压
          </el-button>
          <el-button 
            v-else-if="row.status === 'pending_confirm'"
            type="success" 
            size="small"
            @click="confirm(row)"
          >
            确认
          </el-button>
          <span v-else class="done-text">--</span>
        </template>
      </el-table-column>
      <el-table-column label="采集" width="120">
        <template #default="{ row }">
          <el-button 
            v-if="row.status === 'pending_collect'"
            type="primary" 
            size="small"
            @click="collect(row)"
          >
            采集
          </el-button>
          <el-button 
            v-else-if="row.status === 'completed'"
            type="warning" 
            link
            size="small"
            @click="recollect(row)"
          >
            重新采集
          </el-button>
          <span v-else class="wait-text">等待中</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="80">
        <template #default="{ row, $index }">
          <el-button type="danger" link size="small" @click="removePoint($index)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

type PointStatus = 'pending_press' | 'pending_confirm' | 'pending_collect' | 'completed'

interface PressurePoint {
  targetPressure: number
  status: PointStatus
}

const pointCount = ref(5)
const points = ref<PressurePoint[]>([
  { targetPressure: 10, status: 'completed' },
  { targetPressure: 20, status: 'completed' },
  { targetPressure: 30, status: 'pending_collect' },
  { targetPressure: 40, status: 'pending_confirm' },
  { targetPressure: 50, status: 'pending_press' }
])

const progressPercent = computed(() => {
  const completed = points.value.filter(p => p.status === 'completed').length
  return Math.round((completed / points.value.length) * 100)
})

const getStatusType = (status: PointStatus) => {
  const map: Record<PointStatus, string> = {
    pending_press: 'info',
    pending_confirm: 'warning',
    pending_collect: 'primary',
    completed: 'success'
  }
  return map[status]
}

const getStatusText = (status: PointStatus) => {
  const map: Record<PointStatus, string> = {
    pending_press: '待打压',
    pending_confirm: '待确认',
    pending_collect: '待采集',
    completed: '完成'
  }
  return map[status]
}

const pressurize = (row: PressurePoint) => {
  console.log('打压:', row.targetPressure)
  row.status = 'pending_confirm'
}

const confirm = (row: PressurePoint) => {
  console.log('确认压力:', row.targetPressure)
  row.status = 'pending_collect'
}

const collect = (row: PressurePoint) => {
  console.log('采集数据:', row.targetPressure)
  row.status = 'completed'
}

const recollect = (row: PressurePoint) => {
  console.log('重新采集:', row.targetPressure)
  row.status = 'pending_collect'
}

const removePoint = (index: number) => {
  points.value.splice(index, 1)
}
</script>

<style scoped lang="scss">
.pressure-point-list {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  
  .list-header {
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
      gap: var(--spacing-lg);
      
      .point-count,
      .progress {
        display: flex;
        align-items: center;
        gap: var(--spacing-sm);
        
        label {
          color: var(--text-secondary);
          font-size: 13px;
        }
      }
    }
  }
  
  .point-table {
    :deep(th) {
      background: var(--bg-tertiary);
      color: var(--text-secondary);
    }
    
    :deep(td) {
      color: var(--text-primary);
    }
    
    .done-text,
    .wait-text {
      color: var(--text-muted);
      font-size: 13px;
    }
  }
}
</style>
