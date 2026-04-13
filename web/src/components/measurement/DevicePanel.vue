<template>
  <div class="device-panel">
    <div
      class="panel-header"
      @click="toggleCollapse"
    >
      <span class="title">{{ title }}</span>
      <el-icon :class="{ 'is-collapsed': isCollapsed }">
        <ArrowDown />
      </el-icon>
    </div>
    <div
      v-show="!isCollapsed"
      class="panel-content"
    >
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ArrowDown } from '@element-plus/icons-vue'

defineProps<{
  title: string
}>()

const isCollapsed = ref(false)

const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value
}
</script>

<style scoped lang="scss">
.device-panel {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  margin-bottom: var(--spacing-md);
  
  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--spacing-sm) var(--spacing-md);
    cursor: pointer;
    user-select: none;
    
    .title {
      color: var(--text-primary);
      font-weight: 500;
    }
    
    .el-icon {
      color: var(--text-secondary);
      transition: transform 0.3s;
      
      &.is-collapsed {
        transform: rotate(-90deg);
      }
    }
  }
  
  .panel-content {
    padding: 0 var(--spacing-md) var(--spacing-md);
  }
}
</style>
