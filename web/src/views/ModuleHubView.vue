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
          <p class="module-tag">
            Device Registry
          </p>
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
            进入设备管理
          </RouterLink>
        </article>

        <article class="module-card">
          <p class="module-tag">
            Measurement
          </p>
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
            进入计量模块
          </RouterLink>
        </article>

        <article class="module-card">
          <p class="module-tag">
            Calibration
          </p>
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
            进入标定模块
          </RouterLink>
        </article>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<style scoped>
.module-hub-page {
  background:
    radial-gradient(circle at 10% 20%, rgb(148 163 184 / 18%), transparent 38%),
    linear-gradient(158deg, #dbe3ea 0%, #eef2f6 52%, #dbe3ea 100%);
  min-height: 100vh;
  padding: 16px;
}

.desktop-shell {
  background: #f8fafc;
  border: 1px solid #b8c6d3;
  border-radius: 16px;
  box-shadow: 0 16px 42px rgb(15 23 42 / 14%);
  margin: 0 auto;
  max-width: 1728px;
  min-height: min(972px, calc(100vh - 32px));
  padding: 24px;
  width: min(1728px, calc(100vw - 32px));
}

.hero {
  border-bottom: 1px solid #d6e0ea;
  margin-bottom: 16px;
  padding-bottom: 16px;
}

.hero-caption {
  color: #475569;
  font-size: 12px;
  letter-spacing: 0.1em;
  margin: 0;
  text-transform: uppercase;
}

.hero h1 {
  color: #0f172a;
  margin: 8px 0;
}

.hero p {
  color: #334155;
  margin: 0;
  max-width: 980px;
}

.module-grid {
  display: grid;
  gap: 16px;
  grid-template-columns: repeat(3, minmax(260px, 1fr));
}

.module-card {
  background: #ffffff;
  border: 1px solid #d6e0ea;
  border-radius: 12px;
  box-shadow: 0 10px 28px rgb(15 23 42 / 8%);
  padding: 16px;
}

.module-tag {
  color: #64748b;
  font-size: 12px;
  letter-spacing: 0.1em;
  margin: 0;
  text-transform: uppercase;
}

.module-card h2 {
  color: #0f172a;
  margin: 8px 0 6px;
}

.module-source {
  color: #475569;
  margin: 0;
}

.module-card ul {
  color: #334155;
  margin: 12px 0;
  padding-left: 18px;
}

.module-card li {
  margin-bottom: 6px;
}

.entry-btn {
  background: linear-gradient(180deg, #1e293b 0%, #0f172a 100%);
  border-radius: 8px;
  color: #f8fafc;
  display: inline-block;
  padding: 8px 12px;
  text-decoration: none;
}

@media (max-width: 1380px) {
  .desktop-shell {
    min-height: auto;
  }
}

@media (max-width: 980px) {
  .module-grid {
    grid-template-columns: 1fr;
  }
}
</style>
