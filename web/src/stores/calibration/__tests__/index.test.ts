import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

import { useDeviceControlStore } from '../deviceControl'
import * as sessionApi from '@/api/session'

vi.mock('@/api/session', async () => {
  const actual = await vi.importActual<typeof import('@/api/session')>('@/api/session')
  return {
    ...actual,
    readValveStatus: vi.fn(),
    readMeasureUnit: vi.fn(),
    readDeviceInfo: vi.fn()
  }
})

describe('calibration store refreshDeviceInfo', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.resetAllMocks()
  })

  it('returns true when valve and unit are loaded even if device info fails', async () => {
    vi.mocked(sessionApi.readValveStatus).mockResolvedValue('measurement')
    vi.mocked(sessionApi.readMeasureUnit).mockResolvedValue('kPa')
    vi.mocked(sessionApi.readDeviceInfo).mockRejectedValue(new Error('device info not supported'))

    const deviceControlStore = useDeviceControlStore()
    const loaded = await deviceControlStore.refreshDeviceInfo({ retries: 1 })

    expect(loaded).toBe(true)
    expect(deviceControlStore.valveStatus).toBe('measurement')
    expect(deviceControlStore.measureUnit).toBe('kPa')
  })

  it('returns false when valve or unit is still unavailable', async () => {
    vi.mocked(sessionApi.readValveStatus).mockRejectedValue(new Error('valve read failed'))
    vi.mocked(sessionApi.readMeasureUnit).mockResolvedValue('kPa')
    vi.mocked(sessionApi.readDeviceInfo).mockResolvedValue({ model: 'WTN1604' })

    const deviceControlStore = useDeviceControlStore()
    const loaded = await deviceControlStore.refreshDeviceInfo({ retries: 1 })

    expect(loaded).toBe(false)
  })

  it('returns true when valve and unit succeed on different retries', async () => {
    vi.mocked(sessionApi.readValveStatus)
      .mockResolvedValueOnce('measurement')
      .mockRejectedValueOnce(new Error('valve timeout'))
    vi.mocked(sessionApi.readMeasureUnit)
      .mockRejectedValueOnce(new Error('unit timeout'))
      .mockResolvedValueOnce('kPa')
    vi.mocked(sessionApi.readDeviceInfo).mockRejectedValue(new Error('device info not supported'))

    const deviceControlStore = useDeviceControlStore()
    const loaded = await deviceControlStore.refreshDeviceInfo({ retries: 2 })

    expect(loaded).toBe(true)
    expect(deviceControlStore.valveStatus).toBe('measurement')
    expect(deviceControlStore.measureUnit).toBe('kPa')
  })
})
