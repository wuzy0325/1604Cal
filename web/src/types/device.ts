export interface DeviceDTO {
  id: string
  name: string
  type: 'measure' | 'pressure'
  model: string
  host: string
  port: number
  unit?: string
  status: 'disconnected' | 'connecting' | 'connected' | 'error'
  localAddr?: string
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

export interface SetDevicesRequest {
  measureDeviceId: string
  pressureDeviceId: string
}
