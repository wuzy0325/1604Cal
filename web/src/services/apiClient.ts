export interface HealthResponse {
  status: string
}

export interface ApiResponse<T> {
  success: boolean
  code?: string
  message?: string
  data: T
}

export interface DeviceDTO {
  id: string
  name: string
  type: 'measure' | 'pressure'
  model: string
  host: string
  port: number
  unit?: string
  status: 'disconnected' | 'connecting' | 'connected' | 'error'
  lastErrorReason?: string
  lastErrorAt?: string
}

export interface DeviceStatusChangedEventData {
  id: string
  type?: DeviceDTO['type']
  status?: DeviceDTO['status']
  errorReason?: string
  lastErrorAt?: string
}

export interface UnitConsistencyDTO {
  consistent: boolean
  conflicts: string[]
}

export interface DeviceConnectConfigDTO {
  connectAttemptTimeoutMs: number
  connectMaxAttempts: number
  connectInitialBackoffMs: number
  connectMaxBackoffMs: number
  disconnectAttemptTimeoutMs: number
  disconnectMaxAttempts: number
  disconnectInitialBackoffMs: number
  disconnectMaxBackoffMs: number
}

// 会话状态 - 与后端 domain.SessionState 完全对齐
export type SessionState =
  | 'idle'
  | 'ready'
  | 'pressurizing'
  | 'stabilizing'
  | 'collecting'
  | 'point_done'
  | 'fitting'
  | 'completed'
  | 'paused'
  | 'stopped'
  | 'await_manual_collect'
  | 'await_alarm_resolution'
  | 'recovering'
  | 'error'

export interface SessionStateDTO {
  state: SessionState
}

export interface ReportTemplateDTO {
  filename: string
}

export interface StreamEventPayload {
  type: string
  data?: unknown
}

// 校准相关 DTO
export interface CalibrationConfigDTO {
  channels: number[]
  pressurePoints: number
  averageCount: number
  minPressure: number
  maxPressure: number
  stableWaitMs: number
  controlMode?: 'auto' | 'manual'
  pressureMode?: 'single' | 'return'
}

export interface PressurePointDTO {
  index: number
  targetPressure: number
  status: string
  collectedData?: number[]
  actualPressure?: number
}

export interface SetDevicesRequest {
  measureDeviceId: string
  pressureDeviceId: string
}

// ---------------------------------------------------------------------------
// API 基础路径：桌面模式下指向内嵌 HTTP 服务器，Web 模式下使用相对路径。
// ---------------------------------------------------------------------------
let API_BASE = '/api/v1'

/**
 * 初始化桌面模式下的 API 基础路径。
 * 在 Wails 桌面环境中，通过 Wails 绑定获取内嵌 HTTP 服务器的端口。
 */
export async function initDesktopApiBase(): Promise<void> {
  const w = window as unknown as { go?: { main?: { App?: { GetAPIPort: () => Promise<number> } } } }
  if (typeof window !== 'undefined' && w.go?.main?.App) {
    try {
      const port: number = await w.go.main.App.GetAPIPort()
      API_BASE = `http://127.0.0.1:${port}/api/v1`
    } catch (e) {
      console.warn('Failed to detect Wails API port, falling back to relative path:', e)
    }
    return
  }

  // E2E 测试环境：前端独立运行在 4173 端口时，直接指向后端 API
  if (
    typeof window !== 'undefined' &&
    window.location.hostname === 'localhost' &&
    window.location.port === '4173'
  ) {
    API_BASE = 'http://localhost:18080/api/v1'
  }
}

/** 返回当前 API 基础路径。 */
export function getApiBase(): string {
  return API_BASE
}

export async function fetchHealth(): Promise<HealthResponse> {
  const resp = await fetch(`${API_BASE}/health`)
  if (!resp.ok) {
    throw new Error(`health request failed: ${resp.status}`)
  }

  return (await resp.json()) as HealthResponse
}

async function requestJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const resp = await fetch(`${API_BASE}${path}`, init)
  if (!resp.ok) {
    throw new Error(`request failed: ${resp.status}`)
  }

  return (await resp.json()) as T
}

export async function fetchDevices(): Promise<DeviceDTO[]> {
  const resp = await requestJSON<ApiResponse<DeviceDTO[]>>('/devices')
  return resp.data ?? []
}

