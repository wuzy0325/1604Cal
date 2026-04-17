<template>
  <div class="progress-indicator">
    <h4 class="title">
      校准流程
    </h4>
    <div
      class="steps"
      role="progressbar"
      :aria-valuenow="currentStep"
      aria-valuemin="1"
      :aria-valuemax="steps.length"
    >
      <div
        v-for="(step, index) in steps"
        :key="index"
        class="step"
        :class="{
          completed: currentStep > index,
          active: currentStep === index,
          pending: currentStep < index
        }"
      >
        <div class="step-marker">
          <el-icon v-if="currentStep > index">
            <Check />
          </el-icon>
          <span v-else>{{ index + 1 }}</span>
        </div>
        <div class="step-label">
          {{ step }}
        </div>
        <div
          v-if="index < steps.length - 1"
          class="step-line"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Check } from '@element-plus/icons-vue'

defineProps<{
  currentStep: number
}>()

const steps = [
  '设备连接',
  '通道选择',
  '开始校准',
  '数据采集',
  '数据拟合',
  '完成'
]
</script>

<style scoped lang="scss">
.progress-indicator {
  padding: var(--spacing-md) 0;

  .title {
    color: var(--text-primary);
    margin: 0 0 var(--spacing-md) 0;
    font-size: 13px;
    font-weight: 500;
  }

  .steps {
    display: flex;
    position: relative;

    .step {
      flex: 1;
      display: flex;
      flex-direction: column;
      align-items: center;
      position: relative;

      .step-marker {
        width: 28px;
        height: 28px;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 11px;
        font-weight: bold;
        z-index: 2;

        .el-icon {
          font-size: 14px;
        }
      }

      .step-label {
        margin-top: var(--spacing-xs);
        font-size: 11px;
        text-align: center;
        white-space: nowrap;
      }

      .step-line {
        position: absolute;
        top: 14px;
        left: 50%;
        width: 100%;
        height: 2px;
        background: var(--border-color);
        z-index: 1;
      }

      &.completed {
        .step-marker {
          background: var(--status-success);
          color: var(--bg-primary);
        }

        .step-label {
          color: var(--status-success);
        }

        .step-line {
          background: var(--status-success);
        }
      }

      &.active {
        .step-marker {
          background: var(--accent-primary);
          color: var(--bg-primary);
          box-shadow: 0 0 0 4px rgba(255, 215, 0, 0.15);
        }

        .step-label {
          color: var(--accent-primary);
          font-weight: 600;
        }
      }

      &.pending {
        .step-marker {
          background: var(--bg-tertiary);
          color: var(--text-muted);
          border: 1px solid var(--border-color-strong);
        }

        .step-label {
          color: var(--text-muted);
        }
      }
    }
  }
}
</style>
