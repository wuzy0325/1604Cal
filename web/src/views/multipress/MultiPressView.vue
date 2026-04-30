<template>
  <PageLayout>
    <!-- 页面头部 -->
    <header class="page-header">
      <div class="header-left">
        <button class="back-btn" @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
        </button>
        <div class="header-title">
          <h1>多设备打压控制</h1>
          <p>并发控制多台打压设备</p>
        </div>
      </div>
      <div class="header-actions">
        <el-button
          type="danger"
          size="small"
          :disabled="store.registeredCount === 0"
          @click="handleStopAll"
        >
          全部停止
        </el-button>
      </div>
    </header>

    <!-- 统计栏 -->
    <div class="stats-bar">
      <StatCard label="注册设备" :value="store.registeredCount" />
      <StatCard label="打压中" :value="store.pressurizingCount" color="#ffd700" />
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
    </section>
  </PageLayout>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useMultiPressStore } from '@/stores/multipress'
import PageLayout from '@/components/common/PageLayout.vue'
import StatCard from '@/components/common/StatCard.vue'
import PressureControlCard from './PressureControlCard.vue'

const router = useRouter()
const store = useMultiPressStore()

function goBack(): void {
  router.push('/')
}

async function handleRegister(deviceId: string): Promise<void> {
  await store.registerDevice(deviceId)
}

async function handleSetPressure(deviceId: string, pressure: number): Promise<void> {
  await store.setPressure(deviceId, pressure)
}

async function handleStop(deviceId: string): Promise<void> {
  await store.stopDevice(deviceId)
}

async function handleExhaust(deviceId: string): Promise<void> {
  await store.exhaustDevice(deviceId)
}

async function handleUnregister(deviceId: string): Promise<void> {
  await store.unregisterDevice(deviceId)
}

async function handleSetUnit(deviceId: string, unit: string): Promise<void> {
  try {
    await store.setUnit(deviceId, unit)
    ElMessage.success(`打压单位已切换为 ${unit}`)
  } catch {
    ElMessage.error('设置打压单位失败')
  }
}

async function handleStopAll(): Promise<void> {
  await store.stopAll()
}

onMounted(() => {
  store.setupListeners()
})

onUnmounted(() => {
  store.cleanup()
})
</script>

<style scoped lang="scss">
/* ── 设计系统令牌 ── */
$font-sans: 'Inter', 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
$font-mono: 'JetBrains Mono', 'Fira Code', 'SF Mono', Consolas, monospace;
$mint: #10b981;
$slate-50: #f9fafb;
$slate-100: #f3f4f6;
$slate-200: #e5e7eb;
$slate-300: #d1d5db;
$slate-400: #9ca3af;
$slate-500: #6b7280;
$slate-600: #4b5563;
$slate-700: #374151;
$slate-800: #1f2937;

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

.header-actions {
  display: flex;
  gap: 10px;
}

.stats-bar {
  display: flex;
  gap: 12px;
  flex-shrink: 0;
  padding: 0 24px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: $slate-700;
  margin: 0 0 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid $slate-200;
  font-family: $font-sans;
  position: relative;
  padding-left: 10px;

  &::before {
    content: '';
    position: absolute;
    left: 0;
    top: 50%;
    transform: translateY(-50%);
    width: 3px;
    height: 14px;
    background: linear-gradient(180deg, $mint, #059669);
    border-radius: 2px;
  }
}

.available-section,
.registered-section {
  flex-shrink: 0;
  padding: 0 24px;
}

.available-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
}

.available-card {
  background: #ffffff;
  border: 1px solid $slate-200;
  border-radius: 10px;
  padding: 14px 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  transition: all 0.2s ease;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);

  &:hover {
    border-color: rgba($mint, 0.3);
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.07), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
  }
}

.available-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.available-name {
  font-size: 14px;
  font-weight: 500;
  color: $slate-700;
}

.available-detail {
  font-size: 11px;
  color: $slate-400;
  font-family: $font-mono;
}

.registered-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-text {
  font-size: 14px;
  color: $slate-400;
}

// 响应式适配
@media (max-width: 768px) {
  .page-header {
    padding: 0 16px;
  }

  .header-title h1 {
    font-size: 16px;
  }

  .stats-bar {
    flex-direction: column;
    padding: 0 16px;
  }

  .available-section,
  .registered-section {
    padding: 0 16px;
  }

  .available-grid,
  .registered-grid {
    grid-template-columns: 1fr;
  }
}
</style>
