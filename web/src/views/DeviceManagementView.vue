<template>
  <section class="module-page">
    <div class="desktop-shell">
      <header class="module-header">
        <div>
          <p class="module-caption">
            Unified Device Registry
          </p>
          <h1>设备管理模块</h1>
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
            class="switch-btn"
            :to="{ name: 'module-multi-pressure' }"
          >
            多设备打压
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

      <DeviceManagementPanel class="module-panel" />
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
  padding: var(--spacing-lg);
  box-sizing: border-box;
  height: 100%;
  overflow: hidden;
  background: var(--bg-primary);
}

.desktop-shell {
  max-width: 100%;
  height: 100%;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.module-header {
  align-items: flex-end;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  gap: var(--spacing-lg);
  justify-content: space-between;
  padding-bottom: var(--spacing-lg);
  flex-shrink: 0;
  min-height: 52px;
}

 .module-panel {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
}

.module-caption {
  color: var(--accent-primary);
  font-size: 11px;
  letter-spacing: 0.08em;
  margin: 0 0 var(--spacing-xs);
  text-transform: uppercase;
  font-weight: 600;
}

.module-header h1 {
  color: var(--text-primary);
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.module-switch {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-xs);
}

.switch-btn {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 3px;
  color: var(--text-secondary);
  padding: 5px 12px;
  text-decoration: none;
  font-size: 13px;
  transition: all 0.15s ease;
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);

  &:hover {
    background: var(--bg-quaternary);
    color: var(--text-primary);
  }

  .el-icon {
    font-size: 12px;
  }
}

.switch-btn.active {
  background: var(--accent-primary);
  border-color: var(--accent-primary);
  color: var(--bg-primary);
  font-weight: 600;
}

.switch-btn-ghost {
  background: transparent;
}

@media (max-width: 900px) {
  .module-page {
    padding: var(--spacing-md);
    overflow: auto;
  }

  .desktop-shell {
    max-width: 100%;
    height: auto;
    gap: var(--spacing-md);
  }

  .module-header {
    flex-direction: column;
  }

  .module-switch {
    width: 100%;
  }
}
</style>
