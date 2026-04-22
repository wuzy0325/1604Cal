<template>
  <section class="multi-press-page">
    <div class="desktop-shell">
      <header class="module-header">
        <div>
          <p class="module-caption">
            Multi-Pressure Control
          </p>
          <h1>多设备打压控制</h1>
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
            class="switch-btn"
            :to="{ name: 'module-calibration' }"
          >
            标定模块
          </RouterLink>
          <RouterLink
            class="switch-btn active"
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

      <div class="stats-bar">
        <StatCard label="注册设备" :value="store.registeredCount" />
        <StatCard label="打压中" :value="store.pressurizingCount" color="#ffd700" />
        <el-button
          type="danger"
          size="small"
          :disabled="store.registeredCount === 0"
          @click="handleStopAll"
        >
          全部停止
        </el-button>
      </div>

      <!-- 未注册打压设备 -->
      <section v-if="store.availableDevices.length > 0" class="available-section">
        <h3 class="section-title">可用打压设备</h3>
        <div class="available-grid">
          <div
            v-for="dev in store.availableDevices"
            :key="dev.id"
            class="available-card"
          >
            <div class="available-info">
              <span class="available-name">{{ dev.name }}</span>
              <span class="available-detail">{{ dev.host }}:{{ dev.port }}</span>
            </div>
            <el-button size="small" type="primary" @click="handleRegister(dev.id)">
              注册
            </el-button>
          </div>
        </div>
      </section>

      <!-- 已注册设备卡片网格 -->
      <section v-if="store.registeredDevices.length > 0" class="registered-section">
        <h3 class="section-title">已注册设备</h3>
        <div class="registered-grid">
          <PressureControlCard
            v-for="devState in store.registeredDevices"
            :key="devState.deviceId"
            :state="devState"
            :metadata="store.getMeta(devState.deviceId)"
            @set-pressure="handleSetPressure"
            @stop="handleStop"
            @exhaust="handleExhaust"
            @unregister="handleUnregister"
            @set-unit="handleSetUnit"
          />
        </div>
      </section>

      <!-- 空状态 -->
      <section v-if="store.registeredDevices.length === 0 && store.availableDevices.length === 0" class="empty-state">
        <p class="empty-text">暂无打压设备</p>
        <p class="empty-hint">请先在设备管理模块中添加打压类型设备</p>
        <el-button size="small" @click="$router.push({ name: 'module-device-management' })">
          前往设备管理
        </el-button>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { RouterLink } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import { createEventStream } from "@/api/client"
import type { StreamEventPayload } from "@/types/api"
import { useMultiPressStore } from '@/stores/multipress'
import StatCard from '@/components/common/StatCard.vue'
import PressureControlCard from './PressureControlCard.vue'

const store = useMultiPressStore()
let eventSource: EventSource | null = null

onMounted(async () => {
  await store.loadPressureDevices()
  await store.refreshDeviceStates()
  store.startPolling()

  eventSource = createEventStream((payload: StreamEventPayload) => {
    store.handleSSEEvent(payload)
  })
})

onUnmounted(() => {
  store.stopPolling()
  if (eventSource) {
    eventSource.close()
    eventSource = null
  }
})

async function handleRegister(deviceId: string) {
  await store.registerDevice(deviceId)
}

async function handleUnregister(deviceId: string) {
  await store.unregisterDevice(deviceId)
}

async function handleSetPressure(deviceId: string, target: number) {
  await store.setPressure(deviceId, target)
}

async function handleStop(deviceId: string) {
  await store.stopDevice(deviceId)
}

async function handleExhaust(deviceId: string) {
  await store.exhaustDevice(deviceId)
}

async function handleSetUnit(deviceId: string, unit: string) {
  await store.setUnit(deviceId, unit)
}

async function handleStopAll() {
  await store.stopAll()
}
</script>

<style scoped lang="scss">
.multi-press-page {
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
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  padding: 6px 14px;
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

.stats-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-md);
  flex-shrink: 0;
}

.section-title {
  color: var(--text-secondary);
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 1px;
  margin: 0 0 var(--spacing-sm);
  font-weight: 600;
}

/* 可用设备区 */
.available-section {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: var(--spacing-md);
  flex-shrink: 0;
}

.available-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--spacing-sm);
}

.available-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-sm) var(--spacing-md);
  background: var(--bg-tertiary);
  border-radius: 3px;
  border: 1px solid var(--border-color);
}

.available-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.available-name {
  color: var(--text-primary);
  font-size: 13px;
  font-weight: 600;
}

.available-detail {
  color: var(--text-secondary);
  font-size: 11px;
}

/* 已注册设备区 */
.registered-section {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

.registered-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: var(--spacing-md);
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 0;
  gap: var(--spacing-sm);
  flex: 1;
  min-height: 0;
}

.empty-text {
  color: var(--text-secondary);
  font-size: 16px;
  margin: 0;
}

.empty-hint {
  color: var(--text-muted);
  font-size: 13px;
  margin: 0;
}

@media (max-width: 1200px) {
  .registered-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 900px) {
  .multi-press-page {
    padding: var(--spacing-md);
    overflow: auto;
  }

  .module-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-md);
  }

  .stats-bar {
    flex-direction: column;
    align-items: flex-start;
  }

  .registered-grid {
    grid-template-columns: 1fr;
  }

  .available-grid {
    grid-template-columns: 1fr;
  }

  .registered-section {
    overflow-y: visible;
  }
}
</style>
