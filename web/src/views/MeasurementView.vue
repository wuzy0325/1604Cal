<template>
  <section class="module-page">
    <div class="desktop-shell">
      <header class="module-header">
        <div>
          <p class="module-caption">
            Measurement Workbench
          </p>
          <h1>计量工作台</h1>
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

      <main class="workspace-grid">
        <DeviceSelectionPanel
          module-key="measurement"
          class="grid-device"
        />

        <section class="module-card grid-info">
          <div class="card-header">
            <el-icon class="card-icon">
              <InfoFilled />
            </el-icon>
            <h2>计量准备看板</h2>
          </div>
          <ul class="checklist">
            <li>
              <el-icon class="check-icon">
                <CircleCheck />
              </el-icon>
              确认计量设备与打压设备均已连接。
            </li>
            <li>
              <el-icon class="check-icon">
                <CircleCheck />
              </el-icon>
              确认单位一致性检查通过后再发起会话。
            </li>
            <li>
              <el-icon class="check-icon">
                <CircleCheck />
              </el-icon>
              如需执行流程动作，请切换到标定模块。
            </li>
          </ul>
        </section>

        <RealtimeDataPanel class="grid-data" />
      </main>
    </div>
  </section>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { ArrowLeft, InfoFilled, CircleCheck } from '@element-plus/icons-vue'

import DeviceSelectionPanel from '@/components/DeviceSelectionPanel.vue'
import RealtimeDataPanel from '@/components/RealtimeDataPanel.vue'
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

.workspace-grid {
  display: grid;
  gap: var(--spacing-lg);
  grid-template-columns: 360px 1fr;
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
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
}

.card-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-md);
}

.card-icon {
  font-size: 20px;
  color: var(--accent-primary);
}

.module-card h2 {
  color: var(--text-primary);
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.checklist {
  list-style: none;
  padding: 0;
  margin: 0;
}

.checklist li {
  color: var(--text-secondary);
  margin-bottom: var(--spacing-sm);
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  font-size: 14px;
  line-height: 1.5;
}

.check-icon {
  color: var(--status-success);
  font-size: 16px;
  flex-shrink: 0;
}

@media (max-width: 1200px) {
  .workspace-grid {
    grid-template-columns: 320px 1fr;
  }
}

@media (max-width: 900px) {
  .module-page {
    padding: var(--spacing-md);
  }
  
  .module-header {
    flex-direction: column;
  }
  
  .workspace-grid {
    grid-template-columns: 1fr;
    grid-template-areas:
      "device"
      "info"
      "data";
    min-height: auto;
  }
}
</style>