import type { ApiResponse } from '@/types/api'
import { requestJSON } from './client'

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
  const resp = await requestJSON<ApiResponse<{ state: string }>>('/measurement/state')
  return resp.data.state
}

/** 启动计量采集 */
export async function startMeasurement(channels: number[]): Promise<string> {
  const resp = await requestJSON<ApiResponse<{ state: string }>>('/measurement/start', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ channels })
  })
  return resp.data.state
}

/** 暂停计量采集 */
export async function pauseMeasurement(): Promise<string> {
  const resp = await requestJSON<ApiResponse<{ state: string }>>('/measurement/pause', {
    method: 'POST'
  })
  return resp.data.state
}

/** 停止计量采集 */
export async function stopMeasurement(): Promise<string> {
  const resp = await requestJSON<ApiResponse<{ state: string }>>('/measurement/stop', {
    method: 'POST'
  })
  return resp.data.state
}

/** 获取采集数据 */
export async function fetchMeasurementData(): Promise<MeasurementDataResponse> {
  const resp = await requestJSON<ApiResponse<MeasurementDataResponse>>('/measurement/data')
  return resp.data
}

/** 导出 CSV 下载地址 */
export function getMeasurementExportUrl(): string {
  return '/api/v1/measurement/export?format=csv'
}

/** 获取计量模块参数配置 */
export async function getMeasurementParamsConfig(): Promise<MeasurementParamsPayload> {
  const resp = await requestJSON<ApiResponse<MeasurementParamsPayload>>('/config/measurement')
  return resp.data
}

/** 保存计量模块参数配置 */
export async function saveMeasurementParamsConfig(params: MeasurementParamsPayload): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/config/measurement', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params)
  })
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
  const resp = await requestJSON<ApiResponse<MeasurementPoint[]>>('/measurement/points/generate', {
    method: 'POST'
  })
  return resp.data
}

/** 获取压力点列表 */
export async function fetchMeasurementPoints(): Promise<MeasurementPoint[]> {
  const resp = await requestJSON<ApiResponse<MeasurementPoint[]>>('/measurement/points')
  return resp.data
}

/** 获取计量报警配置 */
export async function getMeasurementAlarmConfig(): Promise<MeasurementAlarmConfig> {
  const resp = await requestJSON<ApiResponse<MeasurementAlarmConfig>>('/config/measurement-alarm')
  return resp.data
}

/** 保存计量报警配置 */
export async function saveMeasurementAlarmConfig(config: MeasurementAlarmConfig): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/config/measurement-alarm', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config)
  })
}

/** 检查是否有待处理的报警 */
export async function checkMeasurementAlarmPending(): Promise<boolean> {
  const resp = await requestJSON<ApiResponse<{ pending: boolean }>>('/measurement/alarm/pending')
  return resp.data.pending
}

/** 处理报警 */
export async function resolveMeasurementAlarm(decision: 'continue' | 'retry'): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/measurement/alarm/resolve', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ decision })
  })
}

/** 自动按点采集（逐点打压→稳定→采集） */
export async function autoCollectMeasurement(): Promise<string> {
  const resp = await requestJSON<ApiResponse<{ state: string }>>('/measurement/auto-collect', {
    method: 'POST'
  })
  return resp.data.state
}

/** 手动打压指定测点 */
export async function manualPressurizeMeasurement(pointIndex: number): Promise<string> {
  const resp = await requestJSON<ApiResponse<{ state: string }>>('/measurement/manual-pressurize', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pointIndex })
  })
  return resp.data.state
}

/** 手动采集指定测点 */
export async function manualCollectMeasurement(pointIndex: number): Promise<string> {
  const resp = await requestJSON<ApiResponse<{ state: string }>>('/measurement/manual-collect', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pointIndex })
  })
  return resp.data.state
}
