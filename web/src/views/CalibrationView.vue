<template>
  <section class="module-page">
    <div class="desktop-shell">
      <header class="module-header">
        <div>
          <p class="module-caption">
            Calibration Workbench
          </p>
          <h1>标定工作台</h1>
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
            <el-icon><ArrowLeft /></el-icon>
            返回
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
            <div class="header-title">
              <el-icon class="card-icon">
                <SetUp />
              </el-icon>
              <div>
                <h2>标定流程控制</h2>
                <p class="card-subtitle">
                  当前会话状态：
                  <span :class="['status-badge', `status-${sessionState}`]">
                    {{ sessionStateText }}
                  </span>
                </p>
              </div>
            </div>
          </header>

          <div class="actions">
            <button
              type="button"
              class="btn btn-success"
              :disabled="sessionState === 'running'"
              @click="handleSessionAction('start')"
            >
              <el-icon><VideoPlay /></el-icon>
              开始
            </button>
            <button
              type="button"
              class="btn btn-warning"
              :disabled="sessionState !== 'running'"
              @click="handleSessionAction('pause')"
            >
              <el-icon><VideoPause /></el-icon>
              暂停
            </button>
            <button
              type="button"
              class="btn btn-primary"
              :disabled="sessionState !== 'paused'"
              @click="handleSessionAction('resume')"
            >
              <el-icon><RefreshRight /></el-icon>
              继续
            </button>
            <button
              type="button"
              class="btn btn-danger"
              :disabled="sessionState === 'idle'"
              @click="handleSessionAction('stop')"
            >
              <el-icon><CloseBold /></el-icon>
              停止
            </button>
          </div>

          <div class="template-section">
            <h3>
              <el-icon><Document /></el-icon>
              报告模板选择
            </h3>
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
                class="btn btn-primary"
                @click="handleSelectTemplate"
              >
                <el-icon><FolderOpened /></el-icon>
                选择模板
              </button>
            </div>            <div class="template-result">
              <el-icon><DocumentChecked /></el-icon>
              <span>当前模板：{{ templateFilename || '未选择' }}</span>
            </div>
          </div>

          <div
            v-if="errorMessage"
            class="error-message"
          >
            <el-icon><Warning /></el-icon>
            {{ errorMessage }}
          </div>
        </section>

        <RealtimeDataPanel class="grid-data" />
      </main>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { RouterLink } from 'vue-router'
import {
  ArrowLeft,
  SetUp,
  VideoPlay,
  VideoPause,
  RefreshRight,
  CloseBold,
  Document,
  DocumentChecked,
  FolderOpened,
  Warning
} from '@element-plus/icons-vue'

import DeviceSelectionPanel from '@/components/DeviceSelectionPanel.vue'
import RealtimeDataPanel from '@/components/RealtimeDataPanel.vue'
import { selectReportTemplate, triggerSessionAction } from '@/services/apiClient'

type SessionState = 'idle' | 'running' | 'paused'

const sessionState = ref<SessionState>('idle')
const templatePoints = ref(5)
const templateMode = ref<'single' | 'return'>('single')
const templateFilename = ref('')
const errorMessage = ref('')

const sessionStateText = computed(() => {
  const map: Record<SessionState, string> = {
    idle: '空闲',
    running: '运行中',
    paused: '已暂停'
  }
  return map[sessionState.value]
})

async function handleSessionAction(action: 'start' | 'pause' | 'resume' | 'stop') {
  errorMessage.value = ''
  try {
    const data = await triggerSessionAction(action)
    sessionState.value = data.state as SessionState
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
    "device control"
    "data data";
  min-height: calc(100vh - 200px);
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
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.card-header {
  margin-bottom: var(--spacing-sm);
}

.header-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.card-icon {
  font-size: 24px;
  color: var(--accent-primary);
}

.card-header h2 {
  color: var(--text-primary);
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.card-subtitle {
  color: var(--text-secondary);
  margin: var(--spacing-xs) 0 0;
  font-size: 14px;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 600;
}

.status-idle {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}

.status-running {
  background: rgba(16, 185, 129, 0.2);
  color: var(--status-success);
}

.status-paused {
  background: rgba(245, 158, 11, 0.2);
  color: var(--status-warning);
}

.actions {
  display: flex;
  gap: var(--spacing-md);
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-sm) var(--spacing-md);
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  
  .el-icon {
    font-size: 16px;
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.btn-success {
  background: var(--status-success);
  color: white;
  
  &:hover:not(:disabled) {
    background: #059669;
  }
}

.btn-warning {
  background: var(--status-warning);
  color: white;
  
  &:hover:not(:disabled) {
    background: #d97706;
  }
}

.btn-primary {
  background: var(--accent-primary);
  color: white;
  
  &:hover:not(:disabled) {
    background: var(--accent-hover);
  }
}

.btn-danger {
  background: var(--status-error);
  color: white;
  
  &:hover:not(:disabled) {
    background: #dc2626;
  }
}

.template-section {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
}

.template-section h3 {
  color: var(--text-primary);
  font-size: 14px;
  margin: 0 0 var(--spacing-md);
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  
  .el-icon {
    color: var(--accent-primary);
  }
}

.inputs {
  align-items: flex-end;
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-md);
}

.inputs label {
  color: var(--text-secondary);
  display: flex;
  flex-direction: column;
  font-size: 13px;
  gap: var(--spacing-xs);
}

.inputs input,
.inputs select {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  padding: var(--spacing-sm) var(--spacing-md);
  min-width: 100px;
}

.template-result {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  color: var(--text-secondary);
  font-size: 14px;
  
  .el-icon {
    color: var(--status-success);
  }
  
  span {
    color: var(--text-primary);
    font-weight: 500;
  }
}

.error-message {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  color: var(--status-error);
  font-size: 14px;
  padding: var(--spacing-sm);
  background: rgba(239, 68, 68, 0.1);
  border-radius: var(--radius-sm);
  
  .el-icon {
    font-size: 16px;
  }
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
      "control"
      "data";
    min-height: auto;
  }
  
  .actions {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>