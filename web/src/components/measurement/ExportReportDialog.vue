<template>
  <el-dialog
    :model-value="visible"
    title="导出报告"
    width="500px"
    :close-on-click-modal="false"
    class="export-dialog"
    @close="emit('close')"
  >
    <el-form label-width="100px">
      <el-form-item label="导出路径">
        <div style="display: flex; gap: 8px">
          <el-input v-model="exportPath" placeholder="请选择导出路径" readonly style="flex: 1" />
          <el-button type="primary" :icon="FolderOpened" @click="selectPath">选择路径</el-button>
        </div>
      </el-form-item>

      <el-form-item label="报告模板">
        <el-tag type="info" effect="dark">{{ templateName }}</el-tag>
      </el-form-item>

      <el-form-item label="校准点数">
        <span style="color: var(--text-primary)">{{ pointCount }} 点</span>
      </el-form-item>

      <el-form-item label="打压模式">
        <span style="color: var(--text-primary)">{{ modeLabel }}</span>
      </el-form-item>

      <el-form-item v-if="channelCount > 0" label="通道数">
        <span style="color: var(--text-primary)">{{ channelCount }}</span>
      </el-form-item>
    </el-form>

    <template #footer>
      <div style="display: flex; gap: 8px; justify-content: flex-end">
        <el-button @click="emit('close')">取消</el-button>
        <el-button
          type="primary"
          :loading="exporting"
          :disabled="!exportPath"
          :icon="Download"
          @click="emit('export', exportPath)"
        >
          {{ exporting ? '导出中...' : '导出报告' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { FolderOpened, Download } from '@element-plus/icons-vue'
import { showSaveDialog } from '@/composables/useFileSaveDialog'

const props = defineProps<{
  visible: boolean
  templateName: string
  pointCount: number
  channelCount: number
  pressureMode: string
  exporting: boolean
}>()

const emit = defineEmits<{
  close: []
  export: [path: string]
}>()

const exportPath = ref('')

const modeLabel = computed(() => props.pressureMode === 'single' ? '单程' : '回程')

async function selectPath() {
  const path = await showSaveDialog('measurement-report.xlsx', 'Excel 文件', '*.xlsx')
  if (path) {
    exportPath.value = path
  }
}
</script>

<style scoped lang="scss">
.export-dialog {
  :deep(.el-form-item) {
    margin-bottom: 18px;
  }
}
</style>
