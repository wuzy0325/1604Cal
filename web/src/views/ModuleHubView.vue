<script setup lang="ts">
/**
 * @file 模块中心视图
 * @description 应用主页面，包含导航和功能入口
 * @version 2.1.0
 */
import { useRouter } from 'vue-router'
import { ArrowRight, Tools, DataLine, SetUp, Odometer } from '@element-plus/icons-vue'
import PageLayout from '@/components/common/PageLayout.vue'

const router = useRouter()

const featureCards = [
  {
    title: '设备管理模块',
    description: '集中维护计量设备与打压设备连接状态，维护单位一致性检查与连接异常追踪。',
    icon: Tools,
    path: '/device-management',
    stats: '统一台账',
    source: '融合能力'
  },
  {
    title: '计量模块',
    description: '统一设备台账、连接状态、单位门禁，适合作业准备、状态巡检与异常定位。',
    icon: DataLine,
    path: '/measurement',
    stats: '数据采集',
    source: '1605 MEASUREAPP'
  },
  {
    title: '标定工作台',
    description: '会话状态机驱动的流程控制，自动/手动动作与过程状态可视化。',
    icon: SetUp,
    path: '/calibration',
    stats: '流程控制',
    source: '1604 标定软件'
  },
  {
    title: '多设备打压控制',
    description: '同时控制多台打压设备，实时压力监控与稳定检测。',
    icon: Odometer,
    path: '/multi-pressure',
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
            <p class="card-desc">{{ card.description }}</p>
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

// 颜色系统 - 参考图片风格
$blue-50: #eff6ff;
$blue-100: #dbeafe;
$blue-200: #bfdbfe;
$blue-500: #3b82f6;
$blue-600: #2563eb;
$blue-700: #1d4ed8;

$gray-50: #f9fafb;
$gray-100: #f3f4f6;
$gray-200: #e5e7eb;
$gray-300: #d1d5db;
$gray-400: #9ca3af;
$gray-500: #6b7280;
$gray-600: #4b5563;
$gray-700: #374151;
$gray-800: #1f2937;
$gray-900: #111827;

$white: #ffffff;

// 页面头部
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
  height: 56px;
  padding: 0 32px;
  background: $white;
  border-bottom: 1px solid $gray-100;
  font-family: $font-sans;

  h2 {
    font-size: 20px;
    font-weight: 700;
    color: $gray-800;
    margin: 0;
    letter-spacing: -0.01em;
  }

  .version {
    font-size: 12px;
    color: $gray-400;
    padding: 4px 10px;
    background: $gray-50;
    border-radius: 999px;
    border: 1px solid $gray-200;
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
  padding: 24px 32px;
  background: $white;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
  flex: 1;
  min-height: 0;
  align-content: start;
}

.feature-card {
  background: $white;
  border: 1px solid $gray-200;
  border-radius: 16px;
  padding: 28px;
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.02);
  font-family: $font-sans;

  &:hover {
    border-color: $blue-200;
    box-shadow: 0 10px 25px -5px rgba(59, 130, 246, 0.08), 0 4px 10px -4px rgba(59, 130, 246, 0.04);
    transform: translateY(-2px);

    .card-footer {
      color: $blue-600;

      .el-icon {
        transform: translateX(4px);
      }
    }

    .card-icon {
      background: $blue-100;
      color: $blue-600;
    }
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.card-icon {
  width: 44px;
  height: 44px;
  background: $blue-50;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: $blue-500;
  font-size: 22px;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  flex-shrink: 0;
}

.card-stats {
  font-size: 12px;
  color: $gray-500;
  padding: 4px 12px;
  background: $gray-50;
  border-radius: 999px;
  border: 1px solid $gray-200;
  font-weight: 500;
  letter-spacing: 0.02em;
  flex-shrink: 0;
}

.card-body {
  flex: 1;

  h4 {
    font-size: 18px;
    font-weight: 600;
    color: $gray-800;
    margin: 0 0 8px;
    line-height: 1.3;
  }

  .card-source {
    font-size: 13px;
    color: $blue-500;
    margin: 0 0 12px;
    font-weight: 500;
    letter-spacing: 0.01em;
  }

  .card-desc {
    font-size: 14px;
    color: $gray-500;
    line-height: 1.7;
    margin: 0;
  }
}

.card-footer {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: $blue-500;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  font-weight: 500;
  margin-top: 20px;

  .el-icon {
    font-size: 16px;
    transition: transform 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  }
}

// 响应式适配
@media (max-width: 768px) {
  .page-header {
    padding: 0 20px;
    height: 52px;
  }

  .features-section {
    padding: 16px 20px;
  }

  .features-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .feature-card {
    padding: 20px;
  }
}
</style>