export async function upsertDevice(payload: DeviceDTO): Promise<DeviceDTO> {
  const resp = await requestJSON<ApiResponse<DeviceDTO>>('/devices', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })

  return resp.data
}

export async function connectDevice(id: string): Promise<DeviceDTO> {
  const resp = await requestJSON<ApiResponse<DeviceDTO>>('/devices/connect', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ id })
  })

  return resp.data
}

export async function disconnectDevice(id: string): Promise<DeviceDTO> {
  const resp = await requestJSON<ApiResponse<DeviceDTO>>('/devices/disconnect', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ id })
  })

  return resp.data
}

export async function setDeviceStatus(
  id: string,
  status: DeviceDTO['status']
): Promise<{ id: string; status: DeviceDTO['status'] }> {
  const resp = await requestJSON<ApiResponse<{ id: string; status: DeviceDTO['status'] }>>('/devices/status', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ id, status })
  })

  return resp.data
}

export async function fetchUnitConsistency(): Promise<UnitConsistencyDTO> {
  const resp = await requestJSON<ApiResponse<UnitConsistencyDTO>>('/checks/unit-consistency')
  return resp.data
}

export async function fetchDeviceConnectConfig(): Promise<DeviceConnectConfigDTO> {
  const resp = await requestJSON<ApiResponse<DeviceConnectConfigDTO>>('/config/device-connect')
  return resp.data
}

export async function fetchSessionState(): Promise<SessionStateDTO> {
  const resp = await requestJSON<ApiResponse<SessionStateDTO>>('/sessions/state')
  return resp.data
}

export async function triggerSessionAction(action: 'start' | 'pause' | 'resume' | 'stop'): Promise<SessionStateDTO> {
  const resp = await requestJSON<ApiResponse<SessionStateDTO>>(`/sessions/${action}`, {
    method: 'POST'
  })
  return resp.data
}

export async function selectReportTemplate(points: number, mode: 'single' | 'return'): Promise<ReportTemplateDTO> {
  const resp = await requestJSON<ApiResponse<ReportTemplateDTO>>(
    `/reports/templates/select?points=${points}&mode=${mode}`
  )
  return resp.data
}

export function createEventStream(onEvent: (payload: StreamEventPayload) => void): EventSource {
  const source = new EventSource(`${API_BASE}/events/stream`)

  source.onmessage = (event) => {
    try {
      const payload = JSON.parse(event.data) as StreamEventPayload
      onEvent(payload)
    } catch {
      // 忽略解析失败的事件，避免影响后续消息处理。
    }
  }

  return source
}

// ---------------------------------------------------------------------------
// 校准流程 API
// ---------------------------------------------------------------------------

/** 设置校准使用的设备 */
export async function setCalibrationDevices(payload: SetDevicesRequest): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/calibration/devices', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  })
}

/** 仅设置计量设备（连接后立即绑定驱动，用于读取阀门/单位/设备信息） */
export async function setCalibrationMeasureDevice(measureDeviceId: string): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/calibration/measure-device', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ measureDeviceId })
  })
}

/** 设置校准配置 */
export async function setCalibrationConfig(config: CalibrationConfigDTO): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/calibration/config', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config)
  })
}

/** 设置采集通道 */
export async function setCalibrationChannels(channels: number[]): Promise<number[]> {
  const resp = await requestJSON<ApiResponse<{ channels: number[] }>>('/calibration/channels', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ channels })
  })
  return resp.data.channels
}

/** 获取当前通道配置 */
export async function getCalibrationChannels(): Promise<number[]> {
  const resp = await requestJSON<ApiResponse<{ channels: number[] }>>('/calibration/channels/list')
  return resp.data.channels
}

/** 生成压力点 */
export async function generatePressurePoints(): Promise<PressurePointDTO[]> {
  const resp = await requestJSON<ApiResponse<PressurePointDTO[]>>('/calibration/points/generate', {
    method: 'POST'
  })
  return resp.data
}

/** 获取压力点列表 */
export async function getPressurePoints(): Promise<PressurePointDTO[]> {
  const resp = await requestJSON<ApiResponse<PressurePointDTO[]>>('/calibration/points')
  return resp.data
}

/** 执行打压 */
export async function pressurize(pointIndex: number): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/calibration/pressurize', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pointIndex })
  })
}

/** 采集数据 */
export async function collectData(pointIndex: number): Promise<number[]> {
  const resp = await requestJSON<ApiResponse<{ data: number[] }>>('/calibration/collect', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pointIndex })
  })
  return resp.data.data
}

