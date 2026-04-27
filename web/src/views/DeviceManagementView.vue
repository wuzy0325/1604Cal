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
      <div class="header-actions">
        <button
          class="action-btn refresh"
          :class="{ loading: refreshing }"
          @click="handleRefresh"
        >
          <el-icon><Refresh /></el-icon>
          <span>刷新</span>
        </button>
      </div>
    </header>

    <!-- 设备管理面板 -->
    <DeviceManagementPanel class="content-panel" />
  </PageLayout>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { Refresh, ArrowLeft } from '@element-plus/icons-vue'
import { useDeviceStore } from '@/stores/deviceStore'
import PageLayout from '@/components/common/PageLayout.vue'
import DeviceManagementPanel from '@/components/device/DeviceManagementPanel.vue'

const router = useRouter()
const deviceStore = useDeviceStore()

/** 刷新加载状态 */
const refreshing = ref(false)

/**
 * 返回首页
 */
function goBack(): void {
  router.push('/')
}

/**
 * 刷新设备列表
 */
async function handleRefresh(): Promise<void> {
  refreshing.value = true
  try {
    await deviceStore.loadDevices()
    await deviceStore.checkUnitConsistency()
  } finally {
    refreshing.value = false
  }
}

// 生命周期：挂载时初始化
onMounted(async () => {
  deviceStore.setupListeners()
  await deviceStore.loadDevices()
  await deviceStore.checkUnitConsistency()
})

// 生命周期：卸载时清理
onUnmounted(() => {
  deviceStore.cleanup()
})
</script>

<style scoped lang="scss">

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  flex-shrink: 0;
  padding-bottom: $spacing-4;
  border-bottom: 1px solid $border-color-light;
}

.header-left {
  display: flex;
  align-items: center;
  gap: $spacing-4;
}

.back-btn {
  width: 40px;
  height: 40px;
  background: rgba($neutral-800, 0.6);
  border: 1px solid $border-color;
  border-radius: $radius-md;
  color: $text-secondary;
  cursor: pointer;
  transition: all $transition-fast;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;

  &:hover {
    background: rgba($neutral-700, 0.8);
    color: $text-primary;
    border-color: $border-color-strong;
  }
}

.header-title {
  h1 {
    font-size: 28px;
    font-weight: $font-weight-bold;
    color: $text-primary;
    margin: 0 0 $spacing-1;
    letter-spacing: -0.02em;
  }

  p {
    font-size: $font-size-sm;
    color: $text-secondary;
    margin: 0;
  }
}

.header-actions {
  display: flex;
  gap: $spacing-3;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: $spacing-2;
  padding: $spacing-2 $spacing-4;
  background: rgba($neutral-800, 0.6);
  border: 1px solid $border-color;
  border-radius: $radius-md;
  color: $text-secondary;
  cursor: pointer;
  transition: all $transition-fast;
  font-size: $font-size-sm;
  font-weight: $font-weight-medium;

  &:hover {
    background: rgba($neutral-700, 0.8);
    color: $text-primary;
    border-color: $border-color-strong;
  }

  &.refresh.loading {
    opacity: 0.6;
    pointer-events: none;
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
    flex-direction: column;
    align-items: flex-start;
    gap: $spacing-4;
  }

  .header-left {
    width: 100%;
  }

  .header-title h1 {
    font-size: $font-size-xl;
  }
}
</style>
