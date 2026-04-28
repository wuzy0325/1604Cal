<template>
  <PageLayout>
    <!-- 页面头部 -->
    <header class="page-header">
      <div class="header-left">
        <button class="back-btn" @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
        </button>
        <div class="header-title">
          <h1>设备管理</h1>
          <p>配置和管理打压设备与计量设备</p>
        </div>
      </div>
    </header>

    <!-- 设备管理面板 -->
    <DeviceManagementPanel class="content-panel" />
  </PageLayout>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import PageLayout from '@/components/common/PageLayout.vue'
import DeviceManagementPanel from '@/components/device/DeviceManagementPanel.vue'

const router = useRouter()

/**
 * 返回首页
 */
function goBack(): void {
  router.push('/')
}
</script>

<style scoped lang="scss">
/* ── 设计系统令牌 ── */
$font-sans: 'Inter', 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
$slate-50: #f9fafb;
$slate-100: #f3f4f6;
$slate-200: #e5e7eb;
$slate-300: #d1d5db;
$slate-400: #9ca3af;
$slate-500: #6b7280;
$slate-600: #4b5563;
$slate-700: #374151;
$slate-800: #1f2937;
$mint: #10b981;

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
  height: 48px;
  padding: 0 24px;
  background: #ffffff;
  border-bottom: 1px solid $slate-200;
  font-family: $font-sans;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-btn {
  width: 28px;
  height: 28px;
  background: transparent;
  border: 1px solid $slate-200;
  border-radius: 8px;
  color: $slate-400;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;

  &:hover {
    background: $slate-50;
    color: $slate-600;
    border-color: $slate-300;
  }
}

.header-title {
  h1 {
    font-size: 18px;
    font-weight: 700;
    color: $slate-800;
    margin: 0;
    letter-spacing: -0.01em;
  }

  p {
    font-size: 12px;
    color: $slate-500;
    margin: 2px 0 0;
  }
}

.content-panel {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

// 响应式适配
@media (max-width: 768px) {
  .page-header {
    padding: 0 16px;
  }
}
</style>
