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
            <div class="header-title">
                <h2>欢迎使用</h2>
                <p>1604 设备管理 + 计量 + 标定融合系统</p>
            </div>
            <div class="header-meta">
                <span class="version">v2.0.0</span>
            </div>
        </header>

        <!-- 欢迎区域 -->
        <section class="welcome-section">
            <div class="welcome-card">
                <div class="welcome-content">
                    <h3>专业计量标定解决方案</h3>
                    <p>
                        本系统融合两个历史程序能力，并把设备管理独立为统一模块，供计量与标定模块共享设备选择结果。
                        采用现代化界面设计，提供流畅的操作体验。
                    </p>
                </div>
                <div class="welcome-decoration">
                    <div class="decoration-circle"></div>
                    <div class="decoration-ring"></div>
                </div>
            </div>
        </section>

        <!-- 功能卡片区域 -->
        <section class="features-section">
            <h3 class="section-title">功能入口</h3>
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

        <!-- 快捷操作区域 -->
        <section class="quick-actions">
            <h3 class="section-title">快捷操作</h3>
            <div class="actions-grid">
                <div class="action-item" @click="navigateTo('/device-management')">
                    <div class="action-icon primary">
                        <el-icon><Tools /></el-icon>
                    </div>
                    <span>设备管理</span>
                </div>
                <div class="action-item" @click="navigateTo('/measurement')">
                    <div class="action-icon accent">
                        <el-icon><DataLine /></el-icon>
                    </div>
                    <span>开始计量</span>
                </div>
                <div class="action-item" @click="navigateTo('/calibration')">
                    <div class="action-icon success">
                        <el-icon><SetUp /></el-icon>
                    </div>
                    <span>开始标定</span>
                </div>
                <div class="action-item" @click="navigateTo('/multi-pressure')">
                    <div class="action-icon warning">
                        <el-icon><Odometer /></el-icon>
                    </div>
                    <span>打压控制</span>
                </div>
            </div>
        </section>
    </PageLayout>
</template>

<style scoped lang="scss">

// 页面头部
.page-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    flex-shrink: 0;
    padding-bottom: $spacing-4;
    border-bottom: 1px solid $border-color-light;
    margin-bottom: $spacing-8;

    .header-title {
        h2 {
            font-size: 28px;
            font-weight: $font-weight-bold;
            color: $text-primary;
            margin: 0 0 $spacing-2;
            letter-spacing: -0.02em;
        }

        p {
            font-size: $font-size-lg;
            color: $text-secondary;
            margin: 0;
        }
    }

    .header-meta {
        .version {
            font-size: 11px;
            color: $text-tertiary;
            padding: 2px 8px;
            background: rgba($neutral-800, 0.5);
            border-radius: $radius-full;
            border: 1px solid $border-color;
            font-family: $font-family-mono;
        }
    }
}

// 欢迎区域
.welcome-section {
    flex-shrink: 0;
    margin-bottom: $spacing-8;
}

.welcome-card {
    position: relative;
    background: linear-gradient(135deg, rgba($primary-900, 0.4) 0%, rgba($primary-800, 0.2) 100%);
    border: 1px solid rgba($primary-500, 0.2);
    border-radius: $radius-xl;
    padding: $spacing-8;
    overflow: hidden;

    &::before {
        content: '';
        position: absolute;
        top: 0;
        right: 0;
        width: 40%;
        height: 100%;
        background: radial-gradient(circle at top right, rgba($primary-500, 0.15), transparent 70%);
        pointer-events: none;
    }
}

.welcome-content {
    position: relative;
    z-index: 1;
    max-width: 600px;

    h3 {
        font-size: $font-size-2xl;
        font-weight: $font-weight-semibold;
        color: $text-primary;
        margin: 0 0 $spacing-3;
    }

    p {
        font-size: $font-size-base;
        color: $text-secondary;
        line-height: $line-height-relaxed;
        margin: 0;
    }
}

// 功能区域
.features-section {
    flex-shrink: 0;
    margin-bottom: $spacing-8;
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

.features-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
    gap: $spacing-6;
}

