import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

import { useCalibrationStore } from '../index'
import * as apiClient from '@/services/apiClient'

vi.mock('@/services/apiClient', async () => {
  const actual = await vi.importActual<typeof import('@/services/apiClient')>('@/services/apiClient')
  return {
    ...actual,
    readCalibrationValve: vi.fn(),
    readCalibrationMeasureUnit: vi.fn(),
    readCalibrationDeviceInfo: vi.fn()
  }
})

describe('calibration store refreshDeviceInfo', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.resetAllMocks()
  })

  it('returns true when valve and unit are loaded even if device info fails', async () => {
    vi.mocked(apiClient.readCalibrationValve).mockResolvedValue('measurement')
    vi.mocked(apiClient.readCalibrationMeasureUnit).mockResolvedValue('kPa')
    vi.mocked(apiClient.readCalibrationDeviceInfo).mockRejectedValue(new Error('device info not supported'))

    const store = useCalibrationStore()
    const loaded = await store.refreshDeviceInfo({ retries: 1 })

    expect(loaded).toBe(true)
    expect(store.valveStatus).toBe('measurement')
    expect(store.measureUnit).toBe('kPa')
  })

  it('returns false when valve or unit is still unavailable', async () => {
    vi.mocked(apiClient.readCalibrationValve).mockRejectedValue(new Error('valve read failed'))
    vi.mocked(apiClient.readCalibrationMeasureUnit).mockResolvedValue('kPa')
    vi.mocked(apiClient.readCalibrationDeviceInfo).mockResolvedValue({ model: 'WTN1604' })

    const store = useCalibrationStore()
    const loaded = await store.refreshDeviceInfo({ retries: 1 })

    expect(loaded).toBe(false)
  })

  it('returns true when valve and unit succeed on different retries', async () => {
    vi.mocked(apiClient.readCalibrationValve)
      .mockResolvedValueOnce('measurement')
      .mockRejectedValueOnce(new Error('valve timeout'))
    vi.mocked(apiClient.readCalibrationMeasureUnit)
      .mockRejectedValueOnce(new Error('unit timeout'))
      .mockResolvedValueOnce('kPa')
    vi.mocked(apiClient.readCalibrationDeviceInfo).mockRejectedValue(new Error('device info not supported'))

    const store = useCalibrationStore()
    const loaded = await store.refreshDeviceInfo({ retries: 2 })

    expect(loaded).toBe(true)
    expect(store.valveStatus).toBe('measurement')
    expect(store.measureUnit).toBe('kPa')
  })
})
