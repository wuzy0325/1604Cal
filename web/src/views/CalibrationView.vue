<template>
  <section class="module-page">
    <div class="desktop-shell">
      <header class="module-header">
        <div>
          <p class="module-caption">
            Calibration Workbench
          </p>
          <h1>标定模块</h1>
          <p class="module-description">
            本模块承接 1604标定软件流程，依赖已选择设备执行会话状态机与报告输出。
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
            class="switch-btn"
            :to="{ name: 'module-measurement' }"
          >
            计量模块
          </RouterLink>
          <RouterLink
            class="switch-btn active"
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

      <main class="workspace-grid">
        <DeviceSelectionPanel
          module-key="calibration"
          class="grid-device"
        />

        <section class="module-card grid-control">
          <header class="card-header">
            <div>
              <h2>标定流程控制</h2>
              <p class="card-subtitle">
                当前会话状态：<strong>{{ sessionState }}</strong>
              </p>
            </div>
          </header>

          <div class="actions">
            <button
              type="button"
              @click="handleSessionAction('start')"
            >
              开始
            </button>
            <button
              type="button"
              @click="handleSessionAction('pause')"
            >
              暂停
            </button>
            <button
              type="button"
              @click="handleSessionAction('resume')"
            >
              继续
            </button>
            <button
              type="button"
              @click="handleSessionAction('stop')"
            >
              停止
            </button>
          </div>

          <div class="template-section">
            <h3>报告模板选择</h3>
            <div class="inputs">
              <label>
                测点数
                <input
                  v-model.number="templatePoints"
                  type="number"
                  min="2"
                  max="11"
                >
              </label>
              <label>
                模式
                <select v-model="templateMode">
                  <option value="single">单程</option>
                  <option value="return">回程</option>
                </select>
              </label>
              <button
                type="button"
                @click="handleSelectTemplate"
              >
                选择模板
              </button>
            </div>
            <p class="template-result">
              当前模板：{{ templateFilename || '未选择' }}
            </p>
          </div>

          <p
            v-if="errorMessage"
            class="error"
          >
            错误：{{ errorMessage }}
          </p>
        </section>

        <RealtimeDataPanel class="grid-data" />
      </main>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'

import DeviceSelectionPanel from '@/components/DeviceSelectionPanel.vue'
import RealtimeDataPanel from '@/components/RealtimeDataPanel.vue'
import { selectReportTemplate, triggerSessionAction } from '@/services/apiClient'

const sessionState = ref('idle')
const templatePoints = ref(5)
const templateMode = ref<'single' | 'return'>('single')
const templateFilename = ref('')
const errorMessage = ref('')

async function handleSessionAction(action: 'start' | 'pause' | 'resume' | 'stop') {
  errorMessage.value = ''
  try {
    const data = await triggerSessionAction(action)
    sessionState.value = data.state
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '会话操作失败'
  }
}

async function handleSelectTemplate() {
  errorMessage.value = ''
  try {
    const result = await selectReportTemplate(templatePoints.value, templateMode.value)
    templateFilename.value = result.filename
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '模板选择失败'
  }
}
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
    "device control"
    "data data";
  min-height: calc(100vh - 200px);
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
      "control"
      "data";
    min-height: auto;
  }
}

.grid-device {
  grid-area: device;
}

.grid-control {
  grid-area: control;
}

.grid-data {
  grid-area: data;
}

.module-card {
  background: #ffffff;
  border: 1px solid #d6e0ea;
  border-radius: 12px;
  box-shadow: 0 10px 28px rgb(15 23 42 / 7%);
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 20px;
}

.card-header h2 {
  color: #0f172a;
  margin: 0;
}

.card-subtitle {
  color: #334155;
  margin: 6px 0 0;
}

.actions {
  display: flex;
  gap: 12px;
}

.actions button {
  background: linear-gradient(180deg, #1e293b 0%, #0f172a 100%);
  border: 1px solid #0f172a;
  border-radius: 8px;
  color: #f8fafc;
  cursor: pointer;
  flex: 1;
  padding: 10px 16px;
}

.template-section {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 16px;
}

.template-section h3 {
  color: #0f172a;
  font-size: 14px;
  margin: 0 0 12px;
}

.inputs {
  align-items: flex-end;
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.inputs label {
  color: #334155;
  display: flex;
  flex-direction: column;
  font-size: 13px;
  gap: 6px;
}

.inputs input,
.inputs select {
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 8px 12px;
}

.inputs button {
  background: linear-gradient(180deg, #1e293b 0%, #0f172a 100%);
  border: 1px solid #0f172a;
  border-radius: 8px;
  color: #f8fafc;
  cursor: pointer;
  padding: 8px 16px;
}

.template-result {
  color: #0f172a;
  margin: 10px 0 0;
}

.error {
  color: #b91c1c;
  margin-top: 10px;
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

  .actions {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
