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
          <button
            type="button"
            class="btn btn-delete"
            @click="handleDeleteDevice(device)"
          >
            <el-icon><Delete /></el-icon>
            删除
          </button>
        </div>
      </article>
    </section>

    <el-dialog
      v-model="showConnectDialog"
      title="设备连接中"
      width="400px"
      :close-on-click-modal="false"
      :show-close="false"
      :close-on-press-escape="false"
    >
      <div style="text-align:center;padding:20px 0">
        <el-icon style="font-size:36px;margin-bottom:12px" class="is-loading"><Loading /></el-icon>
        <p style="margin:8px 0 0;color:#666;font-size:14px">{{ connectProgressMessage }}</p>
      </div>
    </el-dialog>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新增设备配置' : '编辑设备配置'"
      width="520px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div class="form-grid">
        <label v-if="dialogMode === 'edit'">
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
            <option value="" disabled>选择型号</option>
            <option v-for="m in modelOptions" :key="m.value" :value="m.value">
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
        v-if="formError"
        class="form-error"
      >
        <el-icon><Warning /></el-icon>
        {{ formError }}
      </p>

      <template #footer>
        <el-button @click="closeDialog">
          取消
        </el-button>
        <el-button
          data-test="submit-form"
          type="primary"
          @click="submitForm"
        >
          <el-icon><Check /></el-icon>
          保存
        </el-button>
      </template>
    </el-dialog>

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
  Check,
  Delete
} from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'

import {
  connectDevice,
  deleteDevice,
  disconnectDevice,
  fetchDeviceConnectConfig,
  fetchDevices,
  fetchUnitConsistency,
  upsertDevice
} from "@/api/device"
import {
  multipressRegister,
  multipressUnregister
} from "@/api/multipress"
import { createEventStream } from "@/api/client"
import { EVENT_DEVICE_CONNECT_PROGRESS } from "@/shared/events"
import type {
  DeviceConnectConfigDTO,
  DeviceDTO,
  DeviceStatusChangedEventData
} from "@/types/device"
import type { StreamEventPayload } from "@/types/api"

type DeviceFilterType = 'all' | DeviceDTO['type']
type DeviceStatusFilter = 'all' | DeviceDTO['status']

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

const showConnectDialog = ref(false)
const connectingDeviceId = ref<string | null>(null)
const connectProgressMessage = ref('')
let connectTimeoutTimer: ReturnType<typeof setTimeout> | null = null

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
  if (dialogMode.value === 'create') {
    if (form.type === 'measure') {
      form.model = 'WTN1604'
    } else {
      form.model = ''
    }
  }
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
  form.id = generateDeviceId()
  form.name = ''
  form.type = 'measure'
  form.model = ''
  form.host = ''
  form.port = 9000
  form.localAddr = ''
  form.status = 'disconnected'
  formError.value = ''
  dialogVisible.value = true
}

/** 生成唯一设备ID，格式 dev-{timestamp后6位}-{随机4位} */
function generateDeviceId(): string {
  const ts = Date.now().toString().slice(-6)
  const rand = Math.random().toString(36).slice(2, 6)
  return `dev-${ts}-${rand}`
}

function openEditDialog(device: DeviceDTO) {
  dialogMode.value = 'edit'
  form.id = device.id
  form.name = device.name
  form.type = device.type
  form.model = device.model
  form.host = device.host
  form.port = device.port
  form.localAddr = device.localAddr ?? ''
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
      localAddr: form.localAddr || undefined,
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

  if (device.status === 'connecting') {
    errorMessage.value = '设备正在连接中，请稍候'
    return
  }

  try {
    if (device.status === 'connected') {
      if (device.type === 'pressure') {
        await multipressUnregister(device.id)
      } else {
        await disconnectDevice(device.id)
      }
    } else {
      device.status = 'connecting'
      connectingDeviceId.value = device.id
      connectProgressMessage.value = '准备连接...'
      showConnectDialog.value = true

      // 前端超时兜底
      connectTimeoutTimer = setTimeout(() => {
        if (showConnectDialog.value) {
          showConnectDialog.value = false
          connectingDeviceId.value = null
          errorMessage.value = '连接超时，请检查设备网络和地址'
        }
      }, 12000)

      if (device.type === 'pressure') {
        await multipressRegister(device.id)
      } else {
        const result = await connectDevice(device.id)
        if (result.status === 'error') {
          errorMessage.value = result.lastErrorReason || '连接失败，请检查设备地址和网络'
        }
      }
    }
    await refreshAll()
  } catch (error) {
    device.status = 'disconnected'
    errorMessage.value = error instanceof Error ? error.message : '切换连接状态失败'
  } finally {
    showConnectDialog.value = false
    connectingDeviceId.value = null
    if (connectTimeoutTimer) {
      clearTimeout(connectTimeoutTimer)
      connectTimeoutTimer = null
    }
  }
}

