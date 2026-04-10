<template>
  <section class="module-page">
    <div class="desktop-shell">
      <header class="module-header">
        <div>
          <p class="module-caption">
            Measurement Workbench
          </p>
          <h1>计量模块</h1>
          <p class="module-description">
            本模块承接 1605MeassureApp 的作业入口，先做设备选择与状态检查，再进入采集准备。
          </p>
        </div>

        <nav class="module-switch">
          <RouterLink
            class="switch-btn"
            :to="{ name: 'module-device-management' }"
          >
            设备管理
          </RouterLink>
          <RouterLink
            class="switch-btn active"
            :to="{ name: 'module-measurement' }"
          >
            计量模块
          </RouterLink>
          <RouterLink
            class="switch-btn"
            :to="{ name: 'module-calibration' }"
          >
            进入标定模块
          </RouterLink>
          <RouterLink
            class="switch-btn switch-btn-ghost"
            :to="{ name: 'module-hub' }"
          >
            返回模块入口
          </RouterLink>
        </nav>
      </header>

      <main class="workspace-grid">
        <DeviceSelectionPanel
          module-key="measurement"
          class="grid-device"
        />

        <section class="module-card grid-info">
          <h2>计量准备看板</h2>
          <ul>
            <li>确认计量设备与打压设备均已连接。</li>
            <li>确认单位一致性检查通过后再发起会话。</li>
            <li>如需执行流程动作，请切换到标定模块。</li>
          </ul>
        </section>

        <RealtimeDataPanel class="grid-data" />
      </main>
    </div>
  </section>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'

import DeviceSelectionPanel from '@/components/DeviceSelectionPanel.vue'
import RealtimeDataPanel from '@/components/RealtimeDataPanel.vue'
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

/* 间距系统: xs(8) sm(12) md(16) lg(24) xl(32) */
.workspace-grid {
  display: grid;
  gap: clamp(16px, 2vw, 24px);
  grid-template-columns: 320px 1fr;
  grid-template-rows: auto 1fr;
  grid-template-areas:
    "device info"
    "data data";
  min-height: calc(100vh - 200px);
}

.grid-device {
  grid-area: device;
}

.grid-info {
  grid-area: info;
}

.grid-data {
  grid-area: data;
}

.module-card {
  background: #ffffff;
  border: 1px solid #d6e0ea;
  border-radius: 12px;
  box-shadow: 0 10px 28px rgb(15 23 42 / 7%);
  padding: 20px;
}

.module-card h2 {
  color: #0f172a;
  margin: 0 0 12px;
}

.module-card ul {
  color: #334155;
  margin: 0;
  padding-left: 20px;
}

.module-card li {
  margin-bottom: 10px;
  line-height: 1.5;
}

.module-card li:last-child {
  margin-bottom: 0;
}

@media (max-width: 1200px) {
  .workspace-grid {
    grid-template-columns: 280px 1fr;
  }
}

@media (max-width: 900px) {
  .workspace-grid {
    grid-template-columns: 1fr;
    grid-template-areas:
      "device"
      "info"
      "data";
    min-height: auto;
  }

  .module-header {
    flex-direction: column;
  }
}
</style>
