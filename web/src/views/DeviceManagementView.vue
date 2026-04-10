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
            返回模块入口
          </RouterLink>
        </nav>
      </header>

      <DeviceManagementPanel />
    </div>
  </section>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'

import DeviceManagementPanel from '@/components/DeviceManagementPanel.vue'
</script>

<style scoped>
.module-page {
  background:
    radial-gradient(circle at 15% 10%, rgb(148 163 184 / 16%), transparent 35%),
    linear-gradient(160deg, #dbe3ea 0%, #eef2f6 46%, #dbe3ea 100%);
  min-height: 100vh;
  padding: 16px;
}

.desktop-shell {
  background: #f8fafc;
  border: 1px solid #b8c6d3;
  border-radius: 16px;
  box-shadow: 0 16px 42px rgb(15 23 42 / 14%);
  margin: 0 auto;
  max-width: 1728px;
  min-height: min(972px, calc(100vh - 32px));
  padding: 18px;
  width: min(1728px, calc(100vw - 32px));
}

.module-header {
  align-items: flex-start;
  border-bottom: 1px solid #d6e0ea;
  display: flex;
  gap: 12px;
  justify-content: space-between;
  margin-bottom: 14px;
  padding-bottom: 14px;
}

.module-caption {
  color: #475569;
  font-size: 12px;
  letter-spacing: 0.08em;
  margin: 0;
  text-transform: uppercase;
}

.module-header h1 {
  color: #0f172a;
  margin: 6px 0;
}

.module-description {
  color: #334155;
  margin: 0;
  max-width: 760px;
}

.module-switch {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.switch-btn {
  background: #e2e8f0;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  color: #0f172a;
  padding: 7px 11px;
  text-decoration: none;
}

.switch-btn.active {
  background: #0f172a;
  border-color: #0f172a;
  color: #f8fafc;
}

.switch-btn-ghost {
  background: #f8fafc;
}

@media (max-width: 1380px) {
  .desktop-shell {
    min-height: auto;
  }
}

@media (max-width: 900px) {
  .module-header {
    flex-direction: column;
  }
}
</style>
