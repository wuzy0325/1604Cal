import { apiGet, apiPost } from './client'

export interface CollectedRow {
  timestamp: string
  channels: Record<string, number>
}

export interface MeasurementDataResponse {
  rows: CollectedRow[]
  total: number
}

export interface MeasurementParamsPayload {
  minPressure: number
  maxPressure: number
  pointCount: number
  precision: number
  averageCount: number
  stableDurationMs: number
  precisionLevel: number
  pressureMode: string
  controlMode: string
  customPoints?: number[]
}

/** 获取计量状态 */
export async function fetchMeasurementState(): Promise<string> {
  return (await apiGet<{ state: string }>('/measurement/state')).state
}

/** 启动计量采集 */
export async function startMeasurement(channels: number[]): Promise<string> {
  return (await apiPost<{ state: string }>('/measurement/start', { channels })).state
}

/** 暂停计量采集 */
export async function pauseMeasurement(): Promise<string> {
  return (await apiPost<{ state: string }>('/measurement/pause')).state
}

/** 停止计量采集 */
export async function stopMeasurement(): Promise<string> {
  return (await apiPost<{ state: string }>('/measurement/stop')).state
}

/** 获取采集数据 */
export async function fetchMeasurementData(): Promise<MeasurementDataResponse> {
  return apiGet<MeasurementDataResponse>('/measurement/data')
}

/** 导出 CSV 下载地址 */
export function getMeasurementExportUrl(): string {
  return '/api/v1/measurement/export?format=csv'
}

/** 获取计量模块参数配置 */
export async function getMeasurementParamsConfig(): Promise<MeasurementParamsPayload> {
  return apiGet<MeasurementParamsPayload>('/config/measurement')
}

/** 保存计量模块参数配置 */
export async function saveMeasurementParamsConfig(params: MeasurementParamsPayload): Promise<void> {
  await apiPost('/config/measurement', params)
}

export interface MeasurementPoint {
  id: string
  index: number
  targetPressure: number
  direction: string
  status: string
  collectedData?: number[]
  actualPressure?: number
  collectTime?: string
  errorMessage?: string
}

export interface MeasurementAlarmConfig {
  enabled: boolean
  enabledChannels: number[]
  confirmOnAlarm: boolean
  soundEnabled: boolean
  threshold: number
}

export interface MeasurementAlarm {
  pointId: string
  targetPressure: number
  actualPressure: number
  threshold: number
  maxDeviation: number
  overLimitChannels: number[]
}

/** 生成压力点 */
export async function generateMeasurementPoints(): Promise<MeasurementPoint[]> {
  return apiPost<MeasurementPoint[]>('/measurement/points/generate')
}

/** 获取压力点列表 */
export async function fetchMeasurementPoints(): Promise<MeasurementPoint[]> {
  return apiGet<MeasurementPoint[]>('/measurement/points')
}

/** 获取计量报警配置 */
export async function getMeasurementAlarmConfig(): Promise<MeasurementAlarmConfig> {
  return apiGet<MeasurementAlarmConfig>('/config/measurement-alarm')
}

/** 保存计量报警配置 */
export async function saveMeasurementAlarmConfig(config: MeasurementAlarmConfig): Promise<void> {
  await apiPost('/config/measurement-alarm', config)
}

/** 检查是否有待处理的报警 */
export async function checkMeasurementAlarmPending(): Promise<boolean> {
  return (await apiGet<{ pending: boolean }>('/measurement/alarm/pending')).pending
}

/** 处理报警 */
export async function resolveMeasurementAlarm(decision: 'continue' | 'retry'): Promise<void> {
  await apiPost('/measurement/alarm/resolve', { decision })
}

/** 自动按点采集（逐点打压→稳定→采集） */
export async function autoCollectMeasurement(): Promise<string> {
  return (await apiPost<{ state: string }>('/measurement/auto-collect')).state
}

/** 手动打压指定测点 */
export async function manualPressurizeMeasurement(pointIndex: number): Promise<string> {
  return (await apiPost<{ state: string }>('/measurement/manual-pressurize', { pointIndex })).state
}

/** 手动采集指定测点 */
export async function manualCollectMeasurement(pointIndex: number): Promise<string> {
  return (await apiPost<{ state: string }>('/measurement/manual-collect', { pointIndex })).state
}
