<template>
  <section class="device-panel">
    <header class="panel-header">
      <div>
        <h2>设备统一管理面板</h2>
        <p>统一维护计量设备与打压设备，实时观察状态、错误与连接策略。</p>
      </div>

      <div class="header-actions">
        <button
          data-test="refresh-devices"
          type="button"
          class="btn btn-ghost"
          @click="refreshAll"
        >
          立即刷新
        </button>
        <button
          data-test="add-device"
          type="button"
          class="btn btn-primary"
          @click="openCreateDialog"
        >
          新增设备
        </button>
      </div>
    </header>

    <section class="metric-grid">
      <article class="metric-card">
        <p>设备总数</p>
        <strong>{{ devices.length }}</strong>
      </article>
      <article class="metric-card">
        <p>计量设备</p>
        <strong>{{ measureCount }}</strong>
      </article>
      <article class="metric-card">
        <p>打压设备</p>
        <strong>{{ pressureCount }}</strong>
      </article>
      <article class="metric-card">
        <p>在线设备</p>
        <strong>{{ connectedCount }}</strong>
      </article>
      <article class="metric-card">
        <p>异常设备</p>
        <strong class="danger-text">{{ errorCount }}</strong>
      </article>
      <article class="metric-card">
        <p>单位一致性</p>
        <strong>{{ unitStatusText }}</strong>
      </article>
    </section>

    <section class="policy-strip">
      <p data-test="connect-policy">
        连接重试策略：{{ connectPolicyText }}
      </p>
      <p data-test="disconnect-policy">
        断开重试策略：{{ disconnectPolicyText }}
      </p>
      <p>
        最后刷新：{{ lastRefreshText }}
      </p>
      <label class="auto-refresh">
        <input
          v-model="autoRefresh"
          type="checkbox"
        >
        自动刷新（3秒）
      </label>
    </section>

    <section class="filter-bar">
      <label>
        设备类型
        <select v-model="typeFilter">
          <option value="all">全部</option>
          <option value="measure">计量设备</option>
          <option value="pressure">打压设备</option>
        </select>
      </label>

      <label>
        连接状态
        <select v-model="statusFilter">
          <option value="all">全部</option>
          <option value="connected">已连接</option>
          <option value="connecting">连接中</option>
          <option value="disconnected">未连接</option>
          <option value="error">异常</option>
        </select>
      </label>

      <label class="keyword-field">
        检索
        <input
          v-model.trim="keyword"
          type="text"
          placeholder="输入设备ID/名称/型号/IP"
        >
      </label>

      <button
        type="button"
        class="btn btn-ghost"
        @click="resetFilters"
      >
        重置筛选
      </button>
    </section>

    <section class="device-board">
      <p
        v-if="visibleDevices.length === 0"
        class="empty"
      >
        暂无符合筛选条件的设备
      </p>

      <article
        v-for="device in visibleDevices"
        :key="device.id"
        class="device-card"
      >
        <div class="card-top">
          <div class="title-block">
            <strong>{{ device.name || device.id }}</strong>
            <span class="type-badge">
              {{ typeLabel(device.type) }}
            </span>
          </div>
          <span :class="['status-badge', `status-${device.status}`]">
            {{ statusLabel(device.status) }}
          </span>
        </div>

        <div class="card-grid">
          <p>设备ID：{{ device.id }}</p>
          <p>型号：{{ device.model || '-' }}</p>
          <p>地址：{{ device.host }}:{{ device.port }}</p>
          <p>单位：{{ device.unit || '-' }}</p>
        </div>

        <p
          v-if="device.lastErrorReason"
          class="meta-error"
        >
          错误原因：{{ device.lastErrorReason }}
        </p>
        <p
          v-if="device.lastErrorAt"
          class="meta-error"
        >
          最近错误时间：{{ formatErrorTime(device.lastErrorAt) }}
        </p>

        <div class="card-actions">
          <button
            type="button"
            class="btn btn-ghost"
            @click="openEditDialog(device)"
          >
            编辑
          </button>
          <button
            type="button"
            class="btn btn-secondary"
            @click="toggleConnection(device)"
          >
            {{ device.status === 'connected' ? '断开' : '连接' }}
          </button>
        </div>
      </article>
    </section>

    <section
      v-if="dialogVisible"
      class="dialog-mask"
    >
      <div class="dialog-card">
        <h3>{{ dialogMode === 'create' ? '新增设备配置' : '编辑设备配置' }}</h3>

        <div class="form-grid">
          <label>
            设备ID
            <input
              v-model.trim="form.id"
              data-test="form-id"
              type="text"
              :disabled="dialogMode === 'edit'"
            >
          </label>

          <label>
            名称
            <input
              v-model.trim="form.name"
              data-test="form-name"
              type="text"
            >
          </label>

          <label>
            类型
            <select
              v-model="form.type"
              data-test="form-type"
            >
              <option value="measure">计量设备</option>
              <option value="pressure">打压设备</option>
            </select>
          </label>

          <label>
            型号
            <input
              v-model.trim="form.model"
              data-test="form-model"
              type="text"
            >
          </label>

          <label>
            IP
            <input
              v-model.trim="form.host"
              data-test="form-host"
              type="text"
            >
          </label>

          <label>
            端口
            <input
              v-model.number="form.port"
              data-test="form-port"
              type="number"
            >
          </label>

          <label>
            单位
            <input
              v-model.trim="form.unit"
              data-test="form-unit"
              type="text"
            >
          </label>
        </div>

        <p
          v-if="formError"
          class="form-error"
        >
          {{ formError }}
        </p>

        <div class="dialog-actions">
          <button
            type="button"
            class="btn btn-ghost"
            @click="closeDialog"
          >
            取消
          </button>
          <button
            data-test="submit-form"
            type="button"
            class="btn btn-primary"
            @click="submitForm"
          >
            保存
          </button>
        </div>
      </div>
    </section>

    <p
      v-if="errorMessage"
      class="error"
    >
      {{ errorMessage }}
    </p>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'

