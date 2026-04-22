import type { ApiResponse } from '@/types/api'
import { requestJSON } from './client'

/** 绑定计量设备和打压设备到会话 */
export async function bindDevices(measureDeviceId: string, pressureDeviceId: string): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/session/devices', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ measureDeviceId, pressureDeviceId })
  })
}

/** 仅绑定计量设备 */
export async function bindMeasureDevice(measureDeviceId: string): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/session/measure-device', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ measureDeviceId })
  })
}

/** 读取当前压力 */
export async function readPressure(): Promise<number> {
  const resp = await requestJSON<ApiResponse<{ pressure: number }>>('/session/pressure')
  return resp.data.pressure
}

/** 读取稳定状态 */
export async function readStability(): Promise<boolean> {
  const resp = await requestJSON<ApiResponse<{ stable: boolean }>>('/session/stability')
  return resp.data.stable
}

/** 读取计量设备实时数据 */
export async function readMeasureData(): Promise<number[]> {
  const resp = await requestJSON<ApiResponse<{ data: number[] }>>('/session/measure-data')
  return resp.data.data
}

/** 读取阀门状态 */
export async function readValveStatus(): Promise<string> {
  const resp = await requestJSON<ApiResponse<{ status: string }>>('/session/valve')
  return resp.data.status
}

/** 设置阀门状态 */
export async function setValveStatus(status: string): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/session/valve', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status })
  })
}

/** 读取压力单位 */
export async function readMeasureUnit(): Promise<string> {
  const resp = await requestJSON<ApiResponse<{ unit: string }>>('/session/measure-unit')
  return resp.data.unit
}

/** 设置压力单位 */
export async function setMeasureUnit(unit: string): Promise<void> {
  await requestJSON<ApiResponse<{ unit: string }>>('/session/measure-unit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ unit })
  })
}

/** 读取设备信息 */
export async function readDeviceInfo(): Promise<Record<string, string>> {
  const resp = await requestJSON<ApiResponse<{ info: Record<string, string> }>>('/session/device-info')
  return resp.data.info
}

/** 复位设备 */
export async function resetDevice(): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/session/reset', {
    method: 'POST'
  })
}
