export enum CalibrationStep {
  DEVICE_CONNECT = 0,
  CHANNEL_SELECT = 1,
  START_CALIBRATION = 2,
  DATA_COLLECTION = 3,
  DATA_FITTING = 4,
  COMPLETED = 5
}

export interface PressurePoint {
  id: string
  index: number
  targetPressure: number
  status: 'pending' | 'pressurizing' | 'stabilizing' | 'collecting' | 'completed' | 'error'
  direction?: 'forward' | 'backward'
  collectedData?: number[]
  actualPressure?: number
}

import type { PressureMode } from '@/types/calibration'

export interface CalibrationParams {
  minValue: number
  maxValue: number
  points: number
  precision: number
  averageCount: number
  stableTime: number
  precisionLevel: string
  pressureMode: PressureMode
}
