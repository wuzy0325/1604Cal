<template>
  <section class="module-page">
    <div class="desktop-shell">
      <header class="module-header">
        <div>
          <p class="module-caption">Calibration Workbench</p>
          <h1>标定工作台</h1>
        </div>
        <nav class="module-switch">
          <RouterLink class="switch-btn" :to="{ name: 'module-device-management' }">设备管理</RouterLink>
          <RouterLink class="switch-btn" :to="{ name: 'module-measurement' }">计量模块</RouterLink>
          <RouterLink class="switch-btn active" :to="{ name: 'module-calibration' }">标定模块</RouterLink>
          <RouterLink class="switch-btn" :to="{ name: 'module-multi-pressure' }">多设备打压</RouterLink>
          <RouterLink class="switch-btn switch-btn-ghost" :to="{ name: 'module-hub' }">
            <el-icon><ArrowLeft /></el-icon>返回
          </RouterLink>
        </nav>
      </header>

      <div class="calibration-layout">
        <CalibrationSidebar :collapsed="sidebarCollapsed" @toggle="sidebarCollapsed = !sidebarCollapsed" />
        <main class="workbench">
          <ProgressIndicator :current-step="calibrationStore.currentStep" />
          <CalibrationParams />
          <CalibrationControl />
          <CalibrationDataView @select-template="dialogsRef?.openTemplateDialog()" />
          <div v-if="dialogsRef?.templateFilename" class="template-result-bar">
            <el-icon><DocumentChecked /></el-icon>
            <span>当前报告模板：{{ dialogsRef.templateFilename }}</span>
          </div>
        </main>
      </div>

      <CalibrationDialogs ref="dialogsRef" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import { ArrowLeft, DocumentChecked } from '@element-plus/icons-vue'
import CalibrationSidebar from '@/components/calibration/CalibrationSidebar.vue'
import CalibrationParams from '@/components/calibration/CalibrationParams.vue'
import CalibrationControl from '@/components/calibration/CalibrationControl.vue'
import CalibrationDataView from '@/components/calibration/CalibrationDataView.vue'
import CalibrationDialogs from '@/components/calibration/CalibrationDialogs.vue'
import ProgressIndicator from '@/components/calibration/ProgressIndicator.vue'
import { useCalibrationStore } from '@/stores/calibration'
import { useCalibrationSync } from '@/composables/useCalibrationSync'
import { useConfigPersistence } from '@/composables/useConfigPersistence'

const calibrationStore = useCalibrationStore()
const sidebarCollapsed = ref(false)
const dialogsRef = ref<InstanceType<typeof CalibrationDialogs>>()

useCalibrationSync()
useConfigPersistence()
</script>

<style scoped lang="scss">
.module-page { padding: var(--spacing-lg); box-sizing: border-box; height: 100%; overflow: hidden; background: var(--bg-primary); }
.desktop-shell { max-width: 100%; height: 100%; margin: 0 auto; display: flex; flex-direction: column; gap: var(--spacing-md); }
.module-header { align-items: flex-end; border-bottom: 1px solid var(--border-color); display: flex; gap: var(--spacing-lg); justify-content: space-between; padding-bottom: var(--spacing-lg); flex-shrink: 0; min-height: 52px; }
.module-caption { color: var(--accent-primary); font-size: 11px; letter-spacing: 0.08em; margin: 0 0 var(--spacing-xs); text-transform: uppercase; font-weight: 600; }
.module-header h1 { color: var(--text-primary); margin: 0; font-size: 20px; font-weight: 600; }
.module-switch { display: flex; flex-wrap: wrap; gap: var(--spacing-xs); }
.switch-btn { background: var(--bg-tertiary); border: 1px solid var(--border-color); border-radius: var(--radius-sm); color: var(--text-secondary); padding: 5px 12px; text-decoration: none; font-size: 13px; transition: all 0.15s ease; display: inline-flex; align-items: center; gap: var(--spacing-xs);
  &:hover { background: var(--bg-quaternary); color: var(--text-primary); }
  .el-icon { font-size: 12px; }
}
.switch-btn.active { background: var(--accent-primary); border-color: var(--accent-primary); color: var(--bg-primary); font-weight: 600; }
.switch-btn-ghost { background: transparent; }
.calibration-layout { flex: 1; min-height: 0; display: flex; gap: 0; background: var(--bg-primary); }
.workbench { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: var(--spacing-lg); padding: var(--spacing-lg); overflow-y: auto; overflow-x: hidden; }
.template-result-bar { display: flex; align-items: center; gap: var(--spacing-xs); padding: var(--spacing-xs) var(--spacing-md); background: var(--status-success-bg-subtle); border: 1px solid rgba(78, 201, 176, 0.3); border-radius: var(--radius-sm); font-size: 13px; flex-shrink: 0;
  .el-icon { color: var(--status-success); }
  span { color: var(--text-secondary); }
}
@media (max-width: 900px) {
  .module-page { padding: var(--spacing-md); overflow: auto; }
  .desktop-shell { max-width: 100%; height: auto; gap: var(--spacing-md); }
  .module-header { flex-direction: column; }
  .calibration-layout { flex-direction: column; min-height: auto; }
  .workbench { overflow: visible; gap: var(--spacing-md); padding: var(--spacing-md); }
}
</style>
