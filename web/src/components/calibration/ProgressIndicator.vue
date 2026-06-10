<template>
  <div class="progress-indicator">
    <h4 class="title">
      标定流程
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
  '开始标定',
  '数据采集',
  '数据拟合',
  '完成'
]
</script>

<style scoped lang="scss">
.progress-indicator {
  padding: 0;
  font-family: $font-sans;

  .title {
    color: $slate-500;
    margin: 0 0 6px 0;
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.05em;
    text-transform: uppercase;
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
        width: 24px;
        height: 24px;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 10px;
        font-weight: 700;
        z-index: 2;
        transition: all 0.2s ease;

        .el-icon {
          font-size: 14px;
        }
      }

      .step-label {
        margin-top: 4px;
        font-size: 10px;
        text-align: center;
        white-space: nowrap;
        font-weight: 500;
      }

      .step-line {
        position: absolute;
        top: 12px;
        left: 50%;
        width: 100%;
        height: 2px;
        background: $slate-200;
        z-index: 1;
        transition: background 0.3s ease;
      }

      &.completed {
        .step-marker {
          background: $green;
          color: #ffffff;
        }

        .step-label {
          color: $green;
        }

        .step-line {
          background: $green;
        }
      }

      &.active {
        .step-marker {
          background: $mint;
          color: #ffffff;
          box-shadow: 0 0 0 4px rgba(16, 185, 129, 0.15);
        }

        .step-label {
          color: $mint-dark;
          font-weight: 600;
        }
      }

      &.pending {
        .step-marker {
          background: #ffffff;
          color: $slate-400;
          border: 1px solid $slate-300;
        }

        .step-label {
          color: $slate-400;
        }
      }
    }
  }
}
</style>
