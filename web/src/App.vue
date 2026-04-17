<template>
  <div class="app dark-theme">
    <div class="main-wrapper">
      <main class="main-content">
        <RouterView />
      </main>
      <footer class="status-bar">
        <div class="status-left">
          <span class="status-dot" />
          <span>1604 校准系统</span>
          <span class="divider">|</span>
          <span>Go + Vue 3</span>
        </div>
        <div class="status-right">
          <span>v1.0.0</span>
        </div>
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<style scoped lang="scss">
.app {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background: var(--bg-primary);

  .main-wrapper {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .main-content {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    overflow-x: hidden;
  }

  .status-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: var(--bg-tertiary);
    border-top: 1px solid var(--border-color);
    padding: 0 16px;
    height: 24px;
    font-size: 11px;
    color: var(--text-secondary);
    flex-shrink: 0;
  }

  .status-left {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--status-success);
  }

  .divider {
    color: var(--text-muted);
  }

  .status-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }
}
</style>
