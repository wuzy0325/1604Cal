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
