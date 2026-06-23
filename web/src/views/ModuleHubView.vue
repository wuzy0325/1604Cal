<script setup lang="ts">
import { useRouter } from 'vue-router'
import { ArrowRight, Tools, DataLine, SetUp, Odometer } from '@element-plus/icons-vue'
import PageLayout from '@/components/common/PageLayout.vue'

const router = useRouter()

interface FeatureCard {
  title: string
  description: string
  icon: unknown
  path: string
  tag: string
  color: string
  variant: 'primary' | 'secondary'
}

const featureCards: FeatureCard[] = [
  {
    title: '标定工作台',
    description: '会话状态机驱动的流程控制，自动/手动动作与过程状态可视化。',
    icon: SetUp,
    path: '/calibration',
    tag: '流程控制',
    color: '#f59e0b',
    variant: 'primary'
  },
  {
    title: '计量工作台',
    description: '数据采集与压力检测，适合作业准备、状态巡检与异常定位。',
    icon: DataLine,
    path: '/measurement',
    tag: '数据采集',
    color: '#10b981',
    variant: 'primary'
  },
  {
    title: '设备管理',
    description: '集中维护计量设备与打压设备，单位一致性检查与连接异常追踪。',
    icon: Tools,
    path: '/device-management',
    tag: '统一台账',
    color: '#3b82f6',
    variant: 'secondary'
  },
  {
    title: '多设备打压',
    description: '同时控制多台打压设备，实时压力监控与稳定检测。',
    icon: Odometer,
    path: '/multi-pressure',
    tag: '并发控制',
    color: '#8b5cf6',
    variant: 'secondary'
  }
]

function navigateTo(path: string): void {
  router.push(path)
}

function colorToBg(color: string): string {
  const r = parseInt(color.slice(1, 3), 16)
  const g = parseInt(color.slice(3, 5), 16)
  const b = parseInt(color.slice(5, 7), 16)
  return `rgba(${r},${g},${b},0.08)`
}
</script>

<template>
  <PageLayout>
    <div class="hub">
      <!-- 顶部身份区域 -->
      <header class="hub-identity">
        <div class="identity-content">
          <div class="brand">
            <div class="brand-icon">
              <svg viewBox="0 0 32 32" fill="none">
                <path d="M16 3L3 10L16 17L29 10L16 3Z" fill="currentColor" fill-opacity="0.15"/>
                <path d="M3 22L16 29L29 22" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M3 16L16 23L29 16" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <circle cx="16" cy="10" r="3" fill="currentColor" fill-opacity="0.3"/>
              </svg>
            </div>
            <div class="brand-text">
              <h1>Cal1604</h1>
              <span class="brand-sub">工业压力标定系统</span>
            </div>
          </div>
          <span class="version-tag">v2.0.0</span>
        </div>
      </header>

      <!-- Bento 布局 -->
      <section class="hub-content">
        <div class="bento-grid">
          <template v-for="(card, i) in featureCards" :key="card.path">
            <!-- 主入口：标定/计量 -->
            <button
              v-if="card.variant === 'primary'"
              class="entry-primary"
              :style="{ '--accent': card.color, '--i': i }"
              @click="navigateTo(card.path)"
            >
              <div class="primary-icon" :style="{ color: card.color, background: colorToBg(card.color) }">
                <el-icon><component :is="card.icon" /></el-icon>
              </div>
              <div class="primary-body">
                <span class="primary-tag">{{ card.tag }}</span>
                <h2>{{ card.title }}</h2>
                <p>{{ card.description }}</p>
              </div>
              <span class="primary-arrow" aria-hidden="true">
                <el-icon><ArrowRight /></el-icon>
              </span>
            </button>

            <!-- 辅助入口：设备管理/多设备打压 -->
            <button
              v-else
              class="entry-secondary"
              :style="{ '--accent': card.color, '--i': i }"
              @click="navigateTo(card.path)"
            >
              <div class="secondary-icon" :style="{ color: card.color, background: colorToBg(card.color) }">
                <el-icon><component :is="card.icon" /></el-icon>
              </div>
              <div class="secondary-body">
                <h3>{{ card.title }}</h3>
                <span class="secondary-tag">{{ card.tag }}</span>
              </div>
              <span class="secondary-arrow" aria-hidden="true">
                <el-icon><ArrowRight /></el-icon>
              </span>
            </button>
          </template>
        </div>
      </section>
    </div>
  </PageLayout>
</template>

<style scoped lang="scss">
$font-sans: 'DM Sans', 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
$font-mono: 'JetBrains Mono', 'Fira Code', 'SF Mono', Consolas, monospace;
$slate-50: #f9fafb;
$slate-100: #f3f4f6;
$slate-200: #e5e7eb;
$slate-300: #d1d5db;
$slate-400: #9ca3af;
$slate-500: #6b7280;
$slate-600: #4b5563;
$slate-700: #374151;
$slate-800: #1f2937;
$slate-900: #111827;
$mint: #10b981;

/* ════════════════════════════════════════
   Hub 容器
   ════════════════════════════════════════ */
.hub {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* ════════════════════════════════════════
   顶部身份区域
   ════════════════════════════════════════ */
.hub-identity {
  flex-shrink: 0;
  background: $slate-50;
  border-bottom: 1px solid $slate-200;
}

.identity-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 32px 16px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 14px;
}