/** 执行拟合 */
export async function fitData(): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/calibration/fit', {
    method: 'POST'
  })
}

/** 读取当前压力 */
export async function readCurrentPressure(): Promise<number> {
  const resp = await requestJSON<ApiResponse<{ pressure: number }>>('/calibration/pressure')
  return resp.data.pressure
}

/** 读取稳定状态 */
export async function readStability(): Promise<boolean> {
  const resp = await requestJSON<ApiResponse<{ stable: boolean }>>('/calibration/stability')
  return resp.data.stable
}

/** 读取计量设备实时数据 */
export async function readMeasureData(): Promise<number[]> {
  const resp = await requestJSON<ApiResponse<{ data: number[] }>>('/calibration/measure-data')
  return resp.data.data
}

/** 读取计量设备阀门状态 */
export async function readCalibrationValve(): Promise<string> {
  const resp = await requestJSON<ApiResponse<{ status: string }>>('/calibration/valve')
  return resp.data.status
}

/** 设置计量设备阀门状态 */
export async function setCalibrationValve(status: string): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/calibration/valve', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status })
  })
}

/** 读取计量设备压力单位 */
export async function readCalibrationMeasureUnit(): Promise<string> {
  const resp = await requestJSON<ApiResponse<{ unit: string }>>('/calibration/measure-unit')
  return resp.data.unit
}

/** 设置计量设备压力单位 */
export async function setCalibrationMeasureUnit(unit: string): Promise<void> {
  await requestJSON<ApiResponse<{ unit: string }>>('/calibration/measure-unit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ unit })
  })
}

/** 读取计量设备信息 */
export async function readCalibrationDeviceInfo(): Promise<Record<string, string>> {
  const resp = await requestJSON<ApiResponse<{ info: Record<string, string> }>>('/calibration/device-info')
  return resp.data.info
}

/** 复位计量设备 */
export async function resetCalibrationDevice(): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/calibration/reset', {
    method: 'POST'
  })
}

// ---------------------------------------------------------------------------
// 多设备打压控制 API
// ---------------------------------------------------------------------------

/** 多设备打压设备运行状态 */
export interface MultiPressDeviceState {
  deviceId: string
  currentPressure: number
  targetPressure: number
  unit: string
  stable: boolean
  status: 'idle' | 'pressurizing' | 'exhausting' | 'error'
  errorMessage?: string
}

/** 注册打压设备到多设备控制模块 */
export async function multipressRegister(deviceId: string): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/multipress/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ deviceId })
  })
}

/** 注销打压设备 */
export async function multipressUnregister(deviceId: string): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/multipress/unregister', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ deviceId })
  })
}

/** 设置目标压力 */
export async function multipressSetPressure(deviceId: string, targetPressure: number): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/multipress/set-pressure', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ deviceId, targetPressure })
  })
}

/** 停止打压 */
export async function multipressStop(deviceId: string): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/multipress/stop', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ deviceId })
  })
}

/** 排空压力 */
export async function multipressExhaust(deviceId: string): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/multipress/exhaust', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ deviceId })
  })
}

/** 读取指定设备当前压力 */
export async function multipressReadPressure(deviceId: string): Promise<number> {
  const resp = await requestJSON<ApiResponse<{ pressure: number; deviceId: string }>>(
    `/multipress/pressure?deviceId=${encodeURIComponent(deviceId)}`
  )
  return resp.data.pressure
}

/** 读取指定设备稳定状态 */
export async function multipressReadStability(deviceId: string): Promise<boolean> {
  const resp = await requestJSON<ApiResponse<{ stable: boolean; deviceId: string }>>(
    `/multipress/stability?deviceId=${encodeURIComponent(deviceId)}`
  )
  return resp.data.stable
}

/** 设置指定设备压力单位 */
export async function multipressSetUnit(deviceId: string, unit: string): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/multipress/unit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ deviceId, unit })
  })
}

/** 获取所有已注册设备状态 */
export async function multipressListDevices(): Promise<MultiPressDeviceState[]> {
  const resp = await requestJSON<ApiResponse<MultiPressDeviceState[]>>('/multipress/devices')
  return resp.data ?? []
}

/** 紧急停止所有设备 */
export async function multipressStopAll(): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/multipress/stop-all', {
    method: 'POST'
  })
}