.feature-card {
    background: $bg-surface;
    border: 1px solid $border-color;
    border-radius: $radius-lg;
    padding: $spacing-6;
    cursor: pointer;
    transition: all $transition-normal;
    position: relative;
    overflow: hidden;

    &:hover {
        transform: translateY(-4px);
        background: rgba($neutral-800, 0.8);
        border-color: $border-color-strong;
        box-shadow: $shadow-lg;

        .card-footer {
            color: $primary-400;

            .el-icon {
                transform: translateX(4px);
            }
        }
        
        .card-icon {
            background: rgba($primary-500, 0.2);
            color: $primary-400;
        }
    }
    
    &.color-accent:hover {
        .card-footer {
            color: $accent-400;
        }
        
        .card-icon {
            background: rgba($accent-500, 0.2);
            color: $accent-400;
        }
    }
    
    &.color-success:hover {
        .card-footer {
            color: $success-400;
        }
        
        .card-icon {
            background: rgba($success-500, 0.2);
            color: $success-400;
        }
    }
    
    &.color-warning:hover {
        .card-footer {
            color: $warning-400;
        }
        
        .card-icon {
            background: rgba($warning-500, 0.2);
            color: $warning-400;
        }
    }
}

.card-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: $spacing-5;
}

.card-icon {
    width: 48px;
    height: 48px;
    background: rgba($neutral-800, 0.8);
    border-radius: $radius-md;
    display: flex;
    align-items: center;
    justify-content: center;
    color: $text-tertiary;
    font-size: 24px;
    transition: all $transition-fast;
    border: 1px solid rgba($border-color, 0.5);
}

.card-stats {
    font-size: 10px;
    color: $text-tertiary;
    padding: 2px 8px;
    background: rgba($neutral-900, 0.5);
    border-radius: $radius-full;
    border: 1px solid $border-color-light;
    font-weight: 500;
}

.card-body {
    margin-bottom: $spacing-6;

    h4 {
        font-size: $font-size-lg;
        font-weight: $font-weight-semibold;
        color: $text-primary;
        margin: 0 0 $spacing-2;
    }

    .card-source {
        font-size: 11px;
        color: $text-muted;
        margin: 0 0 $spacing-2;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    p {
        font-size: $font-size-sm;
        color: $text-secondary;
        line-height: $line-height-normal;
        margin: 0;
    }
}

.card-footer {
    display: flex;
    align-items: center;
    gap: $spacing-2;
    font-size: $font-size-sm;
    color: $text-tertiary;
    transition: color $transition-fast;
    font-weight: $font-weight-medium;

    .el-icon {
        font-size: 14px;
        transition: transform $transition-fast;
    }
}

// 快捷操作
.quick-actions {
    flex-shrink: 0;
}

.actions-grid {
    display: flex;
    gap: $spacing-4;
    flex-wrap: wrap;
}

.action-item {
    display: flex;
    align-items: center;
    gap: $spacing-3;
    padding: $spacing-4 $spacing-5;
    background: $bg-surface;
    border: 1px solid $border-color;
    border-radius: $radius-md;
    cursor: pointer;
    transition: all $transition-fast;
    min-width: 180px;

    &:hover {
        background: rgba($neutral-800, 0.8);
        border-color: $border-color-strong;
        transform: translateY(-2px);
        box-shadow: $shadow-md;
    }

    span {
        font-size: $font-size-sm;
        color: $text-secondary;
        font-weight: $font-weight-medium;
    }
}

.action-icon {
    width: 32px;
    height: 32px;
    border-radius: $radius-sm;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;

    &.primary {
        background: rgba($primary-500, 0.1);
        color: $primary-400;
        border: 1px solid rgba($primary-500, 0.2);
    }

    &.success {
        background: rgba($success-500, 0.1);
        color: $success-400;
        border: 1px solid rgba($success-500, 0.2);
    }
    
    &.accent {
        background: rgba($accent-500, 0.1);
        color: $accent-400;
        border: 1px solid rgba($accent-500, 0.2);
    }
    
    &.warning {
        background: rgba($warning-500, 0.1);
        color: $warning-400;
        border: 1px solid rgba($warning-500, 0.2);
    }
}

// 响应式适配
@media (max-width: 768px) {
    .features-grid {
        grid-template-columns: 1fr;
    }
    
    .actions-grid {
        flex-direction: column;
    }
    
    .action-item {
        min-width: 100%;
    }
}
</style>
