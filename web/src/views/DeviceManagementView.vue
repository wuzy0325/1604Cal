<template>
  <section class="module-page">
    <div class="desktop-shell">
      <header class="module-header">
        <div>
          <p class="module-caption">
            Unified Device Registry
          </p>
          <h1>设备管理模块</h1>
          <p class="module-description">
            统一维护计量与打压设备台账，所有业务模块共享同一套设备状态与连接信息。
          </p>
        </div>

        <nav class="module-switch">
          <RouterLink
            class="switch-btn active"
            :to="{ name: 'module-device-management' }"
          >
            设备管理
          </RouterLink>
          <RouterLink
            class="switch-btn"
            :to="{ name: 'module-measurement' }"
          >
            计量模块
          </RouterLink>
          <RouterLink
            class="switch-btn"
            :to="{ name: 'module-calibration' }"
          >
            标定模块
          </RouterLink>
          <RouterLink
            class="switch-btn switch-btn-ghost"
            :to="{ name: 'module-hub' }"
          >
            <el-icon><ArrowLeft /></el-icon>
            返回
          </RouterLink>
        </nav>
      </header>

      <DeviceManagementPanel />
    </div>
  </section>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'

import DeviceManagementPanel from '@/components/DeviceManagementPanel.vue'
</script>

<style scoped lang="scss">
.module-page {
  padding: var(--spacing-xl);
  min-height: 100vh;
}

.desktop-shell {
  max-width: 1600px;
  margin: 0 auto;
}

.module-header {
  align-items: flex-start;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  gap: var(--spacing-lg);
  justify-content: space-between;
  margin-bottom: var(--spacing-lg);
  padding-bottom: var(--spacing-lg);
}

.module-caption {
  color: var(--accent-primary);
  font-size: 12px;
  letter-spacing: 0.08em;
  margin: 0 0 var(--spacing-xs);
  text-transform: uppercase;
  font-weight: 600;
}

.module-header h1 {
  color: var(--text-primary);
  margin: 0 0 var(--spacing-sm);
  font-size: 24px;
  font-weight: 600;
}

.module-description {
  color: var(--text-secondary);
  margin: 0;
  max-width: 760px;
  font-size: 14px;
  line-height: 1.5;
}

.module-switch {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
}

.switch-btn {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  padding: var(--spacing-sm) var(--spacing-md);
  text-decoration: none;
  font-size: 14px;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  
  &:hover {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }
  
  .el-icon {
    font-size: 14px;
  }
}

.switch-btn.active {
  background: var(--accent-primary);
  border-color: var(--accent-primary);
  color: var(--text-primary);
}

.switch-btn-ghost {
  background: transparent;
}

@media (max-width: 900px) {
  .module-page {
    padding: var(--spacing-md);
  }
  
  .module-header {
    flex-direction: column;
  }
  
  .module-switch {
    width: 100%;
  }
}
</style>