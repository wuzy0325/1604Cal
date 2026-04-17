<template>
  <section class="module-hub-page">
    <div class="desktop-shell">
      <header class="hero">
        <p class="hero-caption">
          1604 融合系统入口
        </p>
        <h1>设备管理 + 计量 + 标定统一入口</h1>
        <p>
          本系统融合两个历史程序能力，并把设备管理独立为统一模块，供计量与标定模块共享设备选择结果。
        </p>
      </header>

      <section class="module-grid">
        <article class="module-card">
          <div class="card-header">
            <p class="module-tag">
              Device Registry
            </p>
            <el-icon class="card-icon">
              <Tools />
            </el-icon>
          </div>
          <h2>设备管理模块</h2>
          <p class="module-source">
            融合能力：统一设备台账（共享给全部业务模块）
          </p>
          <ul>
            <li>集中维护计量设备与打压设备连接状态</li>
            <li>维护单位一致性检查与连接异常追踪</li>
            <li>作为计量/标定模块的统一设备来源</li>
          </ul>
          <RouterLink
            class="entry-btn"
            :to="{ name: 'module-device-management' }"
          >
            <el-icon><ArrowRight /></el-icon>
            进入设备管理
          </RouterLink>
        </article>

        <article class="module-card">
          <div class="card-header">
            <p class="module-tag">
              Measurement
            </p>
            <el-icon class="card-icon">
              <DataLine />
            </el-icon>
          </div>
          <h2>计量模块</h2>
          <p class="module-source">
            对应旧系统：1605MeassureApp
          </p>
          <ul>
            <li>统一设备台账、连接状态、单位门禁</li>
            <li>适合作业准备、状态巡检与异常定位</li>
            <li>作为标定流程前置检查入口</li>
          </ul>
          <RouterLink
            class="entry-btn"
            :to="{ name: 'module-measurement' }"
          >
            <el-icon><ArrowRight /></el-icon>
            进入计量模块
          </RouterLink>
        </article>

        <article class="module-card">
          <div class="card-header">
            <p class="module-tag">
              Calibration
            </p>
            <el-icon class="card-icon">
              <SetUp />
            </el-icon>
          </div>
          <h2>标定模块</h2>
          <p class="module-source">
            对应旧系统：1604标定软件
          </p>
          <ul>
            <li>会话状态机驱动的流程控制</li>
            <li>自动/手动动作与过程状态可视化</li>
            <li>报告模板选择与结果输出闭环</li>
          </ul>
          <RouterLink
            class="entry-btn"
            :to="{ name: 'module-calibration' }"
          >
            <el-icon><ArrowRight /></el-icon>
            进入标定模块
          </RouterLink>
        </article>

        <article class="module-card">
          <div class="card-header">
            <p class="module-tag">
              Multi-Pressure
            </p>
            <el-icon class="card-icon">
              <Odometer />
            </el-icon>
          </div>
          <h2>多设备打压模块</h2>
          <p class="module-source">
            独立模块：并发打压控制
          </p>
          <ul>
            <li>同时控制多台打压设备</li>
            <li>实时压力监控与稳定检测</li>
            <li>独立设备注册/注销管理</li>
          </ul>
          <RouterLink
            class="entry-btn"
            :to="{ name: 'module-multi-pressure' }"
          >
            <el-icon><ArrowRight /></el-icon>
            进入打压模块
          </RouterLink>
        </article>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { Tools, DataLine, SetUp, ArrowRight, Odometer } from '@element-plus/icons-vue'
</script>

<style scoped lang="scss">
.module-hub-page {
  padding: var(--spacing-lg);
  box-sizing: border-box;
  height: 100%;
  overflow: hidden;
  background: var(--bg-primary);
}

.desktop-shell {
  max-width: 1200px;
  height: 100%;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.hero {
  margin-bottom: 0;
  padding-bottom: var(--spacing-xl);
  border-bottom: 1px solid var(--border-color);
}

.hero-caption {
  color: var(--accent-primary);
  font-size: 11px;
  letter-spacing: 0.1em;
  margin: 0 0 var(--spacing-sm);
  text-transform: uppercase;
  font-weight: 600;
}

.hero h1 {
  color: var(--text-primary);
  margin: 0 0 var(--spacing-sm);
  font-size: 22px;
  font-weight: 600;
}

.hero > p {
  color: var(--text-secondary);
  margin: 0;
  max-width: 800px;
  font-size: 14px;
  line-height: 1.6;
}

.module-grid {
  display: grid;
  gap: var(--spacing-md);
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

.module-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: var(--spacing-lg);
  display: flex;
  flex-direction: column;
  transition: all 0.15s ease;

  &:hover {
    border-color: var(--accent-primary);
  }

  &:first-child {
    border-top: 2px solid var(--accent-primary);
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-md);
}

.module-tag {
  color: var(--text-muted);
  font-size: 10px;
  letter-spacing: 0.1em;
  margin: 0;
  text-transform: uppercase;
}

.card-icon {
  font-size: 20px;
  color: var(--accent-primary);
}

.module-card h2 {
  color: var(--text-primary);
  margin: 0 0 var(--spacing-xs);
  font-size: 18px;
  font-weight: 600;
}

.module-source {
  color: var(--text-secondary);
  margin: 0 0 var(--spacing-md);
  font-size: 13px;
}

.module-card ul {
  color: var(--text-secondary);
  margin: 0 0 var(--spacing-md);
  padding-left: 16px;
  flex: 1;
}

.module-card li {
  margin-bottom: 2px;
  font-size: 13px;
  line-height: 1.6;
}

.entry-btn {
  background: var(--accent-primary);
  border-radius: 3px;
  color: var(--bg-primary);
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: 6px 14px;
  text-decoration: none;
  font-weight: 600;
  font-size: 13px;
  transition: all 0.15s ease;

  &:hover {
    background: var(--accent-secondary);
  }

  .el-icon {
    font-size: 14px;
  }
}

@media (max-width: 1200px) {
  .module-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 900px) {
  .module-hub-page {
    padding: var(--spacing-md);
  }

  .desktop-shell {
    max-width: 100%;
    height: auto;
    gap: var(--spacing-md);
  }

  .module-grid {
    grid-template-columns: 1fr;
    min-height: auto;
  }

  .hero h1 {
    font-size: 20px;
  }
}
</style>
