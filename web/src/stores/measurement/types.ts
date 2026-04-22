/** 计量模块采集状态 */
export type MeasurementState = 'idle' | 'collecting' | 'paused'

/** 单次采集的数据行 */
export interface CollectedRow {
  timestamp: string
  channels: Record<string, number>
}
