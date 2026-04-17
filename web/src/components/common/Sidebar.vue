<template>
  <aside class="sidebar">
    <div class="logo">
      <span class="logo-icon">&#9733;</span>
      <span class="logo-text">1604校准系统</span>
    </div>
    <div class="sidebar-search">
      <input type="text" placeholder="搜索..." aria-label="搜索" />
    </div>
    <nav class="nav-menu">
      <div class="nav-section-title">模块</div>
      <router-link
        :to="{ name: 'module-hub' }"
        class="nav-item"
        :class="{ active: $route.name === 'module-hub' }"
      >
        <el-icon>
          <HomeFilled />
        </el-icon>
        <span>模块入口</span>
      </router-link>
      <router-link
        :to="{ name: 'module-device-management' }"
        class="nav-item"
        :class="{ active: $route.name === 'module-device-management' }"
      >
        <el-icon>
          <Setting />
        </el-icon>
        <span>设备管理</span>
      </router-link>
      <router-link
        :to="{ name: 'module-measurement' }"
        class="nav-item"
        :class="{ active: $route.name === 'module-measurement' }"
      >
        <el-icon>
          <Tools />
        </el-icon>
        <span>计量工作台</span>
      </router-link>
      <router-link
        :to="{ name: 'module-calibration' }"
        class="nav-item"
        :class="{ active: $route.name === 'module-calibration' }"
      >
        <el-icon>
          <SetUp />
        </el-icon>
        <span>标定工作台</span>
      </router-link>
      <router-link
        :to="{ name: 'module-multi-pressure' }"
        class="nav-item"
        :class="{ active: $route.name === 'module-multi-pressure' }"
      >
        <el-icon>
          <Odometer />
        </el-icon>
        <span>多设备打压</span>
      </router-link>
    </nav>
  </aside>
</template>

<script setup lang="ts">
import { HomeFilled, Odometer, Setting, SetUp, Tools } from '@element-plus/icons-vue'
</script>

<style scoped lang="scss">
.sidebar {
  width: 220px;
  height: 100vh;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;

  .logo {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 14px 16px;
    border-bottom: 1px solid var(--border-color);

    .logo-icon {
      color: var(--accent-primary);
      font-size: 18px;
    }

    .logo-text {
      color: var(--text-primary);
      font-size: 14px;
      font-weight: 700;
      letter-spacing: 0.5px;
    }
  }

  .sidebar-search {
    padding: 10px 12px;
    border-bottom: 1px solid var(--border-color);

    input {
      width: 100%;
      background: var(--bg-primary);
      border: 1px solid var(--border-color-strong);
      color: var(--text-primary);
      padding: 5px 10px;
      border-radius: 3px;
      font-size: 12px;
      outline: none;
      &:focus-visible {
        outline: 2px solid var(--accent-primary);
        outline-offset: -2px;
        border-radius: var(--radius-sm);
      }

      &:focus {
        border-color: var(--accent-primary);
      }

      &::placeholder {
        color: var(--text-muted);
      }
    }
  }

  .nav-section-title {
    padding: 16px 16px 4px;
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 1.5px;
    color: var(--text-muted);
    font-weight: 600;
  }

  .nav-menu {
    flex: 1;
    padding: 0;
    overflow-y: auto;

    .nav-item {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 8px 16px;
      color: var(--text-secondary);
      text-decoration: none;
      font-size: 13px;
      border-left: 2px solid transparent;
      transition: all 0.15s;

      &:hover {
        background: var(--bg-tertiary);
        color: var(--text-primary);
      }

      &.active {
        background: var(--bg-tertiary);
        color: var(--accent-primary);
        border-left-color: var(--accent-primary);
      }

      .el-icon {
        font-size: 16px;
      }

      /* 模块入口之后的分组间距 */
      &:first-child {
        margin-top: 4px;
      }
    }
  }
}
</style>