import {
  connectDevice,
  createEventStream,
  disconnectDevice,
  fetchDeviceConnectConfig,
  fetchDevices,
  fetchUnitConsistency,
  upsertDevice,
  type DeviceConnectConfigDTO,
  type DeviceDTO,
  type DeviceStatusChangedEventData,
  type StreamEventPayload
} from '@/services/apiClient'

type DeviceFilterType = 'all' | DeviceDTO['type']
type DeviceStatusFilter = 'all' | DeviceDTO['status']

type DeviceFormState = {
  id: string
  name: string
  type: DeviceDTO['type']
  model: string
  host: string
  port: number
  unit: string
  status: DeviceDTO['status']
}

const devices = ref<DeviceDTO[]>([])
const connectConfig = ref<DeviceConnectConfigDTO | null>(null)
const typeFilter = ref<DeviceFilterType>('all')
const statusFilter = ref<DeviceStatusFilter>('all')
const keyword = ref('')

const unitStatusText = ref('未检查')
const autoRefresh = ref(true)
const errorMessage = ref('')
const formError = ref('')
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const lastRefreshAt = ref<Date | null>(null)

const form = reactive<DeviceFormState>({
  id: '',
  name: '',
  type: 'measure',
  model: '',
  host: '',
  port: 9000,
  unit: 'kPa',
  status: 'disconnected'
})

const lastRefreshText = computed(() => {
  if (!lastRefreshAt.value) {
    return '--'
  }
  return lastRefreshAt.value.toLocaleTimeString()
})

const measureCount = computed(() => devices.value.filter((item) => item.type === 'measure').length)
const pressureCount = computed(() => devices.value.filter((item) => item.type === 'pressure').length)
const connectedCount = computed(() => devices.value.filter((item) => item.status === 'connected').length)
const errorCount = computed(() => devices.value.filter((item) => item.status === 'error').length)

const normalizedKeyword = computed(() => keyword.value.trim().toLowerCase())