.brand-icon {
  width: 40px;
  height: 40px;
  color: $mint;
  background: rgba(16, 185, 129, 0.08);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;

  svg { width: 22px; height: 22px; }
}

.brand-text {
  h1 {
    font-size: 18px;
    font-weight: 700;
    color: $slate-800;
    margin: 0;
    letter-spacing: -0.01em;
    font-family: $font-sans;
  }
}

.brand-sub {
  font-size: 11px;
  color: $slate-400;
  font-weight: 500;
  letter-spacing: 0.05em;
  font-family: $font-sans;
}

.version-tag {
  font-size: 11px;
  color: $slate-400;
  padding: 3px 10px;
  background: $slate-100;
  border-radius: 999px;
  font-family: $font-mono;
  font-weight: 500;
}

/* ════════════════════════════════════════
   Bento 布局
   ════════════════════════════════════════ */
.hub-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 24px 32px 32px;
  overflow-y: auto;
  position: relative;

  /* 坐标纸背景 */
  &::before {
    content: '';
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(rgba(16, 185, 129, 0.035) 1px, transparent 1px),
      linear-gradient(90deg, rgba(16, 185, 129, 0.035) 1px, transparent 1px);
    background-size: 24px 24px;
    pointer-events: none;
  }
}

.bento-grid {
  flex: 1;
  display: grid;
  /* 两行布局：上行两个主入口各占1列，下行两个辅助入口各占1列 */
  grid-template-columns: 1fr 1fr;
  grid-template-rows: auto auto;
  gap: 16px;
  align-content: center;
  position: relative;
  z-index: 1;
  max-width: 800px;
  width: 100%;
  margin: 0 auto;
}

/* ── 主入口：标定/计量 ── */
.entry-primary {
  background: #ffffff;
  border: 1px solid $slate-200;
  border-radius: 12px;
  padding: 28px 24px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 18px;
  text-align: left;
  font-family: $font-sans;
  position: relative;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  transition: border-color 0.2s ease, box-shadow 0.25s ease;
  animation: card-rise 0.4s cubic-bezier(0.25, 1, 0.5, 1) both;
  animation-delay: calc(var(--i) * 80ms + 50ms);

  &:hover {
    border-color: var(--accent);
    box-shadow: 0 8px 16px -4px rgba(0, 0, 0, 0.08), 0 0 0 1px var(--accent);

    .primary-arrow { color: var(--accent); transform: translateX(3px); }
  }

  &:active { transform: scale(0.995); }
}

.primary-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  flex-shrink: 0;
}

.primary-body {
  flex: 1;
  min-width: 0;
}

.primary-tag {
  font-size: 11px;
  font-weight: 600;
  color: var(--accent);
  letter-spacing: 0.04em;
  text-transform: uppercase;
  display: block;
  margin-bottom: 6px;
}

.primary-body h2 {
  font-size: 20px;
  font-weight: 700;
  color: $slate-800;
  margin: 0 0 6px;
  line-height: 1.2;
}

.primary-body p {
  font-size: 13px;
  color: $slate-500;
  line-height: 1.5;
  margin: 0;
}

.primary-arrow {
  color: $slate-300;
  font-size: 18px;
  flex-shrink: 0;
  margin-top: 4px;
  transition: color 0.2s ease, transform 0.2s ease;
}

/* ── 辅助入口：设备管理/多设备打压 ── */
.entry-secondary {
  background: $slate-50;
  border: 1px solid $slate-200;
  border-radius: 10px;
  padding: 14px 18px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 12px;
  text-align: left;
  font-family: $font-sans;
  transition: border-color 0.2s ease, background 0.2s ease;
  animation: card-rise 0.4s cubic-bezier(0.25, 1, 0.5, 1) both;
  animation-delay: calc(var(--i) * 80ms + 50ms);

  &:hover {
    background: #ffffff;
    border-color: var(--accent);

    .secondary-arrow { color: var(--accent); transform: translateX(2px); }
  }

  &:active { transform: scale(0.995); }
}

.secondary-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  flex-shrink: 0;
}

.secondary-body {
  flex: 1;
  min-width: 0;
}

.secondary-body h3 {
  font-size: 14px;
  font-weight: 600;
  color: $slate-800;
  margin: 0;
  line-height: 1.3;
}

.secondary-tag {
  font-size: 11px;
  color: $slate-400;
  font-weight: 500;
}

.secondary-arrow {
  color: $slate-300;
  font-size: 14px;
  flex-shrink: 0;
  transition: color 0.2s ease, transform 0.2s ease;
}

/* ── 入场动画 ── */
@keyframes card-rise {
  from {
    opacity: 0;
    transform: translateY(16px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* ════════════════════════════════════════
   响应式
   ════════════════════════════════════════ */
@media (max-width: 860px) {
  .identity-content { padding: 16px 20px 12px; }
  .hub-content { padding: 20px; }

  .bento-grid {
    grid-template-columns: 1fr;
    grid-template-rows: auto;
    max-width: none;
  }

  .entry-primary { padding: 20px 18px; }
  .primary-icon { width: 40px; height: 40px; font-size: 18px; }
  .primary-body h2 { font-size: 17px; }
}
</style>
