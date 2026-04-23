/** 计量模块采集状态 */
export type MeasurementState =
  | 'idle'
  | 'pressuring'
  | 'stabilizing'
  | 'collecting'
  | 'completed'
  | 'error'
  | 'paused'

/** 单次采集的数据行 */
export interface CollectedRow {
  timestamp: string
  channels: Record<string, number>
}
