<script setup lang="ts">
/**
 * @file 模块中心视图
 * @description 应用主页面，包含导航和功能入口
 * @version 2.0.0
 */
import { useRouter } from 'vue-router'
import { ArrowRight, Tools, DataLine, SetUp, Odometer } from '@element-plus/icons-vue'
import PageLayout from '@/components/common/PageLayout.vue'

const router = useRouter()

const featureCards = [
  {
    title: '设备管理模块',
    description: '集中维护计量设备与打压设备连接状态，维护单位一致性检查与连接异常追踪',
    icon: Tools,
    path: '/device-management',
    color: 'primary',
    stats: '统一台账',
    source: '融合能力'
  },
  {
    title: '计量模块',
    description: '统一设备台账、连接状态、单位门禁，适合作业准备、状态巡检与异常定位',
    icon: DataLine,
    path: '/measurement',
    color: 'accent',
    stats: '数据采集',
    source: '1605MeassureApp'
  },
  {
    title: '标定工作台',
    description: '会话状态机驱动的流程控制，自动/手动动作与过程状态可视化',
    icon: SetUp,
    path: '/calibration',
    color: 'success',
    stats: '流程控制',
    source: '1604标定软件'
  },
  {
    title: '多设备打压控制',
    description: '同时控制多台打压设备，实时压力监控与稳定检测',
    icon: Odometer,
    path: '/multi-pressure',
    color: 'warning',
    stats: '并发控制',
    source: '独立模块'
  }
]

function navigateTo(path: string): void {
  router.push(path)
}
</script>

<template>
  <PageLayout>
    <!-- 顶部标题栏 -->
    <header class="page-header">
      <h2>Cal1604</h2>
      <span class="version">v2.0.0</span>
    </header>

    <!-- 功能卡片区域 -->
    <section class="features-section">
      <div class="features-grid">
        <div
          v-for="card in featureCards"
          :key="card.path"
          class="feature-card"
          :class="`color-${card.color}`"
          @click="navigateTo(card.path)"
        >
          <div class="card-header">
            <div class="card-icon">
              <el-icon>
                <component :is="card.icon" />
              </el-icon>
            </div>
            <span class="card-stats">{{ card.stats }}</span>
          </div>
          <div class="card-body">
            <h4>{{ card.title }}</h4>
            <p class="card-source">{{ card.source }}</p>
            <p>{{ card.description }}</p>
          </div>
          <div class="card-footer">
            <span>进入功能</span>
            <el-icon><ArrowRight /></el-icon>
          </div>
        </div>
      </div>
    </section>
  </PageLayout>
</template>

<style scoped lang="scss">
/* ── 设计系统令牌 ── */
$font-sans: 'Inter', 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
$font-mono: 'JetBrains Mono', 'Fira Code', 'SF Mono', Consolas, monospace;
$mint: #10b981;
$mint-light: #34d399;
$mint-dark: #059669;
$slate-50: #f9fafb;
$slate-100: #f3f4f6;
$slate-200: #e5e7eb;
$slate-300: #d1d5db;
$slate-400: #9ca3af;
$slate-500: #6b7280;
$slate-600: #4b5563;
$slate-700: #374151;
$slate-800: #1f2937;
$green: #22c55e;
$blue: #3b82f6;
$amber: #f59e0b;

// 页面头部
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

  h2 {
    font-size: 18px;
    font-weight: 700;
    color: $slate-800;
    margin: 0;
    letter-spacing: -0.01em;
  }

  .version {
    font-size: 11px;
    color: $slate-400;
    padding: 2px 8px;
    background: $slate-50;
    border-radius: 999px;
    border: 1px solid $slate-200;
    font-family: $font-mono;
    font-weight: 500;
  }
}

// 功能区域
.features-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 16px 24px;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  grid-template-rows: repeat(2, 1fr);
  gap: 16px;
  flex: 1;
  min-height: 0;
}

.feature-card {
  background: #ffffff;
  border: 1px solid $slate-200;
  border-radius: 12px;
  padding: 24px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  overflow: hidden;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  font-family: $font-sans;

  &:hover {
    border-color: rgba($mint, 0.3);
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.07), 0 2px 4px -2px rgba(0, 0, 0, 0.05);

    .card-footer {
      color: $mint;

      .el-icon {
        transform: translateX(4px);
      }
    }

    .card-icon {
      background: rgba($mint, 0.12);
      color: $mint;
      border-color: rgba($mint, 0.2);
    }
  }

  &.color-accent:hover {
    border-color: rgba($blue, 0.3);

    .card-footer {
      color: $blue;
    }

    .card-icon {
      background: rgba($blue, 0.12);
      color: $blue;
      border-color: rgba($blue, 0.2);
    }
  }

  &.color-success:hover {
    border-color: rgba($green, 0.3);

    .card-footer {
      color: $green;
    }

    .card-icon {
      background: rgba($green, 0.12);
      color: $green;
      border-color: rgba($green, 0.2);
    }
  }

  &.color-warning:hover {
    border-color: rgba($amber, 0.3);

    .card-footer {
      color: $amber;
    }

    .card-icon {
      background: rgba($amber, 0.12);
      color: $amber;
      border-color: rgba($amber, 0.2);
    }
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.card-icon {
  width: 48px;
  height: 48px;
  background: $slate-50;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: $slate-400;
  font-size: 24px;
  transition: all 0.2s ease;
  border: 1px solid $slate-200;
}

.card-stats {
  font-size: 11px;
  color: $slate-500;
  padding: 3px 10px;
  background: $slate-50;
  border-radius: 999px;
  border: 1px solid $slate-200;
  font-weight: 500;
  letter-spacing: 0.02em;
}

.card-body {
  h4 {
    font-size: 16px;
    font-weight: 600;
    color: $slate-800;
    margin: 0 0 4px;
  }

  .card-source {
    font-size: 11px;
    color: $slate-400;
    margin: 0 0 10px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-weight: 500;
  }

  p {
    font-size: 13px;
    color: $slate-500;
    line-height: 1.6;
    margin: 0;
  }
}

.card-footer {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: $slate-400;
  transition: color 0.2s ease;
  font-weight: 500;
  margin-top: 12px;

  .el-icon {
    font-size: 16px;
    transition: transform 0.2s ease;
  }
}

// 响应式适配
@media (max-width: 768px) {
  .features-grid {
    grid-template-columns: 1fr;
    grid-template-rows: none;
    gap: 12px;
  }

  .feature-card {
    padding: 16px;
  }
}
</style>
