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
          <el-icon><Refresh /></el-icon>
          立即刷新
        </button>
        <button
          data-test="add-device"
          type="button"
          class="btn btn-primary"
          @click="openCreateDialog"
        >
          <el-icon><Plus /></el-icon>
          新增设备
        </button>
      </div>
    </header>

    <section class="metric-grid">
      <article class="metric-card">
        <p class="metric-label">
          设备总数
        </p>
        <strong class="metric-value">{{ devices.length }}</strong>
      </article>
      <article class="metric-card">
        <p class="metric-label">
          计量设备
        </p>
        <strong class="metric-value">{{ measureCount }}</strong>
      </article>
      <article class="metric-card">
        <p class="metric-label">
          打压设备
        </p>
        <strong class="metric-value">{{ pressureCount }}</strong>
      </article>
      <article class="metric-card">
        <p class="metric-label">
          在线设备
        </p>
        <strong class="metric-value success">{{ connectedCount }}</strong>
      </article>
      <article class="metric-card">
        <p class="metric-label">
          异常设备
        </p>
        <strong class="metric-value danger">{{ errorCount }}</strong>
      </article>
      <article class="metric-card">
        <p class="metric-label">
          单位一致性
        </p>
        <strong :class="['metric-value', unitConsistent ? 'success' : 'warning']">
          {{ unitStatusText }}
        </strong>
      </article>
    </section>

    <section class="policy-strip">
      <div class="policy-item">
        <el-icon><Link /></el-icon>
        <span data-test="connect-policy">{{ connectPolicyText }}</span>
      </div>
      <div class="policy-item">
        <el-icon><Close /></el-icon>
        <span data-test="disconnect-policy">{{ disconnectPolicyText }}</span>
      </div>
      <div class="policy-item">
        <el-icon><Timer /></el-icon>
        <span>最后刷新：{{ lastRefreshText }}</span>
      </div>
      <label class="auto-refresh">
        <input
          v-model="autoRefresh"
          type="checkbox"
        >
        <span>自动刷新（3秒）</span>
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
        <el-icon><RefreshRight /></el-icon>
        重置筛选
      </button>
    </section>

    <section class="device-board">
      <p
        v-if="visibleDevices.length === 0"
        class="empty"
      >
        <el-icon><InfoFilled /></el-icon>
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
            <span :class="['type-badge', device.type]">
              {{ typeLabel(device.type) }}
            </span>
          </div>
          <span :class="['status-badge', `status-${device.status}`]">
            <el-icon v-if="device.status === 'connected'"><CircleCheck /></el-icon>
            <el-icon v-else-if="device.status === 'error'"><CircleClose /></el-icon>
            <el-icon v-else-if="device.status === 'connecting'"><Loading /></el-icon>
            <el-icon v-else><Remove /></el-icon>
            {{ statusLabel(device.status) }}
          </span>
        </div>

        <div class="card-grid">
          <div class="info-item">
            <span class="info-label">设备ID</span>
            <span class="info-value">{{ device.id }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">型号</span>
            <span class="info-value">{{ device.model || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">地址</span>
            <span class="info-value">{{ device.host }}:{{ device.port }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">单位</span>
            <span class="info-value">{{ device.unit || '-' }}</span>
          </div>
        </div>

        <div
          v-if="device.lastErrorReason || device.lastErrorAt"
          class="error-section"
        >
          <p
            v-if="device.lastErrorReason"
            class="meta-error"
          >
            <el-icon><Warning /></el-icon>
            错误原因：{{ device.lastErrorReason }}
          </p>
          <p
            v-if="device.lastErrorAt"
            class="meta-error"
          >
            <el-icon><Clock /></el-icon>
            最近错误时间：{{ formatErrorTime(device.lastErrorAt) }}
          </p>
        </div>

        <div class="card-actions">
          <button
            type="button"
            class="btn btn-ghost"
            @click="openEditDialog(device)"
          >
            <el-icon><Edit /></el-icon>
            编辑
          </button>
          <button
            type="button"
            :class="['btn', device.status === 'connected' ? 'btn-danger' : 'btn-success']"
            @click="toggleConnection(device)"
          >
            <el-icon v-if="device.status === 'connected'">
              <Close />
            </el-icon>
            <el-icon v-else>
              <Link />
            </el-icon>
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
        <header class="dialog-header">
          <h3>{{ dialogMode === 'create' ? '新增设备配置' : '编辑设备配置' }}</h3>
          <button
            class="btn-close"
            @click="closeDialog"
          >
            <el-icon><Close /></el-icon>
          </button>
        </header>

        <div class="form-grid">
          <label>
            <span>设备ID</span>
            <input
              v-model.trim="form.id"
              data-test="form-id"
              type="text"
              :disabled="dialogMode === 'edit'"
              placeholder="输入唯一标识"
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
            <input
              v-model.trim="form.model"
              data-test="form-model"
              type="text"
              placeholder="输入设备型号"
            >
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

          <label class="form-full">
            <span>单位</span>
            <input
              v-model.trim="form.unit"
              data-test="form-unit"
              type="text"
              placeholder="kPa / MPa / bar"
            >
          </label>
        </div>

        <p
          v-if="formError"
          class="form-error"
        >
          <el-icon><Warning /></el-icon>
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
            <el-icon><Check /></el-icon>
            保存
          </button>
        </div>
      </div>
    </section>

    <p
      v-if="errorMessage"
      class="error-banner"
    >
      <el-icon><Warning /></el-icon>
      {{ errorMessage }}
    </p>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import {
  Refresh,
  Plus,
  Link,
  Close,
  Timer,
  RefreshRight,
  InfoFilled,
  CircleCheck,
  CircleClose,
  Loading,
  Remove,
  Warning,
  Clock,
  Edit,
  Check
} from '@element-plus/icons-vue'

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
const unitConsistent = ref(false)
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
  return `连接超时 ${cfg.connectAttemptTimeoutMs}ms，重试 ${cfg.connectMaxAttempts} 次`
})

const disconnectPolicyText = computed(() => {
  if (!connectConfig.value) {
    return '--'
  }

  const cfg = connectConfig.value
  return `断开超时 ${cfg.disconnectAttemptTimeoutMs}ms，重试 ${cfg.disconnectMaxAttempts} 次`
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
  return type === 'measure' ? '计量' : '打压'
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
    unitConsistent.value = consistency.consistent
    unitStatusText.value = consistency.consistent
      ? '一致'
      : '冲突'
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

<style scoped lang="scss">
.device-panel {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
}

.panel-header {
  align-items: flex-start;
  display: flex;
  gap: var(--spacing-md);
  justify-content: space-between;
  margin-bottom: var(--spacing-lg);
}

.panel-header h2 {
  color: var(--text-primary);
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.panel-header p {
  color: var(--text-secondary);
  margin: var(--spacing-xs) 0 0;
  font-size: 14px;
}

.header-actions {
  display: flex;
  gap: var(--spacing-sm);
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-sm) var(--spacing-md);
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  
  .el-icon {
    font-size: 16px;
  }
}

.btn-primary {
  background: var(--accent-primary);
  color: white;
  
  &:hover {
    background: #ff5773;
  }
}

.btn-success {
  background: var(--status-success);
  color: white;
  
  &:hover {
    background: #059669;
  }
}

.btn-danger {
  background: var(--status-error);
  color: white;
  
  &:hover {
    background: #dc2626;
  }
}

.btn-ghost {
  background: transparent;
  border-color: var(--border-color);
  color: var(--text-secondary);
  
  &:hover {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }
}

.metric-grid {
  display: grid;
  gap: var(--spacing-md);
  grid-template-columns: repeat(6, 1fr);
  margin-bottom: var(--spacing-lg);
}

.metric-card {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  text-align: center;
}

.metric-label {
  color: var(--text-secondary);
  font-size: 12px;
  margin: 0 0 var(--spacing-xs);
}

.metric-value {
  color: var(--text-primary);
  display: block;
  font-size: 24px;
  font-weight: 600;
  
  &.success {
    color: var(--status-success);
  }
  
  &.danger {
    color: var(--status-error);
  }
  
  &.warning {
    color: var(--status-warning);
  }
}

.policy-strip {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  margin-bottom: var(--spacing-lg);
  padding: var(--spacing-md) var(--spacing-lg);
}

.policy-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  color: var(--text-secondary);
  font-size: 13px;
  
  .el-icon {
    color: var(--accent-primary);
    font-size: 14px;
  }
}

.auto-refresh {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  color: var(--text-secondary);
  font-size: 13px;
  margin-left: auto;
  cursor: pointer;
  
  input[type="checkbox"] {
    width: 16px;
    height: 16px;
    accent-color: var(--accent-primary);
  }
}

.filter-bar {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  display: flex;
  align-items: flex-end;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-lg);
  padding: var(--spacing-md);
}

.filter-bar label {
  color: var(--text-secondary);
  display: flex;
  flex-direction: column;
  font-size: 13px;
  gap: var(--spacing-xs);
  flex: 1;
}

.keyword-field {
  flex: 2;
}

.filter-bar select,
.filter-bar input {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  padding: var(--spacing-sm);
}

.device-board {
  display: grid;
  gap: var(--spacing-md);
  grid-template-columns: repeat(2, 1fr);
}

.empty {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  color: var(--text-muted);
  padding: var(--spacing-xl);
  
  .el-icon {
    font-size: 20px;
  }
}

.device-card {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
}

.card-top {
  align-items: center;
  display: flex;
  gap: var(--spacing-md);
  justify-content: space-between;
  margin-bottom: var(--spacing-md);
}

.title-block {
  align-items: center;
  display: flex;
  gap: var(--spacing-sm);
}

.title-block strong {
  color: var(--text-primary);
  font-size: 16px;
}

.type-badge {
  background: var(--bg-secondary);
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  
  &.measure {
    color: var(--status-info);
  }
  
  &.pressure {
    color: var(--status-warning);
  }
}

.card-grid {
  display: grid;
  gap: var(--spacing-sm);
  grid-template-columns: 1fr 1fr;
  margin-bottom: var(--spacing-md);
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.info-label {
  color: var(--text-muted);
  font-size: 11px;
}

.info-value {
  color: var(--text-secondary);
  font-size: 13px;
}

.error-section {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: var(--radius-sm);
  padding: var(--spacing-sm);
  margin-bottom: var(--spacing-md);
}

.meta-error {
  color: var(--status-error);
  font-size: 12px;
  margin: 0;
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  
  & + .meta-error {
    margin-top: var(--spacing-xs);
  }
  
  .el-icon {
    font-size: 14px;
  }
}

.status-badge {
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 600;
  padding: 4px 8px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  
  .el-icon {
    font-size: 12px;
  }
}

.status-connected {
  background: rgba(16, 185, 129, 0.2);
  color: var(--status-success);
}

.status-connecting {
  background: rgba(245, 158, 11, 0.2);
  color: var(--status-warning);
}

.status-disconnected {
  background: var(--bg-secondary);
  color: var(--text-muted);
}

.status-error {
  background: rgba(239, 68, 68, 0.2);
  color: var(--status-error);
}

.card-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-sm);
}

.dialog-mask {
  align-items: center;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  inset: 0;
  justify-content: center;
  position: fixed;
  z-index: 1000;
  padding: var(--spacing-lg);
}

.dialog-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  max-width: 600px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--border-color);
  
  h3 {
    color: var(--text-primary);
    margin: 0;
    font-size: 18px;
  }
}