const visibleDevices = computed(() => {
  return devices.value.filter((item) => {
    if (typeFilter.value !== 'all' && item.type !== typeFilter.value) {
      return false
    }
    if (statusFilter.value !== 'all' && item.status !== statusFilter.value) {
      return false
    }

    if (!normalizedKeyword.value) {
      return true
    }

    const fields = [item.id, item.name, item.model, item.host]
      .map((field) => field.toLowerCase())
      .join(' ')
    return fields.includes(normalizedKeyword.value)
  })
})

const connectPolicyText = computed(() => {
  if (!connectConfig.value) {
    return '--'
  }

  const cfg = connectConfig.value
  return `连接超时 ${cfg.connectAttemptTimeoutMs}ms，重试 ${cfg.connectMaxAttempts} 次，退避 ${cfg.connectInitialBackoffMs}-${cfg.connectMaxBackoffMs}ms`
})

const disconnectPolicyText = computed(() => {
  if (!connectConfig.value) {
    return '--'
  }

  const cfg = connectConfig.value
  return `断开超时 ${cfg.disconnectAttemptTimeoutMs}ms，重试 ${cfg.disconnectMaxAttempts} 次，退避 ${cfg.disconnectInitialBackoffMs}-${cfg.disconnectMaxBackoffMs}ms`
})

let pollTimer: ReturnType<typeof setInterval> | null = null
let eventSource: EventSource | null = null

function statusLabel(status: DeviceDTO['status']) {
  switch (status) {
    case 'connected':
      return '已连接'
    case 'connecting':
      return '连接中'
    case 'error':
      return '异常'
    default:
      return '未连接'
  }
}

function typeLabel(type: DeviceDTO['type']) {
  return type === 'measure' ? '计量设备' : '打压设备'
}

function formatErrorTime(value: string) {
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) {
    return value
  }
  return parsed.toLocaleString()
}

function isDeviceStatusChangedEventData(data: unknown): data is DeviceStatusChangedEventData {
  if (!data || typeof data !== 'object') {
    return false
  }

  const candidate = data as Record<string, unknown>
  return typeof candidate.id === 'string'
}

function applyDeviceStatusChangedEvent(data: DeviceStatusChangedEventData) {
  const index = devices.value.findIndex((item) => item.id === data.id)
  if (index < 0) {
    void refreshAll()
    return
  }

  const current = devices.value[index]
  const nextStatus = data.status ?? current.status

  let nextReason = current.lastErrorReason
  let nextTime = current.lastErrorAt

  if (nextStatus === 'connected' || nextStatus === 'disconnected') {
    nextReason = ''
    nextTime = undefined
  }

  if (typeof data.errorReason === 'string') {
    nextReason = data.errorReason
  }
  if (typeof data.lastErrorAt === 'string') {
    nextTime = data.lastErrorAt
  }

  devices.value.splice(index, 1, {
    ...current,
    status: nextStatus,
    lastErrorReason: nextReason,
    lastErrorAt: nextTime
  })
}

async function refreshAll() {
  errorMessage.value = ''
  try {
    const [list, consistency, policy] = await Promise.all([
      fetchDevices(),
      fetchUnitConsistency(),
      fetchDeviceConnectConfig()
    ])
    devices.value = list
    connectConfig.value = policy
    unitStatusText.value = consistency.consistent
      ? '全部设备单位一致'
      : `存在冲突：${consistency.conflicts.join(', ')}`
    lastRefreshAt.value = new Date()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '刷新设备状态失败'
  }
}

function resetFilters() {
  typeFilter.value = 'all'
  statusFilter.value = 'all'
  keyword.value = ''
}

function openCreateDialog() {
  dialogMode.value = 'create'
  form.id = ''
  form.name = ''
  form.type = 'measure'
  form.model = ''
  form.host = ''
  form.port = 9000
  form.unit = 'kPa'
  form.status = 'disconnected'
  formError.value = ''
  dialogVisible.value = true
}

function openEditDialog(device: DeviceDTO) {
  dialogMode.value = 'edit'
  form.id = device.id
  form.name = device.name
  form.type = device.type
  form.model = device.model
  form.host = device.host
  form.port = device.port
  form.unit = device.unit
  form.status = device.status
  formError.value = ''
  dialogVisible.value = true
}

