<template>
  <div class="stat-card">
    <div class="label">
      {{ label }}
    </div>
    <div
      class="value"
      :style="{ color: color }"
    >
      {{ value }}
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  label: string
  value: string | number
  color?: string
}>()
</script>

<style scoped lang="scss">
.stat-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  text-align: center;
  
  .label {
    color: var(--text-secondary);
    font-size: 12px;
    margin-bottom: var(--spacing-xs);
  }
  
  .value {
    color: var(--text-primary);
    font-size: 24px;
    font-weight: bold;
  }
}
</style>
