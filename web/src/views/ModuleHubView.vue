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
}

const featureCards: FeatureCard[] = [
  {
    title: '设备管理',
    description: '集中维护计量设备与打压设备，单位一致性检查与连接异常追踪。',
    icon: Tools,
    path: '/device-management',
    tag: '统一台账',
    color: '#3b82f6'
  },
  {
    title: '标定工作台',
    description: '会话状态机驱动的流程控制，自动/手动动作与过程状态可视化。',
    icon: SetUp,
    path: '/calibration',
    tag: '流程控制',
    color: '#f59e0b'
  },
  {
    title: '计量工作台',
    description: '数据采集与压力检测，适合作业准备、状态巡检与异常定位。',
    icon: DataLine,
    path: '/measurement',
    tag: '数据采集',
    color: '#10b981'
  },
  {
    title: '多设备打压控制',
    description: '同时控制多台打压设备，实时压力监控与稳定检测。',
    icon: Odometer,
    path: '/multi-pressure',
    tag: '并发控制',
    color: '#8b5cf6'
  }
]

function navigateTo(path: string): void {
  router.push(path)
}

function colorToBg(color: string): string {
  const r = parseInt(color.slice(1,3), 16)
  const g = parseInt(color.slice(3,5), 16)
  const b = parseInt(color.slice(5,7), 16)
  return `rgba(${r},${g},${b},0.1)`
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

      <!-- 功能卡片网格 -->
      <section class="hub-grid-wrap">
        <div class="hub-grid">
          <div
            v-for="(card, i) in featureCards"
            :key="card.path"
            class="feature-card"
            :style="{ '--accent': card.color, '--i': i }"
            @click="navigateTo(card.path)"
          >
            <div class="card-accent-line" />

            <div class="card-top">
              <div class="card-icon" :style="{ color: card.color, background: colorToBg(card.color) }">
                <el-icon>
                  <component :is="card.icon" />
                </el-icon>
              </div>
              <span class="card-tag">{{ card.tag }}</span>
            </div>

            <div class="card-body">
              <h3>{{ card.title }}</h3>
              <p>{{ card.description }}</p>
            </div>

            <div class="card-foot">
              <span class="card-action">进入模块</span>
              <el-icon><ArrowRight /></el-icon>
            </div>
          </div>
        </div>
      </section>
    </div>
  </PageLayout>
</template>

<style scoped lang="scss">
$font-sans: 'Inter', 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
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
  padding: 24px 32px 20px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 16px;
}

.brand-icon {
  width: 48px;
  height: 48px;
  color: $mint;
  background: rgba(16, 185, 129, 0.08);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;

  svg {
    width: 26px;
    height: 26px;
  }
}

.brand-text {
  h1 {
    font-size: 22px;
    font-weight: 700;
    color: $slate-800;
    margin: 0;
    letter-spacing: -0.01em;
    font-family: $font-sans;
  }
}

.brand-sub {
  font-size: 12px;
  color: $slate-400;
  font-weight: 500;
  letter-spacing: 0.05em;
  font-family: $font-sans;
}

.version-tag {
  font-size: 12px;
  color: $slate-400;
  padding: 4px 12px;
  background: $slate-100;
  border-radius: 999px;
  font-family: $font-mono;
  font-weight: 500;
  border: 1px solid rgba(255, 255, 255, 0.06);
}

/* ════════════════════════════════════════
   卡片网格
   ════════════════════════════════════════ */
.hub-grid-wrap {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 24px 32px 32px;
  overflow-y: auto;
  position: relative;

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

.hub-grid {
  flex: 1;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
  align-content: center;
  position: relative;
  z-index: 1;
}

.feature-card {
  background: #ffffff;
  border-radius: 12px;
  padding: 28px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  animation: card-rise 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) both;
  animation-delay: calc(var(--i) * 80ms + 50ms);

  &:hover {
    transform: translateY(-4px);
    box-shadow:
      0 20px 25px -5px rgba(0, 0, 0, 0.08),
      0 8px 10px -6px rgba(0, 0, 0, 0.05),
      0 0 0 1px var(--accent);

    .card-action {
      color: var(--accent);
      gap: 10px;
    }
  }
}

@keyframes card-rise {
  from {
    opacity: 0;
    transform: translateY(20px) scale(0.97);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.card-accent-line {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: var(--accent);
}

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.card-icon {
  width: 42px;
  height: 42px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);

  .feature-card:hover & {
    transform: scale(1.1);
  }
}

.card-tag {
  font-size: 11px;
  color: $slate-400;
  padding: 3px 10px;
  background: $slate-50;
  border-radius: 999px;
  font-weight: 500;
  border: 1px solid $slate-200;
  font-family: $font-sans;
}

.card-body {
  flex: 1;

  h3 {
    font-size: 17px;
    font-weight: 600;
    color: $slate-800;
    margin: 0 0 8px;
    font-family: $font-sans;
  }

  p {
    font-size: 13px;
    color: $slate-500;
    line-height: 1.6;
    margin: 0;
    font-family: $font-sans;
  }
}

.card-foot {
  display: flex;
  align-items: center;
  gap: 6px;
  font-family: $font-sans;
}

.card-action {
  font-size: 13px;
  font-weight: 500;
  color: $slate-400;
  transition: all 0.3s ease;
}

.card-foot .el-icon {
  font-size: 15px;
  color: $slate-300;
  transition: all 0.3s ease;
}

.feature-card:hover .card-foot .el-icon {
  color: var(--accent);
}

/* ════════════════════════════════════════
   Color-to-bg helper
   ════════════════════════════════════════ */
@function color-to-bg-hex($color) {
  @return $color;
}

/* 响应式 */
@media (max-width: 860px) {
  .identity-content {
    padding: 20px 24px 16px;
  }

  .hub-grid-wrap {
    padding: 20px 24px;
  }

  .hub-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .feature-card {
    padding: 22px;
  }
}
</style>
