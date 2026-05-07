import { apiGet, apiPost } from './client'

/** 绑定计量设备和打压设备到会话 */
export async function bindDevices(measureDeviceId: string, pressureDeviceId: string): Promise<void> {
  await apiPost('/session/devices', { measureDeviceId, pressureDeviceId })
}

/** 仅绑定计量设备 */
export async function bindMeasureDevice(measureDeviceId: string): Promise<void> {
  await apiPost('/session/measure-device', { measureDeviceId })
}

/** 读取当前压力 */
export async function readPressure(): Promise<number> {
  return (await apiGet<{ pressure: number }>('/session/pressure')).pressure
}

/** 读取稳定状态 */
export async function readStability(): Promise<boolean> {
  return (await apiGet<{ stable: boolean }>('/session/stability')).stable
}

/** 读取计量设备实时数据 */
export async function readMeasureData(): Promise<number[]> {
  return (await apiGet<{ data: number[] }>('/session/measure-data')).data
}

/** 读取阀门状态 */
export async function readValveStatus(): Promise<string> {
  return (await apiGet<{ status: string }>('/session/valve')).status
}

/** 设置阀门状态 */
export async function setValveStatus(status: string): Promise<void> {
  await apiPost('/session/valve', { status })
}

/** 读取压力单位 */
export async function readMeasureUnit(): Promise<string> {
  return (await apiGet<{ unit: string }>('/session/measure-unit')).unit
}

/** 设置压力单位 */
export async function setMeasureUnit(unit: string): Promise<void> {
  await apiPost('/session/measure-unit', { unit })
}

/** 读取设备信息 */
export async function readDeviceInfo(): Promise<Record<string, string>> {
  return (await apiGet<{ info: Record<string, string> }>('/session/device-info')).info
}

/** 复位设备 */
export async function resetDevice(): Promise<void> {
  await apiPost('/session/reset')
}
