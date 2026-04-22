import type { ApiResponse } from '@/types/api'
import type { MultiPressDeviceState } from '@/types/multipress'
import { requestJSON } from './client'

/** 注册打压设备到多设备控制模块 */
export async function multipressRegister(deviceId: string): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/multipress/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ deviceId })
  })
}

/** 注销打压设备 */
export async function multipressUnregister(deviceId: string): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/multipress/unregister', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ deviceId })
  })
}

/** 设置目标压力 */
export async function multipressSetPressure(deviceId: string, targetPressure: number): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/multipress/set-pressure', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ deviceId, targetPressure })
  })
}

/** 停止打压 */
export async function multipressStop(deviceId: string): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/multipress/stop', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ deviceId })
  })
}

/** 排空压力 */
export async function multipressExhaust(deviceId: string): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/multipress/exhaust', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ deviceId })
  })
}

/** 读取指定设备当前压力 */
export async function multipressReadPressure(deviceId: string): Promise<number> {
  const resp = await requestJSON<ApiResponse<{ pressure: number; deviceId: string }>>(
    `/multipress/pressure?deviceId=${encodeURIComponent(deviceId)}`
  )
  return resp.data.pressure
}

/** 读取指定设备稳定状态 */
export async function multipressReadStability(deviceId: string): Promise<boolean> {
  const resp = await requestJSON<ApiResponse<{ stable: boolean; deviceId: string }>>(
    `/multipress/stability?deviceId=${encodeURIComponent(deviceId)}`
  )
  return resp.data.stable
}

/** 设置指定设备压力单位 */
export async function multipressSetUnit(deviceId: string, unit: string): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/multipress/unit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ deviceId, unit })
  })
}

/** 获取所有已注册设备状态 */
export async function multipressListDevices(): Promise<MultiPressDeviceState[]> {
  const resp = await requestJSON<ApiResponse<MultiPressDeviceState[]>>('/multipress/devices')
  return resp.data ?? []
}

/** 紧急停止所有设备 */
export async function multipressStopAll(): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/multipress/stop-all', {
    method: 'POST'
  })
}
