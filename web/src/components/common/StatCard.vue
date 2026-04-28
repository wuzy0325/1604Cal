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
/* ── 设计系统令牌 ── */
$font-sans: 'Inter', 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
$font-mono: 'JetBrains Mono', 'Fira Code', 'SF Mono', Consolas, monospace;
$mint: #10b981;
$slate-50: #f9fafb;
$slate-100: #f3f4f6;
$slate-200: #e5e7eb;
$slate-300: #d1d5db;
$slate-400: #9ca3af;
$slate-500: #6b7280;
$slate-600: #4b5563;
$slate-700: #374151;
$slate-800: #1f2937;

.stat-card {
  background: #ffffff;
  border: 1px solid $slate-200;
  border-radius: 10px;
  padding: 12px 20px;
  text-align: center;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  transition: all 0.2s ease;
  font-family: $font-sans;
  min-width: 100px;

  &:hover {
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.07), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
    border-color: rgba($mint, 0.3);
  }

  .label {
    color: $slate-500;
    font-size: 12px;
    margin-bottom: 4px;
    font-weight: 500;
    letter-spacing: 0.05em;
  }

  .value {
    color: $slate-800;
    font-size: 24px;
    font-weight: 700;
    font-family: $font-mono;
    line-height: 1.2;
  }
}
</style>
