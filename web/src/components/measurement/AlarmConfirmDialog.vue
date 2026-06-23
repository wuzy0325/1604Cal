<template>
  <el-dialog
    v-model="visible"
    title="计量报警"
    width="520px"
    :close-on-click-modal="false"
    :show-close="false"
    draggable
    class="measurement-alarm-dialog"
  >
    <div class="alarm-body">
      <el-alert type="warning" :closable="false" show-icon>
        <template #title>
          <span>采集数据精度超限 — 压力点 {{ point?.id || '未知' }}</span>
        </template>
        <template #default>
          <div class="alarm-summary">
            <div class="alarm-kv">
              <span class="kv-label">目标压力</span>
              <span class="kv-value">{{ (point?.targetPressure ?? 0).toFixed(4) }}</span>
            </div>
            <div class="alarm-kv">
              <span class="kv-label">实际压力</span>
              <span class="kv-value">{{ (point?.actualPressure ?? 0).toFixed(4) }}</span>
            </div>
            <div class="alarm-kv">
              <span class="kv-label">报警阈值</span>
              <span class="kv-value warn">{{ (alarm?.threshold ?? 0) * 100 }}%</span>
            </div>
            <div class="alarm-kv">
              <span class="kv-label">最大偏差</span>
              <span class="kv-value warn">{{ ((alarm?.maxDeviation ?? 0) * 100).toFixed(2) }}%</span>
            </div>
          </div>
        </template>
      </el-alert>

      <div v-if="alarm?.overLimitChannels?.length" class="overlimit-section">
        <p class="overlimit-label">超限通道（{{ alarm.overLimitChannels.length }} 个）：</p>
        <div class="channel-tags">
          <el-tag
            v-for="ch in alarm.overLimitChannels"
            :key="ch"
            type="danger"
            size="small"
            effect="dark"
          >
            通道 {{ ch }}
          </el-tag>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="decide('recollect')" type="warning" plain>重新采集</el-button>
        <el-button @click="decide('continue')" type="primary">忽略继续</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import type { MeasurementPoint } from '@/api/measurement'

const visible = defineModel<boolean>('visible', { default: false })

defineProps<{
  point?: Partial<MeasurementPoint>
  alarm?: {
    pointId: string
    targetPressure: number
    actualPressure: number
    threshold: number
    maxDeviation: number
    overLimitChannels: number[]
  } | null
}>()

const emit = defineEmits<{
  decision: [action: 'continue' | 'recollect']
}>()

function decide(action: 'continue' | 'recollect') {
  emit('decision', action)
  visible.value = false
}
</script>

<style scoped lang="scss">
.measurement-alarm-dialog {
  .alarm-body {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .alarm-summary {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px 16px;
    margin-top: 8px;
  }

  .alarm-kv {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .kv-label {
    color: var(--text-secondary);
    font-size: 12px;
  }

  .kv-value {
    font-weight: 600;
    font-variant-numeric: tabular-nums;

    &.warn {
      color: var(--status-warning);
    }
  }

  .overlimit-section {
    margin-top: 4px;
  }

  .overlimit-label {
    color: var(--text-secondary);
    font-size: 13px;
    margin: 0 0 8px;
  }

  .channel-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }
}
</style>
