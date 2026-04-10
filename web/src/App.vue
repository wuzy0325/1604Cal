<template>
  <div class="app dark-theme">
    <Sidebar />
    <main class="main-content">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'
import Sidebar from './components/common/Sidebar.vue'
</script>

<style scoped lang="scss">
.app {
  display: flex;
  min-height: 100vh;
  background: var(--bg-primary);

  .main-content {
    flex: 1;
    overflow: auto;
  }
}
</style>
