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
  height: 100%;
  background: #f6f7f6;
  overflow: hidden;
  font-family: 'Inter', 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 20px;
  gap: 16px;

  // 仅在内容确实溢出时才显示滚动条
  overflow-y: auto;

  // 滚动条样式
  &::-webkit-scrollbar {
    width: 6px;
  }

  &::-webkit-scrollbar-thumb {
    background: #d1d5db;
    border-radius: 3px;
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
    padding: 16px;
    gap: 16px;
  }
}
</style>
