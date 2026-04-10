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
  unit: string
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

export interface SessionStateDTO {
  state: string
}

export interface ReportTemplateDTO {
  filename: string
}

export interface StreamEventPayload {
  type: string
  data?: unknown
}

const API_BASE = '/api/v1'

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
  return resp.data
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