function closeDialog() {
  dialogVisible.value = false
  formError.value = ''
}

async function submitForm() {
  formError.value = ''
  const validationMessage = validateForm()
  if (validationMessage) {
    formError.value = validationMessage
    return
  }

  try {
    await upsertDevice({
      id: form.id,
      name: form.name,
      type: form.type,
      model: form.model,
      host: form.host,
      port: form.port,
      unit: form.unit,
      status: form.status
    })
    dialogVisible.value = false
    await refreshAll()
  } catch (error) {
    formError.value = error instanceof Error ? error.message : '保存设备失败'
  }
}

async function toggleConnection(device: DeviceDTO) {
  errorMessage.value = ''
  try {
    if (device.status === 'connected') {
      await disconnectDevice(device.id)
    } else {
      await connectDevice(device.id)
    }
    await refreshAll()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '切换连接状态失败'
  }
}

function validateForm() {
  if (!form.id || !form.host || !form.unit) {
    return '请填写设备ID、IP和单位。'
  }

  if (dialogMode.value === 'create' && devices.value.some((item) => item.id === form.id)) {
    return '设备ID已存在'
  }

  if (!isValidIPv4(form.host)) {
    return 'IP地址格式不正确'
  }

  if (!Number.isInteger(form.port) || form.port < 1 || form.port > 65535) {
    return '端口必须在1-65535之间'
  }

  return ''
}

function isValidIPv4(value: string) {
  const segments = value.split('.')
  if (segments.length !== 4) {
    return false
  }

  return segments.every((segment) => {
    if (!/^\d+$/.test(segment)) {
      return false
    }

    const num = Number(segment)
    return num >= 0 && num <= 255
  })
}

function startPolling() {
  if (pollTimer) {
    return
  }
  pollTimer = setInterval(() => {
    void refreshAll()
  }, 3000)
}

function startEventStream() {
  if (eventSource) {
    return
  }

  eventSource = createEventStream((payload: StreamEventPayload) => {
    if (payload.type === 'device.status.changed') {
      if (isDeviceStatusChangedEventData(payload.data)) {
        applyDeviceStatusChangedEvent(payload.data)
      } else {
        void refreshAll()
      }
    }
  })
}

function stopEventStream() {
  if (!eventSource) {
    return
  }

  eventSource.close()
  eventSource = null
}

function stopPolling() {
  if (!pollTimer) {
    return
  }
  clearInterval(pollTimer)
  pollTimer = null
}

watch(autoRefresh, (enabled) => {
  if (enabled) {
    startPolling()
  } else {
    stopPolling()
  }
})

onMounted(() => {
  void refreshAll()
  startPolling()
  startEventStream()
})

onUnmounted(() => {
  stopPolling()
  stopEventStream()
})
</script>

<style scoped>
.device-panel {
  background:
    linear-gradient(180deg, #f9fbfd 0%, #eef3f8 100%),
    repeating-linear-gradient(90deg, rgb(148 163 184 / 8%) 0 1px, transparent 1px 24px);
  border: 1px solid #b8c6d3;
  border-radius: 14px;
  padding: 14px;
}

.panel-header {
  align-items: flex-start;
  display: flex;
  gap: 10px;
  justify-content: space-between;
}

.panel-header h2 {
  color: #0f172a;
  margin: 0;
}

.panel-header p {
  color: #334155;
  margin: 6px 0 0;
}

.header-actions,
.card-actions,
.dialog-actions {
  display: flex;
  gap: 8px;
}

.metric-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(3, minmax(120px, 1fr));
  margin-top: 12px;
}

.metric-card {
  background: #ffffff;
  border: 1px solid #d6e0ea;
  border-radius: 10px;
  padding: 8px;
}

.metric-card p {
  color: #475569;
  font-size: 12px;
  margin: 0;
}

.metric-card strong {
  color: #0f172a;
  display: block;
  font-size: 16px;
  margin-top: 2px;
}

.danger-text {
  color: #b91c1c;
}

.policy-strip {
  align-items: center;
  background: #e7edf4;
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  display: grid;
  gap: 8px;
  grid-template-columns: 1fr 1fr;
  margin-top: 10px;
  padding: 10px;
}

