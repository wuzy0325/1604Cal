<template>
  <aside class="sidebar">
    <div class="logo">
      <h2>1604校准系统</h2>
    </div>
    <nav class="nav-menu">
      <router-link
        to="/"
        class="nav-item"
        :class="{ active: $route.path === '/' }"
      >
        <el-icon>
          <HomeFilled />
        </el-icon>
        <span>首页</span>
      </router-link>
      <router-link
        to="/measurement"
        class="nav-item"
        :class="{ active: $route.path === '/measurement' }"
      >
        <el-icon>
          <Tools />
        </el-icon>
        <span>计量工作台</span>
      </router-link>
      <router-link
        to="/multi-pressure"
        class="nav-item"
        :class="{ active: $route.path === '/multi-pressure' }"
      >
        <el-icon>
          <CircleCheck />
        </el-icon>
        <span>多设备打压</span>
      </router-link>
      <router-link
        to="/calibration"
        class="nav-item"
        :class="{ active: $route.path === '/calibration' }"
      >
        <el-icon>
          <SetUp />
        </el-icon>
        <span>标定工作台</span>
      </router-link>
    </nav>
  </aside>
</template>

<script setup lang="ts">
import { HomeFilled, Tools, CircleCheck, SetUp } from '@element-plus/icons-vue'
</script>

<style scoped lang="scss">
.sidebar {
  width: 240px;
  height: 100vh;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;

  .logo {
    padding: var(--spacing-lg);
    border-bottom: 1px solid var(--border-color);

    h2 {
      color: var(--text-primary);
      font-size: 18px;
      margin: 0;
    }
  }

  .nav-menu {
    flex: 1;
    padding: var(--spacing-md);

    .nav-item {
      display: flex;
      align-items: center;
      gap: var(--spacing-sm);
      padding: var(--spacing-sm) var(--spacing-md);
      color: var(--text-secondary);
      text-decoration: none;
      border-radius: var(--radius-sm);
      margin-bottom: var(--spacing-xs);
      transition: all 0.3s;

      &:hover {
        background: var(--bg-tertiary);
        color: var(--text-primary);
      }

      &.active {
        background: var(--accent-primary);
        color: var(--text-primary);
      }

      .el-icon {
        font-size: 18px;
      }
    }
  }
}
</style>
