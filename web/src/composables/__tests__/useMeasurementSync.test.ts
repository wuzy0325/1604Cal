import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { useMeasurementStore } from '@/stores/measurement'
import * as measurementApi from '@/api/measurement'
import { useMeasurementSync } from '../useMeasurementSync'

const subscribe = vi.fn(() => vi.fn())
const registerPoll = vi.fn(() => vi.fn())

vi.mock('@/composables/useEventHub', () => ({
  useEventHub: () => ({ subscribe, registerPoll })
}))

vi.mock('@/stores/device/inventoryStore', () => ({
  useDeviceInventoryStore: () => ({ updateDevicePressure: vi.fn() })
}))

vi.mock('@/api/multipress', () => ({
  multipressListDevices: vi.fn().mockResolvedValue([])
}))

vi.mock('@/api/session', () => ({
  readPressure: vi.fn(),
  readStability: vi.fn(),
  readMeasureData: vi.fn(),
  bindDevices: vi.fn(),
  bindMeasureDevice: vi.fn(),
  readValveStatus: vi.fn(),
  setValveStatus: vi.fn(),
  readMeasureUnit: vi.fn(),
  setMeasureUnit: vi.fn(),
  readDeviceInfo: vi.fn(),
  resetDevice: vi.fn()
}))

vi.mock('@/api/measurement', () => ({
  fetchMeasurementState: vi.fn(),
  fetchMeasurementData: vi.fn(),
  fetchMeasurementPoints: vi.fn(),
  getMeasurementAlarmConfig: vi.fn(),
  checkMeasurementAlarmPending: vi.fn(),
  getMeasurementParamsConfig: vi.fn(),
  saveMeasurementParamsConfig: vi.fn(),
  startMeasurement: vi.fn(),
  pauseMeasurement: vi.fn(),
  stopMeasurement: vi.fn(),
  generateMeasurementPoints: vi.fn(),
  resolveMeasurementAlarm: vi.fn(),
  autoCollectMeasurement: vi.fn(),
  manualPressurizeMeasurement: vi.fn(),
  manualCollectMeasurement: vi.fn(),
  manualStartMeasurement: vi.fn(),
  saveMeasurementAlarmConfig: vi.fn(),
  resolveStabilityTimeout: vi.fn()
}))

function mountSync() {
  return mount({
    setup() {
      useMeasurementSync()
      return () => null
    }
  })
}

describe('useMeasurementSync', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.mocked(measurementApi.getMeasurementAlarmConfig).mockResolvedValue({
      enabled: true,
      enabledChannels: [1],
      confirmOnAlarm: false,
      soundEnabled: true
    })
    vi.mocked(measurementApi.fetchMeasurementState).mockResolvedValue('collecting')
    vi.mocked(measurementApi.fetchMeasurementPoints).mockResolvedValue([])
    vi.mocked(measurementApi.fetchMeasurementData).mockResolvedValue({ rows: [], total: 0 })
    vi.mocked(measurementApi.checkMeasurementAlarmPending).mockResolvedValue(false)
  })

  it('does not register subscriptions after async setup finishes after unmount', async () => {
    let resolveState!: (state: string) => void
    vi.mocked(measurementApi.fetchMeasurementState).mockImplementationOnce(
      () => new Promise(resolve => { resolveState = resolve })
    )

    const wrapper = mountSync()
    wrapper.unmount()
    resolveState('collecting')
    await Promise.resolve()
    await Promise.resolve()

    expect(subscribe).not.toHaveBeenCalled()
    expect(registerPoll).not.toHaveBeenCalled()
  })

  it('rehydrates the active workflow after mounting', async () => {
    const rows = [{ timestamp: '2026-08-25T10:00:00Z', channels: { '1': 1.2 } }]
    const points = [{ id: 'p1', index: 1, targetPressure: 1, direction: 'up', status: 'completed' }]
    vi.mocked(measurementApi.fetchMeasurementData).mockResolvedValue({ rows, total: 1 })
    vi.mocked(measurementApi.fetchMeasurementPoints).mockResolvedValue(points)
    vi.mocked(measurementApi.checkMeasurementAlarmPending).mockResolvedValue(true)

    const wrapper = mountSync()
    await vi.waitFor(() => expect(useMeasurementStore().state).toBe('collecting'))

    const store = useMeasurementStore()
    expect(store.state).toBe('collecting')
    expect(store.rows).toEqual(rows)
    expect(store.points).toEqual(points)
    expect(store.alarmPending).toBe(true)
    expect(subscribe).toHaveBeenCalled()
    expect(registerPoll).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })
})
