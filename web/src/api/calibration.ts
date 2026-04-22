import type { ApiResponse } from '@/types/api'
import type { SessionStateDTO, ReportTemplateDTO, CalibrationConfigDTO, PressurePointDTO, FittingResultDTO } from '@/types/calibration'
import { requestJSON } from './client'

export async function fetchSessionState(): Promise<SessionStateDTO> {
  const resp = await requestJSON<ApiResponse<SessionStateDTO>>('/sessions/state')
  return resp.data
}

export async function triggerSessionAction(action: 'start' | 'pause' | 'resume' | 'stop'): Promise<SessionStateDTO> {
  const resp = await requestJSON<ApiResponse<SessionStateDTO>>(`/sessions/${action}`, {
    method: 'POST'
  })
  return resp.data
}

export async function selectReportTemplate(points: number, mode: 'single' | 'return'): Promise<ReportTemplateDTO> {
  const resp = await requestJSON<ApiResponse<ReportTemplateDTO>>(
    `/reports/templates/select?points=${points}&mode=${mode}`
  )
  return resp.data
}

// ---------------------------------------------------------------------------
// 校准流程 API
// ---------------------------------------------------------------------------

/** 设置校准配置 */
export async function setCalibrationConfig(config: CalibrationConfigDTO): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/calibration/config', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config)
  })
}

/** 设置采集通道 */
export async function setCalibrationChannels(channels: number[]): Promise<number[]> {
  const resp = await requestJSON<ApiResponse<{ channels: number[] }>>('/calibration/channels', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ channels })
  })
  return resp.data.channels
}

/** 获取当前通道配置 */
export async function getCalibrationChannels(): Promise<number[]> {
  const resp = await requestJSON<ApiResponse<{ channels: number[] }>>('/calibration/channels/list')
  return resp.data.channels
}

/** 生成压力点 */
export async function generatePressurePoints(): Promise<PressurePointDTO[]> {
  const resp = await requestJSON<ApiResponse<PressurePointDTO[]>>('/calibration/points/generate', {
    method: 'POST'
  })
  return resp.data
}

/** 获取压力点列表 */
export async function getPressurePoints(): Promise<PressurePointDTO[]> {
  const resp = await requestJSON<ApiResponse<PressurePointDTO[]>>('/calibration/points')
  return resp.data
}

/** 执行打压 */
export async function pressurize(pointIndex: number): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/calibration/pressurize', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pointIndex })
  })
}

/** 采集数据 */
export async function collectData(pointIndex: number): Promise<number[]> {
  const resp = await requestJSON<ApiResponse<{ data: number[] }>>('/calibration/collect', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pointIndex })
  })
  return resp.data.data
}

/** 执行拟合 */
export async function fitData(): Promise<FittingResultDTO> {
  const resp = await requestJSON<ApiResponse<FittingResultDTO>>('/calibration/fit', {
    method: 'POST'
  })
  return resp.data
}

/** 确认报警决策（continue 或 retry） */
export async function resolveAlarm(decision: 'continue' | 'retry'): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/calibration/resolve-alarm', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ decision })
  })
}

/** 重试指定压力点 */
export async function retryPoint(pointIndex: number): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/calibration/retry-point', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ pointIndex })
  })
}

// ---------------------------------------------------------------------------
// 配置持久化 API
// ---------------------------------------------------------------------------

export interface CalibrationParamsPayload {
  minPressure: number
  maxPressure: number
  pointCount: number
  precision: number
  averageCount: number
  stableDurationMs: number
  precisionLevel: number
  pressureMode: string
  controlMode: string
}

export interface AlarmConfigPayload {
  enabled: boolean
  precisionThreshold: number
  soundEnabled: boolean
  confirmOnAlarm: boolean
  enabledChannels: number[]
}

/** 获取持久化校准参数配置 */
export async function getCalibrationParamsConfig(): Promise<CalibrationParamsPayload> {
  const resp = await requestJSON<ApiResponse<CalibrationParamsPayload>>('/config/calibration')
  return resp.data
}

/** 保存校准参数配置 */
export async function saveCalibrationParamsConfig(params: CalibrationParamsPayload): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/config/calibration', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params)
  })
}

/** 获取持久化报警配置 */
export async function getAlarmConfig(): Promise<AlarmConfigPayload> {
  const resp = await requestJSON<ApiResponse<AlarmConfigPayload>>('/config/alarm')
  return resp.data
}

/** 保存报警配置 */
export async function saveAlarmConfig(config: AlarmConfigPayload): Promise<void> {
  await requestJSON<ApiResponse<{ status: string }>>('/config/alarm', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config)
  })
}
