<template>
  <div class="channel-matrix">
    <div class="matrix-header">
      <h4>通道选择</h4>
      <div class="actions">
        <span class="count">已选: {{ selectedCount }}/16</span>
        <el-button type="primary" link size="small" @click="selectAll">全选</el-button>
        <el-button type="danger" link size="small" @click="clearAll">清空</el-button>
      </div>
    </div>
    
    <div class="matrix-grid">
      <div 
        v-for="(selected, index) in channels" 
        :key="index"
        class="channel-item"
        :class="{ selected }"
        @click="toggleChannel(index)"
      >
        <el-checkbox v-model="channels[index]" @click.stop>
          CH{{ index + 1 }}
        </el-checkbox>
      </div>
    </div>
    
    <div v-if="selectedCount === 0" class="warning">
      <el-icon><Warning /></el-icon>
      <span>请至少选择一个通道</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Warning } from '@element-plus/icons-vue'

const channels = ref<boolean[]>(new Array(16).fill(false))

const selectedCount = computed(() => channels.value.filter(Boolean).length)

const toggleChannel = (index: number) => {
  channels.value[index] = !channels.value[index]
}

const selectAll = () => {
  channels.value.fill(true)
}

const clearAll = () => {
  channels.value.fill(false)
}
</script>

<style scoped lang="scss">
.channel-matrix {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  
  .matrix-header {
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
      gap: var(--spacing-sm);
      
      .count {
        color: var(--text-secondary);
        font-size: 13px;
      }
    }
  }
  
  .matrix-grid {
    display: grid;
    grid-template-columns: repeat(8, 1fr);
    gap: var(--spacing-sm);
    
    @media (max-width: 1200px) {
      grid-template-columns: repeat(4, 1fr);
    }
    
    .channel-item {
      background: var(--bg-tertiary);
      border: 2px solid transparent;
      border-radius: var(--radius-sm);
      padding: var(--spacing-sm);
      cursor: pointer;
      transition: all 0.2s;
      
      &:hover {
        border-color: var(--accent-primary);
      }
      
      &.selected {
        background: rgba(16, 185, 129, 0.2);
        border-color: var(--status-success);
      }
      
      .el-checkbox {
        color: var(--text-primary);
        
        :deep(.el-checkbox__input.is-checked + .el-checkbox__label) {
          color: var(--status-success);
        }
      }
    }
  }
  
  .warning {
    display: flex;
    align-items: center;
    gap: var(--spacing-xs);
    color: var(--status-warning);
    font-size: 12px;
    margin-top: var(--spacing-md);
    padding: var(--spacing-sm);
    background: rgba(245, 158, 11, 0.1);
    border-radius: var(--radius-sm);
  }
}
</style>