.btn-close {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: var(--spacing-xs);
  
  .el-icon {
    font-size: 20px;
  }
  
  &:hover {
    color: var(--text-primary);
  }
}

.form-grid {
  display: grid;
  gap: var(--spacing-md);
  grid-template-columns: repeat(2, 1fr);
  padding: var(--spacing-lg);
}

.form-full {
  grid-column: 1 / -1;
}

.form-grid label {
  color: var(--text-secondary);
  display: flex;
  flex-direction: column;
  font-size: 13px;
  gap: var(--spacing-xs);
}

.form-grid input,
.form-grid select {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  padding: var(--spacing-sm);
}

.form-error {
  color: var(--status-error);
  margin: 0 var(--spacing-lg) var(--spacing-md);
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: 14px;
  
  .el-icon {
    font-size: 16px;
  }
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-sm);
  padding: 0 var(--spacing-lg) var(--spacing-lg);
}

.error-banner {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  color: var(--status-error);
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  margin-top: var(--spacing-lg);
  
  .el-icon {
    font-size: 20px;
  }
}

@media (max-width: 1200px) {
  .metric-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 900px) {
  .panel-header {
    flex-direction: column;
  }
  
  .metric-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .policy-strip {
    flex-wrap: wrap;
    gap: var(--spacing-sm);
  }
  
  .filter-bar {
    flex-wrap: wrap;
  }
  
  .keyword-field {
    flex: 1 1 100%;
  }
  
  .device-board {
    grid-template-columns: 1fr;
  }
  
  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>