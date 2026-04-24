<template>
  <section class="point-flow-panel">
    <header class="panel-header">
      <div>
        <p class="panel-caption">Pressure Point Flow</p>
        <h3>测点流程</h3>
      </div>
      <el-tag :type="pressureReady ? 'success' : 'warning'" size="small">
        {{ pressureReady ? '打压设备就绪' : '打压设备未就绪' }}
      </el-tag>
    </header>

    <div v-if="points.length === 0" class="empty-state">
      暂无测点，请先调整参数并生成测点。
    </div>

    <div v-else class="point-list">
      <article
        v-for="point in points"
        :key="point.index"
        class="point-item"
        :class="[
          `status-${point.status}`,
          { active: point.active }
        ]"
      >
        <div class="point-main">
          <span class="point-index">P{{ point.index + 1 }}</span>
          <div class="point-meta">
            <span class="target-pressure">{{ point.targetPressure.toFixed(2) }} kPa</span>
            <span class="point-status">{{ statusText(point.status) }}</span>
          </div>
        </div>

        <div class="point-actions">
          <el-button
            size="small"
            :disabled="!pressureReady || loading"
            @click="$emit('pressurize', point.index)"
          >
            打压
          </el-button>
          <el-button
            text
            size="small"
            @click="$emit('focus', point.index)"
          >
            定位
          </el-button>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
type PointFlowStatus =
  | 'pending'
  | 'pressurizing'
  | 'stabilizing'
  | 'collecting'
  | 'paused'
  | 'completed'
  | 'error'

interface PointFlowItem {
  index: number
  targetPressure: number
  status: PointFlowStatus
  active: boolean
}

defineProps<{
  points: PointFlowItem[]
  pressureReady: boolean
  loading: boolean
}>()

defineEmits<{
  pressurize: [pointIndex: number]
  focus: [pointIndex: number]
}>()

function statusText(status: PointFlowStatus): string {
  switch (status) {
    case 'pressurizing':
      return '打压中'
    case 'stabilizing':
      return '稳定中'
    case 'collecting':
      return '采集中'
    case 'paused':
      return '已暂停'
    case 'completed':
      return '已完成'
    case 'error':
      return '异常'
    default:
      return '待执行'
  }
}
</script>

<style scoped lang="scss">
.point-flow-panel {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-sm) var(--spacing-md);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
  flex-shrink: 0;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;

  h3 {
    margin: 0;
    color: var(--text-primary);
    font-size: 14px;
    font-weight: 600;
  }
}

.panel-caption {
  margin: 0;
  color: var(--text-muted);
  font-size: 11px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.point-list {
  display: flex;
  gap: var(--spacing-xs);
  overflow-x: auto;
  padding-bottom: 2px;
}

.point-item {
  min-width: 220px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: 6px var(--spacing-xs);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;

  &.active {
    border-color: var(--accent-primary);
    box-shadow: 0 0 0 1px color-mix(in srgb, var(--accent-primary) 45%, transparent);
  }

  &.status-completed {
    border-color: color-mix(in srgb, var(--status-success) 35%, var(--border-color));
  }

  &.status-error {
    border-color: color-mix(in srgb, var(--status-error) 50%, var(--border-color));
  }
}

.point-main {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.point-index {
  width: 28px;
  height: 28px;
  border-radius: 999px;
  background: var(--bg-quaternary);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--text-primary);
  font-size: 11px;
  font-weight: 600;
  flex-shrink: 0;
}

.point-meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.target-pressure {
  color: var(--text-primary);
  font-size: 12px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.point-status {
  color: var(--text-secondary);
  font-size: 10px;
}

.point-actions {
  display: flex;
  align-items: center;
  gap: 2px;
}

.empty-state {
  border: 1px dashed var(--border-color-strong);
  border-radius: var(--radius-sm);
  padding: var(--spacing-md);
  color: var(--text-muted);
  font-size: 13px;
}

@media (max-width: 900px) {
  .point-flow-panel {
    padding: var(--spacing-sm);
  }

  .panel-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-xs);
  }
}
</style>
