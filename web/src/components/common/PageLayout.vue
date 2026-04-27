<template>
  <div class="page-layout">
    <!-- 侧边栏 -->
    <Sidebar />
    
    <!-- 主内容区 -->
    <main class="main-content">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
import Sidebar from './Sidebar.vue'
</script>

<style scoped lang="scss">

.page-layout {
  display: flex;
  width: 100%;
  height: 100vh;
  background: $bg-page;
  overflow: hidden;
  font-family: $font-family-base;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: $spacing-8;
  gap: $spacing-6;
  overflow-y: auto;
  
  // 滚动条样式
  &::-webkit-scrollbar {
    width: 8px;
  }
  
  &::-webkit-scrollbar-thumb {
    background: $neutral-700;
    border-radius: 4px;
  }
  
  &::-webkit-scrollbar-track {
    background: transparent;
  }
}

// 响应式适配
@media (max-width: 768px) {
  .page-layout {
    flex-direction: column;
  }
  
  .main-content {
    padding: $spacing-4;
    gap: $spacing-4;
  }
}
</style>
