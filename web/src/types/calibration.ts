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

// 校准相关 DTO
export interface CalibrationConfigDTO {
  channels: number[]
  pressurePoints: number
  averageCount: number
  minPressure: number
  maxPressure: number
  stableWaitMs: number
  controlMode?: 'auto' | 'manual'
  pressureMode?: 'single' | 'roundTrip'
  precision?: number
  precisionLevel?: number
}

export interface PressurePointDTO {
  index: number
  targetPressure: number
  status: string
  direction?: 'forward' | 'backward'
  collectedData?: number[]
  actualPressure?: number
}

export interface FittingResultDTO {
  slope: number
  intercept: number
  r2: number
  points: number
}
