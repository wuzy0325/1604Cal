<template>
  <el-dialog
    :model-value="props.visible"
    :title="mode === 'create' ? '新增设备配置' : '编辑设备配置'"
    width="520px"
    :close-on-click-modal="false"
    destroy-on-close
    @update:model-value="(val: boolean) => val ? null : emit('update:visible', false)"
    @closed="handleClosed"
  >
    <div class="form-grid">
      <label v-if="mode === 'edit'">
        <span>设备ID</span>
        <input
          :value="form.id"
          data-test="form-id"
          type="text"
          disabled
        >
      </label>

      <label>
        <span>名称</span>
        <input
          v-model.trim="form.name"
          data-test="form-name"
          type="text"
          placeholder="输入设备名称"
        >
      </label>

      <label>
        <span>类型</span>
        <select
          v-model="form.type"
          data-test="form-type"
        >
          <option value="measure">计量设备</option>
          <option value="pressure">打压设备</option>
        </select>
      </label>

      <label>
        <span>型号</span>
        <select
          v-model="form.model"
          data-test="form-model"
        >
          <option
            value=""
            disabled
          >选择型号</option>
          <option
            v-for="m in modelOptions"
            :key="m.value"
            :value="m.value"
          >
            {{ m.label }}
          </option>
        </select>
      </label>

      <label>
        <span>IP地址</span>
        <input
          v-model.trim="form.host"
          data-test="form-host"
          type="text"
          placeholder="192.168.1.xxx"
        >
      </label>

      <label>
        <span>端口</span>
        <input
          v-model.number="form.port"
          data-test="form-port"
          type="number"
          placeholder="9000"
        >
      </label>

      <label>
        <span>绑定本地IP</span>
        <input
          v-model.trim="form.localAddr"
          data-test="form-localAddr"
          type="text"
          placeholder="留空自动选择 / 多网卡时指定"
        >
      </label>
    </div>

    <p
      v-if="errorMessage"
      class="form-error"
    >
      <el-icon><Warning /></el-icon>
      {{ errorMessage }}
    </p>

    <template #footer>
      <el-button @click="handleCancel">
        取消
      </el-button>
      <el-button
        data-test="submit-form"
        type="primary"
        @click="handleSubmit"
      >
        <el-icon><Check /></el-icon>
        保存
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { Warning, Check } from '@element-plus/icons-vue'
import type { DeviceDTO } from '@/types/device'

// ---- Props & Emits ----

const props = defineProps<{
  /** 对话框是否可见 */
  visible: boolean
  /** 创建/编辑模式 */
  mode: 'create' | 'edit'
  /** 已有设备 ID 列表（创建模式下用于 ID 去重） */
  existingIds: string[]
  /** 编辑模式下的初始设备数据 */
  initialDevice?: DeviceDTO | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  /** 保存设备配置 */
  (e: 'save', device: DeviceDTO): void
  /** 取消关闭 */
  (e: 'cancel'): void
}>()

// ---- 表单状态 ----

type DeviceFormState = {
  id: string
  name: string
  type: DeviceDTO['type']
  model: string
  host: string
  port: number
  localAddr: string
  status: DeviceDTO['status']
}

const form = reactive<DeviceFormState>({
  id: '',
  name: '',
  type: 'measure',
  model: '',
  host: '',
  port: 9000,
  localAddr: '',
  status: 'disconnected'
})

const errorMessage = ref('')

// 设备型号选项，与设备类型联动
const modelOptions = computed(() => {
  if (form.type === 'measure') {
    return [{ value: 'WTN1604', label: 'WTN1604' }]
  }
  return [
    { value: 'ConST811A', label: 'ConST811A' },
    { value: 'ConST820', label: 'ConST820' },
    { value: 'ConST860', label: 'ConST860' },
    { value: 'SPC4000', label: 'SPC4000' }
  ]
})

// 切换设备类型时自动设置型号：仅在创建模式下生效
watch(() => form.type, () => {
  if (props.mode === 'create') {
    if (form.type === 'measure') {
      form.model = 'WTN1604'
    } else {
      form.model = ''
    }
  }
})

// ---- 对话框打开时初始化表单 ----

watch(() => props.visible, (isVisible) => {
  if (!isVisible) return
  if (props.mode === 'create') {
    initCreate()
  } else if (props.initialDevice) {
    initEdit(props.initialDevice)
  }
})

/** 重置为创建模式表单 */
function initCreate() {
  form.id = generateDeviceId()
  form.name = ''
  form.type = 'measure'
  form.model = ''
  form.host = ''
  form.port = 9000
  form.localAddr = ''
  form.status = 'disconnected'
  errorMessage.value = ''
}

/** 使用已有设备数据填充编辑模式表单 */
function initEdit(device: DeviceDTO) {
  form.id = device.id
  form.name = device.name
  form.type = device.type
  form.model = device.model
  form.host = device.host
  form.port = device.port
  form.localAddr = device.localAddr ?? ''
  form.status = device.status
  errorMessage.value = ''
}

// ---- 内部方法 ----

/** 生成唯一设备ID，格式 dev-{timestamp后6位}-{随机4位} */
function generateDeviceId(): string {
  const ts = Date.now().toString().slice(-6)
  const rand = Math.random().toString(36).slice(2, 6)
  return `dev-${ts}-${rand}`
}

function isValidIPv4(value: string): boolean {
  const segments = value.split('.')
  if (segments.length !== 4) return false
  return segments.every((segment) => {
    if (!/^\d+$/.test(segment)) return false
    const num = Number(segment)
    return num >= 0 && num <= 255
  })
}

function validateForm(): string {
  if (!form.host) return '请填写IP地址。'
  if (!isValidIPv4(form.host)) return 'IP地址格式不正确'
  if (!Number.isInteger(form.port) || form.port < 1 || form.port > 65535) return '端口必须在1-65535之间'
  if (!form.model) return '请选择设备型号'
  return ''
}

function handleCancel() {
  errorMessage.value = ''
  emit('cancel')
  emit('update:visible', false)
}

function handleSubmit() {
  errorMessage.value = ''
  const msg = validateForm()
  if (msg) {
    errorMessage.value = msg
    return
  }

  // 创建模式下，如果 ID 已存在则重新生成
  let id = form.id
  if (props.mode === 'create' && props.existingIds.includes(id)) {
    id = generateDeviceId()
  }

  emit('save', {
    id,
    name: form.name,
    type: form.type,
    model: form.model,
    host: form.host,
    port: form.port,
    localAddr: form.localAddr || undefined,
    status: 'disconnected'
  })
}

function handleClosed() {
  errorMessage.value = ''
}
</script>

<style scoped lang="scss">
.form-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(2, 1fr);
}

.form-full {
  grid-column: 1 / -1;
}

.form-grid label {
  color: $slate-500;
  display: flex;
  flex-direction: column;
  font-size: 12px;
  gap: 4px;
  font-weight: 500;
}

.form-grid input,
.form-grid select {
  background: #ffffff;
  border: 1px solid $slate-300;
  border-radius: 6px;
  color: $slate-700;
  padding: 6px 8px;
  font-size: 12px;
  font-family: $font-sans;
  outline: none;

  &:focus {
    border-color: $mint;
    box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.12);
  }
}

.form-error {
  color: $red;
  margin: 8px 0 0;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;

  .el-icon {
    font-size: 14px;
  }
}

@media (max-width: 900px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>