.policy-strip p {
  color: #1e293b;
  margin: 0;
}

.auto-refresh {
  align-items: center;
  color: #1e293b;
  display: flex;
  gap: 6px;
}

.filter-bar {
  align-items: end;
  background: #f8fafc;
  border: 1px solid #d6e0ea;
  border-radius: 10px;
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(4, minmax(110px, 1fr));
  margin-top: 10px;
  padding: 10px;
}

.filter-bar label {
  color: #334155;
  display: flex;
  flex-direction: column;
  font-size: 12px;
  gap: 4px;
}

.keyword-field {
  grid-column: span 2;
}

.filter-bar select,
.filter-bar input {
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 7px 8px;
}

.device-board {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(2, minmax(260px, 1fr));
  margin-top: 10px;
}

.empty {
  color: #64748b;
  grid-column: 1 / -1;
  margin: 6px 0;
}

.device-card {
  background: #ffffff;
  border: 1px solid #d6e0ea;
  border-radius: 10px;
  padding: 10px;
}

.card-top {
  align-items: center;
  display: flex;
  gap: 10px;
  justify-content: space-between;
}

.title-block {
  align-items: center;
  display: flex;
  gap: 8px;
}

.title-block strong {
  color: #0f172a;
}

.type-badge {
  background: #dbeafe;
  border-radius: 999px;
  color: #1d4ed8;
  font-size: 11px;
  font-weight: 600;
  padding: 2px 7px;
}

.card-grid {
  display: grid;
  gap: 4px;
  grid-template-columns: 1fr 1fr;
  margin-top: 8px;
}

.card-grid p {
  color: #334155;
  font-size: 13px;
  margin: 0;
}

.meta-error {
  color: #991b1b;
  margin: 8px 0 0;
}

.status-badge {
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  padding: 2px 8px;
}

.status-connected {
  background: #dcfce7;
  color: #166534;
}

.status-connecting {
  background: #fef3c7;
  color: #92400e;
}

.status-disconnected {
  background: #e5e7eb;
  color: #374151;
}

.status-error {
  background: #fee2e2;
  color: #991b1b;
}

.card-actions {
  justify-content: flex-end;
  margin-top: 10px;
}

.btn {
  border-radius: 8px;
  border: 1px solid transparent;
  cursor: pointer;
  font-size: 13px;
  padding: 6px 10px;
}

.btn-primary {
  background: linear-gradient(180deg, #1e293b 0%, #0f172a 100%);
  border-color: #0f172a;
  color: #f9fafb;
}

.btn-secondary {
  background: #1d4ed8;
  border-color: #1d4ed8;
  color: #f8fafc;
}

.btn-ghost {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #0f172a;
}

.dialog-mask {
  align-items: center;
  background: rgb(15 23 42 / 46%);
  display: flex;
  inset: 0;
  justify-content: center;
  position: fixed;
  z-index: 10;
}

.dialog-card {
  background: #f8fafc;
  border: 1px solid #cbd5e1;
  border-radius: 12px;
  min-width: min(680px, 92vw);
  padding: 16px;
}

.dialog-card h3 {
  color: #0f172a;
  margin: 0 0 12px;
}

.form-grid {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(2, minmax(140px, 1fr));
}

.form-grid label {
  color: #1f2937;
  display: flex;
  flex-direction: column;
  font-size: 13px;
  gap: 4px;
}

.form-grid input,
.form-grid select {
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 6px 8px;
}

.form-error,
.error {
  color: #b91c1c;
  margin: 10px 0 0;
}

.dialog-actions {
  justify-content: flex-end;
  margin-top: 14px;
}

@media (max-width: 1180px) {
  .metric-grid {
    grid-template-columns: repeat(2, minmax(120px, 1fr));
  }

  .device-board {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 900px) {
  .panel-header {
    flex-direction: column;
  }

  .policy-strip,
  .filter-bar,
  .card-grid,
  .form-grid {
    grid-template-columns: 1fr;
  }

  .keyword-field {
    grid-column: auto;
  }
}
</style>
