import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import {
  setCalibrationConfig,
  generatePressurePoints as apiGeneratePoints,
  getPressurePoints as apiGetPoints,
  pressurize as apiPressurize,
  collectData as apiCollectData
} from "@/api/calibration"
import type { FittingResultDTO } from "@/types/calibration"
import type { PressurePoint } from './types'

export type { PressurePoint } from './types'

const POINTS_KEY = 'cal1604_pressure_points'

function loadSavedPoints(): PressurePoint[] {
  try {
    const raw = localStorage.getItem(POINTS_KEY)
    if (raw) {
      const saved = JSON.parse(raw) as PressurePoint[]
      return saved.map(p => ({ ...p, status: 'pending' as const, collectedData: undefined, actualPressure: undefined }))
    }
  } catch { /* ignore */ }
  return []
}

function savePoints(points: PressurePoint[]) {
  try {
    localStorage.setItem(POINTS_KEY, JSON.stringify(points))
  } catch { /* ignore */ }
}

function clearSavedPoints() {
  try {
    localStorage.removeItem(POINTS_KEY)
  } catch { /* ignore */ }
}

export const usePressurePointStore = defineStore('pressurePoint', () => {
  // State
  const pressurePoints = ref<PressurePoint[]>(loadSavedPoints())
  const fittingResult = ref<FittingResultDTO | null>(null)

  // Getters
  const hasCollectedData = computed(() =>
    pressurePoints.value.some(p => p.status === 'completed')
  )

  // Actions
  // 生成压力点（标定模块使用）
  const generatePressurePoints = async (opts?: {
    controlMode?: string
    pressureMode?: string
    channels?: number[]
    params?: { points: number; averageCount: number; minValue: number; maxValue: number; stableTime: number }
  }) => {
    try {
      const channels = opts?.channels ?? []
      const params = opts?.params

      await setCalibrationConfig({
        channels,
        pressurePoints: params?.points ?? 2,
        averageCount: params?.averageCount ?? 5,
        minPressure: params?.minValue ?? 0,
        maxPressure: params?.maxValue ?? 100,
        stableWaitMs: (params?.stableTime ?? 3) * 1000,
        controlMode: (opts?.controlMode as 'auto' | 'manual') || undefined
      })

      const points = await apiGeneratePoints()
      pressurePoints.value = points.map(p => ({
        id: `point-${p.index}`,
        index: p.index,
        targetPressure: p.targetPressure,
        status: p.status as PressurePoint['status'],
        collectedData: p.collectedData,
        actualPressure: p.actualPressure
      }))

      ElMessage.success(`已生成 ${points.length} 个压力点`)
      savePoints(pressurePoints.value)
      return true
    } catch (error) {
      console.error('生成压力点失败:', error)
      ElMessage.error('生成压力点失败')
      return false
    }
  }

  // 添加压力点
  const addPressurePoint = (point: Omit<PressurePoint, 'id'>) => {
    pressurePoints.value.push({
      ...point,
      id: crypto.randomUUID()
    })
  }

  // 删除压力点
  const removePressurePoint = (index: number) => {
    pressurePoints.value.splice(index, 1)
  }

  // 更新压力点状态
  const updatePointStatus = (pointId: string, status: PressurePoint['status']) => {
    const point = pressurePoints.value.find(p => p.id === pointId)
    if (point) {
      point.status = status
    }
  }

  // 打压
  const pressurize = async (pointId: string) => {
    const point = pressurePoints.value.find(p => p.id === pointId)
    if (!point) return

    try {
      point.status = 'pressurizing'
      await apiPressurize(point.index)

      // 打压完成后刷新压力点状态
      const points = await apiGetPoints()
      const updatedPoint = points.find(p => p.index === point.index)
      if (updatedPoint) {
        point.status = updatedPoint.status as PressurePoint['status']
        point.actualPressure = updatedPoint.actualPressure
      } else {
        point.status = 'stabilizing'
      }

      ElMessage.success(`压力点 ${point.index} 打压完成，压力已稳定`)
    } catch (error) {
      console.error('打压失败:', error)
      point.status = 'error'
      ElMessage.error('打压失败')
    }
  }

  // 确认压力
  const confirmPressure = (pointId: string) => {
    const point = pressurePoints.value.find(p => p.id === pointId)
    if (!point) return

    if (point.status === 'pressurizing') {
      point.status = 'stabilizing'
    }

    if (point.status !== 'stabilizing') {
      ElMessage.warning('当前点未进入可确认状态')
      return
    }

    ElMessage.success('压力已确认，可以进行采集')
  }

  // 采集数据
  const collectData = async (pointId: string) => {
    const point = pressurePoints.value.find(p => p.id === pointId)
    if (!point) return

    try {
      point.status = 'collecting'
      const data = await apiCollectData(point.index)

      point.collectedData = data
      point.status = 'completed'

      ElMessage.success(`压力点 ${point.index} 采集完成`)
    } catch (error) {
      console.error('采集数据失败:', error)
      point.status = 'error'
      const detail = error instanceof Error ? error.message : String(error)
      ElMessage.error(`采集数据失败: ${detail}`)
    }
  }

  // 重置采集数据（标定模块专用，仅重置测点状态，保留配置）
  const resetCollection = () => {
    pressurePoints.value = pressurePoints.value.map(p => ({
      ...p,
      status: 'pending' as PressurePoint['status'],
      collectedData: undefined,
      actualPressure: undefined
    }))
  }

  // 清空压力点
  const clearPoints = () => {
    pressurePoints.value = []
    clearSavedPoints()
  }

  return {
    // State
    pressurePoints,
    fittingResult,
    // Getters
    hasCollectedData,
    // Actions
    generatePressurePoints,
    addPressurePoint,
    removePressurePoint,
    updatePointStatus,
    pressurize,
    confirmPressure,
    collectData,
    resetCollection,
    clearPoints
  }
})
