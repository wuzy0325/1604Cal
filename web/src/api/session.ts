import { apiGet, apiPost } from './client'

/** 绑定计量设备（支持多台）和打压设备到会话 */
export async function bindDevices(
  measureDeviceId: string | string[],
  pressureDeviceId: string,
  moduleName = 'measurement'
): Promise<void> {
  const ids = Array.isArray(measureDeviceId) ? measureDeviceId : [measureDeviceId]
  await apiPost('/session/devices', {
    measureDeviceId: ids[0] ?? '',
    measureDeviceIds: ids,
    pressureDeviceId,
    moduleName
  })
}

/** 绑定多台计量设备与打压设备到会话 */
export async function bindMeasureDevices(
  measureDeviceIds: string[],
  pressureDeviceId: string,
  moduleName = 'measurement'
): Promise<void> {
  await apiPost('/session/devices', { measureDeviceIds, pressureDeviceId, moduleName })
}

/** 仅绑定计量设备（保留当前打压设备绑定）；支持多设备 */
export async function bindMeasureDevice(
  measureDeviceId: string | string[],
  moduleName = 'measurement'
): Promise<void> {
  const ids = Array.isArray(measureDeviceId) ? measureDeviceId : [measureDeviceId]
  await apiPost('/session/measure-device', {
    measureDeviceId: ids[0] ?? '',
    measureDeviceIds: ids,
    moduleName
  })
}

/** 读取当前压力 */
export async function readPressure(): Promise<number> {
  return (await apiGet<{ pressure: number }>('/session/pressure')).pressure
}

/** 读取稳定状态 */
export async function readStability(): Promise<boolean> {
  return (await apiGet<{ stable: boolean }>('/session/stability')).stable
}

/** 读取计量设备实时数据（首个绑定设备，兼容单设备场景） */
export async function readMeasureData(): Promise<number[]> {
  return (await apiGet<{ data: number[] }>('/session/measure-data')).data
}

/** 读取所有已绑定计量设备的实时数据（deviceID -> 通道数据），供多设备展示 */
export async function readMeasureDataAllDevices(): Promise<Record<string, number[]>> {
  return (await apiGet<{ data: number[]; devices: Record<string, number[]> }>('/session/measure-data')).devices
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

/** 对指定通道执行校零，返回各通道校零偏移；偏移会持久化到本地并在重连后自动应用 */
export async function calibrateZero(channels: number[]): Promise<number[]> {
  return (await apiPost<{ data: number[] }>('/session/calibrate-zero', { channels })).data
}

/** 复位设备 */
export async function resetDevice(): Promise<void> {
  await apiPost('/session/reset')
}
