/** 计量模块采集状态 */
export type MeasurementState =
  | 'idle'
  | 'ready'
  | 'pressurizing'
  | 'stabilizing'
  | 'collecting'
  | 'completed'
  | 'error'
  | 'paused'
  | 'stopped'

/** 单次采集的数据行 */
export interface CollectedRow {
  timestamp: string
  channels: Record<string, number>
}

/** 稳定性监控更新载荷 */
export interface StabilityUpdate {
  pointIndex: number
  isStable: boolean
  isInRange: boolean
  currentValue: number
  stableDurationMs: number
  requiredDurationMs: number
  progress: number
}

/** 报警数据载荷 */
export interface AlarmData {
  pointId: string
  targetPressure: number
  actualPressure: number
  threshold: number
  maxDeviation: number
  overLimitChannels: number[]
}