async function handleDeleteDevice(device: DeviceDTO) {
  errorMessage.value = ''
  try {
    await ElMessageBox.confirm(
      `确定要删除设备「${device.name || device.id}」吗？此操作不可撤销。`,
      '确认删除',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
    )
    await deleteDevice(device.id)
    await refreshAll()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      errorMessage.value = error instanceof Error ? error.message : '删除设备失败'
    }
  }
}

function validateForm() {
  if (!form.host) {
    return '请填写IP地址。'
  }

  if (dialogMode.value === 'create' && devices.value.some((item) => item.id === form.id)) {
    form.id = generateDeviceId()
  }

  if (!isValidIPv4(form.host)) {
    return 'IP地址格式不正确'
  }

  if (!Number.isInteger(form.port) || form.port < 1 || form.port > 65535) {
    return '端口必须在1-65535之间'
  }

  if (!form.model) {
    return '请选择设备型号'
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

  eventSource = createEventStream({
    onEvent: (payload: StreamEventPayload) => {
      if (payload.type === 'device.status.changed') {
        if (isDeviceStatusChangedEventData(payload.data)) {
          applyDeviceStatusChangedEvent(payload.data)
        } else {
          void refreshAll()
        }
      }
      if (payload.type === EVENT_DEVICE_CONNECT_PROGRESS) {
        const data = payload.data as { deviceId?: string; message?: string }
        if (data.deviceId && data.message) {
          connectProgressMessage.value = data.message
        }
      }
    },
    onError: (error) => {
      console.warn('[DeviceManagementPanel] SSE 连接断开:', error)
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
  if (connectTimeoutTimer) {
    clearTimeout(connectTimeoutTimer)
    connectTimeoutTimer = null
  }
})
</script>

<style scoped lang="scss">
/* ── 设计系统令牌 ── */
$font-sans: 'Inter', 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
$font-mono: 'JetBrains Mono', 'Fira Code', 'SF Mono', Consolas, monospace;
$mint: #10b981;
$mint-dark: #059669;
$slate-50: #f9fafb;
$slate-100: #f3f4f6;
$slate-200: #e5e7eb;
$slate-300: #d1d5db;
$slate-400: #9ca3af;
$slate-500: #6b7280;
$slate-600: #4b5563;
$slate-700: #374151;
$slate-800: #1f2937;
$green: #22c55e;
$red: #ef4444;
$blue: #3b82f6;
$amber: #f59e0b;

.device-panel {
  background: #ffffff;
  border: 1px solid $slate-200;
  border-radius: 12px;
  padding: 16px;
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  font-family: $font-sans;
}

.panel-header {
  align-items: flex-start;
  display: flex;
  gap: 12px;
  justify-content: space-between;
  margin-bottom: 16px;
  flex-shrink: 0;
}

.panel-header h2 {
  color: $slate-800;
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.panel-header p {
  color: $slate-500;
  margin: 4px 0 0;
  font-size: 12px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border: 1px solid $slate-200;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  background: transparent;
  font-family: $font-sans;

  .el-icon {
    font-size: 14px;
  }
}

.btn-primary {
  background: linear-gradient(135deg, $mint, $mint-dark);
  color: #ffffff;
  border-color: transparent;

  &:hover {
    background: linear-gradient(135deg, #34d399, $mint);
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.25);
  }
}

.btn-success {
  background: $green;
  color: #ffffff;
  border-color: $green;

  &:hover {
    background: #16a34a;
  }
}

.btn-danger {
  background: $red;
  color: #fff;
  border-color: $red;

  &:hover {
    background: #dc2626;
  }
}

.btn-ghost {
  color: $slate-600;
  background: $slate-50;

  &:hover {
    background: $slate-100;
    color: $slate-800;
    border-color: $slate-300;
  }
}

.btn-delete {
  color: $red;
  border-color: rgba(239, 68, 68, 0.2);

  &:hover {
    background: rgba(239, 68, 68, 0.08);
    border-color: rgba(239, 68, 68, 0.35);
  }
}

.metric-grid {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(6, 1fr);
  margin-bottom: 12px;
  flex-shrink: 0;
}

.metric-card {
  background: $slate-50;
  border: 1px solid $slate-200;
  border-radius: 8px;
  padding: 10px 12px;
  text-align: center;
}

.metric-label {
  color: $slate-500;
  font-size: 11px;
  margin: 0 0 4px;
  font-weight: 500;
  letter-spacing: 0.05em;
}

.metric-value {
  color: $slate-800;
  display: block;
  font-size: 18px;
  font-weight: 600;
  font-family: $font-mono;

  &.success {
    color: $green;
  }

  &.danger {
    color: $red;
  }

  &.warning {
    color: $amber;
  }
}

.policy-strip {
  background: $slate-50;
  border: 1px solid $slate-200;
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 12px;
  padding: 8px 12px;
  flex-wrap: wrap;
  flex-shrink: 0;
}

.policy-item {
  display: flex;
  align-items: center;
  gap: 6px;
  color: $slate-500;
  font-size: 12px;

  .el-icon {
    color: $mint;
    font-size: 13px;
  }
}

.auto-refresh {
  display: flex;
  align-items: center;
  gap: 6px;
  color: $slate-500;
  font-size: 12px;
  margin-left: auto;
  cursor: pointer;
  font-weight: 500;

  input[type="checkbox"] {
    width: 14px;
    height: 14px;
    accent-color: $mint;
  }
}

.filter-bar {
  background: $slate-50;
  border: 1px solid $slate-200;
  border-radius: 8px;
  display: flex;
  align-items: flex-end;
  gap: 12px;
  margin-bottom: 12px;
  padding: 10px 12px;
  flex-shrink: 0;
}

.filter-bar label {
  color: $slate-500;
  display: flex;
  flex-direction: column;
  font-size: 12px;
  gap: 4px;
  flex: 1;
  font-weight: 500;
}

.keyword-field {
  flex: 2;
}

.filter-bar select,
.filter-bar input {
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

.device-board {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(2, 1fr);
  flex: 1;
  min-height: 0;
  overflow: auto;
  align-content: start;
  padding-right: 4px;

  &::-webkit-scrollbar { width: 4px; }
  &::-webkit-scrollbar-thumb { background: $slate-300; border-radius: 4px; }
  &::-webkit-scrollbar-track { background: transparent; }
}

.empty {
  grid-column: 1 / -1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: $slate-400;
  padding: 40px 0;

  .el-icon {
    font-size: 18px;
  }
}

.device-card {
  background: #ffffff;
  border: 1px solid $slate-200;
  border-radius: 10px;
  padding: 14px;
  transition: all 0.2s ease;

  &:hover {
    border-color: $slate-300;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.07), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
  }
}

.card-top {
  align-items: center;
  display: flex;
  gap: 12px;
  justify-content: space-between;
  margin-bottom: 12px;
}

.title-block {
  align-items: center;
  display: flex;
  gap: 8px;
}

.title-block strong {
  color: $slate-800;
  font-size: 14px;
  font-weight: 600;
}

.type-badge {
  background: $slate-50;
  border-radius: 999px;
  font-size: 10px;
  font-weight: 600;
  padding: 2px 8px;
  border: 1px solid $slate-200;

  &.measure {
    color: $blue;
    background: rgba(59, 130, 246, 0.08);
    border-color: rgba(59, 130, 246, 0.15);
  }

  &.pressure {
    color: $amber;
    background: rgba(245, 158, 11, 0.08);
    border-color: rgba(245, 158, 11, 0.15);
  }
}

.card-grid {
  display: grid;
  gap: 10px;
  grid-template-columns: 1fr 1fr;
  margin-bottom: 12px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.info-label {
  color: $slate-400;
  font-size: 10px;
  font-weight: 500;
  letter-spacing: 0.05em;
}

.info-value {
  color: $slate-600;
  font-size: 12px;
  font-weight: 500;
}

.error-section {
  background: rgba(239, 68, 68, 0.06);
  border: 1px solid rgba(239, 68, 68, 0.15);
  border-radius: 6px;
  padding: 8px 10px;
  margin-bottom: 12px;
}

.meta-error {
  color: $red;
  font-size: 11px;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 4px;
  font-weight: 500;

  & + .meta-error {
    margin-top: 4px;
  }

  .el-icon {
    font-size: 12px;
  }
}

.status-badge {
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  padding: 3px 8px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  line-height: 1.5;

  .el-icon {
    font-size: 10px;
  }
}

.status-connected {
  background: rgba(16, 185, 129, 0.12);
  border: 1px solid rgba(16, 185, 129, 0.25);
  color: #059669;
}

.status-connecting {
  background: rgba(245, 158, 11, 0.12);
  border: 1px solid rgba(245, 158, 11, 0.25);
  color: #d97706;
}

.status-disconnected {
  background: rgba(107, 114, 128, 0.12);
  border: 1px solid rgba(107, 114, 128, 0.25);
  color: $slate-500;
}

.status-error {
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.25);
  color: #dc2626;
}

.card-actions {
  display: flex;
  justify-content: flex-end;
  gap: 6px;
}

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

.error-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #dc2626;
  background: rgba(239, 68, 68, 0.06);
  border: 1px solid rgba(239, 68, 68, 0.15);
  border-radius: 8px;
  padding: 10px 14px;
  margin-top: 12px;
  flex-shrink: 0;
  font-weight: 500;
  font-size: 13px;

  .el-icon {
    font-size: 18px;
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
    gap: 8px;
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
