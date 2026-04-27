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
  background: $bg-card;
  border: 1px solid $border-color;
  border-radius: $radius-lg;
  padding: $spacing-4 $spacing-6;
  text-align: center;
  box-shadow: $shadow-card;
  transition: all $transition-normal;

  &:hover {
    box-shadow: $shadow-card-hover;
    border-color: rgba($primary-500, 0.3);
  }

  .label {
    color: $text-secondary;
    font-size: $font-size-xs;
    margin-bottom: $spacing-1;
    font-weight: $font-weight-medium;
  }

  .value {
    color: $text-primary;
    font-size: $font-size-3xl;
    font-weight: $font-weight-bold;
    font-family: $font-family-mono;
  }
}
</style>
