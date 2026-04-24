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

async function handleSetPressure(payload: { deviceId: string; pressure: number }): Promise<void> {
  await store.setPressure(payload.deviceId, payload.pressure)
}

async function handleStop(deviceId: string): Promise<void> {
  await store.stopPressurizing(deviceId)
}

async function handleExhaust(deviceId: string): Promise<void> {
  await store.exhaust(deviceId)
}

async function handleUnregister(deviceId: string): Promise<void> {
  await store.unregisterDevice(deviceId)
}

async function handleSetUnit(payload: { deviceId: string; unit: string }): Promise<void> {
  await store.setUnit(payload.deviceId, payload.unit)
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
@import '@/styles/variables.scss';

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

.stats-bar {
  display: flex;
  gap: $spacing-4;
  flex-shrink: 0;
}

.section-title {
  font-size: $font-size-lg;
  font-weight: $font-weight-semibold;
  color: $text-primary;
  margin: 0 0 $spacing-4;
  display: flex;
  align-items: center;
  gap: $spacing-2;
  
  &::before {
    content: '';
    width: 4px;
    height: 16px;
    background: $primary-500;
    border-radius: 2px;
  }
}

.available-section,
.registered-section {
  flex-shrink: 0;
}

.available-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: $spacing-4;
}

.available-card {
  background: $bg-card;
  border: 1px solid $border-color;
  border-radius: $radius-lg;
  padding: $spacing-4;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: $spacing-4;
  transition: all $transition-fast;

  &:hover {
    border-color: rgba($primary-500, 0.3);
    box-shadow: $shadow-md;
  }
}

.available-info {
  display: flex;
  flex-direction: column;
  gap: $spacing-1;
}

.available-name {
  font-size: $font-size-base;
  font-weight: $font-weight-medium;
  color: $text-primary;
}

.available-detail {
  font-size: $font-size-xs;
  color: $text-tertiary;
  font-family: $font-family-mono;
}

.registered-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: $spacing-6;
}

.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-text {
  font-size: $font-size-lg;
  color: $text-tertiary;
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
  
  .stats-bar {
    flex-direction: column;
  }
  
  .available-grid,
  .registered-grid {
    grid-template-columns: 1fr;
  }
}
</style>
