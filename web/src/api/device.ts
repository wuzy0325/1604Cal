import type { ApiResponse } from '@/types/api'
import type { DeviceDTO, UnitConsistencyDTO, DeviceConnectConfigDTO } from '@/types/device'
import { requestJSON } from './client'

export async function fetchDevices(): Promise<DeviceDTO[]> {
  const resp = await requestJSON<ApiResponse<DeviceDTO[]>>('/devices')
  return resp.data ?? []
}

export async function upsertDevice(payload: DeviceDTO): Promise<DeviceDTO> {
  const resp = await requestJSON<ApiResponse<DeviceDTO>>('/devices', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })

  return resp.data
}

export async function connectDevice(id: string): Promise<DeviceDTO> {
  const resp = await requestJSON<ApiResponse<DeviceDTO>>('/devices/connect', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ id })
  })

  return resp.data
}

export async function disconnectDevice(id: string): Promise<DeviceDTO> {
  const resp = await requestJSON<ApiResponse<DeviceDTO>>('/devices/disconnect', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ id })
  })

  return resp.data
}

export async function setDeviceStatus(
  id: string,
  status: DeviceDTO['status']
): Promise<{ id: string; status: DeviceDTO['status'] }> {
  const resp = await requestJSON<ApiResponse<{ id: string; status: DeviceDTO['status'] }>>('/devices/status', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ id, status })
  })

  return resp.data
}

export async function fetchUnitConsistency(): Promise<UnitConsistencyDTO> {
  const resp = await requestJSON<ApiResponse<UnitConsistencyDTO>>('/checks/unit-consistency')
  return resp.data
}

export async function fetchDeviceConnectConfig(): Promise<DeviceConnectConfigDTO> {
  const resp = await requestJSON<ApiResponse<DeviceConnectConfigDTO>>('/config/device-connect')
  return resp.data
}
