<template>
  <el-dialog v-model="visible" title="导出标定报告" width="500px" :close-on-click-modal="false" @close="reset">
    <el-form label-width="100px" label-position="left">
      <el-form-item label="保存路径">
        <el-input v-model="outputPath" placeholder="请选择导出路径" readonly>
          <template #append>
            <el-button @click="selectPath">浏览</el-button>
          </template>
        </el-input>
      </el-form-item>

      <el-form-item v-if="templateInfo" label="报告模板">
        <el-tag type="info">{{ templateInfo }}</el-tag>
      </el-form-item>

      <el-form-item v-if="exporting" label="导出进度">
        <el-progress :percentage="progress" :status="progressStatus" />
      </el-form-item>

      <el-form-item v-if="errorMessage" label="">
        <el-alert :title="errorMessage" type="error" show-icon :closable="false" />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="exporting" :disabled="!outputPath" @click="handleExport">
        导出报告
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { apiPost } from '@/api/client'
import { showSaveDialog } from '@/composables/useFileSaveDialog'

const emit = defineEmits<{
  done: []
}>()

const visible = ref(false)
const outputPath = ref('')
const exporting = ref(false)
const progress = ref(0)
const errorMessage = ref('')

const templateInfo = computed(() => {
  // 由父组件通过 v-if 更新，这里仅展示
  return ''
})

const progressStatus = computed(() => {
  if (errorMessage.value) return 'exception' as const
  if (progress.value >= 100) return 'success' as const
  return undefined
})

function open() {
  visible.value = true
  reset()
}

async function selectPath() {
  const path = await showSaveDialog('calibration_report.xlsx', 'Excel 文件', '*.xlsx')
  if (path) {
    outputPath.value = path
  }
}

async function handleExport() {
  if (!outputPath.value) return

  exporting.value = true
  progress.value = 10
  errorMessage.value = ''

  try {
    progress.value = 30
    await apiPost<{ status: string; path: string }>('/reports/export', { outputPath: outputPath.value })
    progress.value = 100
    ElMessage.success('报告导出成功')
    visible.value = false
    emit('done')
  } catch (e) {
    errorMessage.value = e instanceof Error ? e.message : '导出失败'
    progress.value = 0
  } finally {
    exporting.value = false
  }
}

function reset() {
  outputPath.value = ''
  exporting.value = false
  progress.value = 0
  errorMessage.value = ''
}
</script>
