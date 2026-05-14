import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

import { useMeasurementStore } from '../index'
import * as sessionApi from '@/api/session'
import * as measurementApi from '@/api/measurement'

// ── Mock API 层 ──

vi.mock('@/api/session', () => ({
  bindDevices: vi.fn(),
  bindMeasureDevice: vi.fn(),
  readPressure: vi.fn(),
  readStability: vi.fn(),
  readMeasureData: vi.fn(),
  readValveStatus: vi.fn(),
  setValveStatus: vi.fn(),
  readMeasureUnit: vi.fn(),
  setMeasureUnit: vi.fn(),
  readDeviceInfo: vi.fn(),
  resetDevice: vi.fn()
}))

vi.mock('@/api/measurement', () => ({
  fetchMeasurementState: vi.fn(),
  startMeasurement: vi.fn(),
  pauseMeasurement: vi.fn(),
  stopMeasurement: vi.fn(),
  fetchMeasurementData: vi.fn(),
  generateMeasurementPoints: vi.fn(),
  saveMeasurementParamsConfig: vi.fn()
}))

vi.mock('element-plus', async () => {
  const actual = await vi.importActual<typeof import('element-plus')>('element-plus')
  return {
    ...actual,
    ElMessage: { success: vi.fn(), warning: vi.fn(), error: vi.fn(), info: vi.fn() }
  }
})

describe('useMeasurementStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  // ── 初始状态 ──

  describe('initial state', () => {
    it('starts idle with empty data', () => {
      const store = useMeasurementStore()
      expect(store.state).toBe('idle')
      expect(store.rows).toEqual([])
      expect(store.channels).toHaveLength(16)
      expect(store.measureDeviceId).toBe('')
      expect(store.pressureDeviceId).toBe('')
      expect(store.currentPressure).toBe(0)
      expect(store.isStable).toBe(false)
    })

    it('computes isCollecting/isPaused/isIdle correctly', () => {
      const store = useMeasurementStore()
      expect(store.isIdle).toBe(true)
      expect(store.isRunning).toBe(false)
      expect(store.isCollecting).toBe(false)
      expect(store.isPaused).toBe(false)
      expect(store.deviceBound).toBe(false)
    })
  })

  describe('isRunning', () => {
    it('is true for pressurizing/stabilizing/collecting', () => {
      const store = useMeasurementStore()

      store.syncState('pressurizing')
      expect(store.isRunning).toBe(true)

      store.syncState('stabilizing')
      expect(store.isRunning).toBe(true)

      store.syncState('collecting')
      expect(store.isRunning).toBe(true)
    })

    it('is false for idle/paused/completed/error', () => {
      const store = useMeasurementStore()

      store.syncState('idle')
      expect(store.isRunning).toBe(false)

      store.syncState('paused')
      expect(store.isRunning).toBe(false)

      store.syncState('completed')
      expect(store.isRunning).toBe(false)

      store.syncState('error')
      expect(store.isRunning).toBe(false)
    })
  })

  // ── 设备绑定 ──

  describe('bindDevices', () => {
    it('calls API and stores both device IDs', async () => {
      vi.mocked(sessionApi.bindDevices).mockResolvedValue(undefined)
      const store = useMeasurementStore()
      await store.bindDevices('m1', 'p1')
      expect(sessionApi.bindDevices).toHaveBeenCalledWith('m1', 'p1', 'measurement')
      expect(store.measureDeviceId).toBe('m1')
      expect(store.pressureDeviceId).toBe('p1')
      expect(store.deviceBound).toBe(true)
    })
  })

  describe('bindMeasureDevice', () => {
    it('calls API and stores measure device ID', async () => {
      vi.mocked(sessionApi.bindMeasureDevice).mockResolvedValue(undefined)
      const store = useMeasurementStore()
      await store.bindMeasureDevice('m2')
      expect(sessionApi.bindMeasureDevice).toHaveBeenCalledWith('m2', 'measurement')
      expect(store.measureDeviceId).toBe('m2')
      expect(store.deviceBound).toBe(true)
    })
  })

  describe('unbind device ids', () => {
    it('clears pressure device id when unbinding pressure device', async () => {
      vi.mocked(sessionApi.bindDevices).mockResolvedValue(undefined)
      const store = useMeasurementStore()

      await store.bindDevices('m1', 'p1')
      expect(store.pressureDeviceId).toBe('p1')

      store.unbindPressureDevice()
      expect(store.pressureDeviceId).toBe('')
      expect(store.measureDeviceId).toBe('m1')
      expect(store.deviceBound).toBe(true)
    })

    it('clears both ids when unbinding measure device', async () => {
      vi.mocked(sessionApi.bindDevices).mockResolvedValue(undefined)
      const store = useMeasurementStore()

      await store.bindDevices('m1', 'p1')
      expect(store.deviceBound).toBe(true)

      store.unbindMeasureDevice()
      expect(store.measureDeviceId).toBe('')
      expect(store.pressureDeviceId).toBe('')
      expect(store.deviceBound).toBe(false)
    })
  })

  // ── 实时数据刷新 ──

  describe('refreshPressure', () => {
    it('updates currentPressure on success', async () => {
      vi.mocked(sessionApi.readPressure).mockResolvedValue(42.5)
      const store = useMeasurementStore()
      await store.refreshPressure()
      expect(store.currentPressure).toBe(42.5)
    })

    it('silently ignores errors', async () => {
      vi.mocked(sessionApi.readPressure).mockRejectedValue(new Error('no device'))
      const store = useMeasurementStore()
      await expect(store.refreshPressure()).resolves.toBeUndefined()
      expect(store.currentPressure).toBe(0)
    })
  })

  describe('refreshStability', () => {
    it('updates isStable', async () => {
      vi.mocked(sessionApi.readStability).mockResolvedValue(true)
      const store = useMeasurementStore()
      await store.refreshStability()
      expect(store.isStable).toBe(true)
    })
  })

  describe('refreshMeasureData', () => {
    it('updates channelData', async () => {
      vi.mocked(sessionApi.readMeasureData).mockResolvedValue([1.1, 2.2, 3.3])
      const store = useMeasurementStore()
      await store.refreshMeasureData()
      expect(store.channelData).toEqual([1.1, 2.2, 3.3])
    })
  })

  describe('refreshValveStatus', () => {
    it('updates valveStatus', async () => {
      vi.mocked(sessionApi.readValveStatus).mockResolvedValue('measurement')
      const store = useMeasurementStore()
      await store.refreshValveStatus()
      expect(store.valveStatus).toBe('measurement')
    })
  })

  describe('setValveStatus', () => {
    it('calls API and updates local state', async () => {
      vi.mocked(sessionApi.setValveStatus).mockResolvedValue(undefined)
      const store = useMeasurementStore()
      await store.setValveStatus('calibration')
      expect(sessionApi.setValveStatus).toHaveBeenCalledWith('calibration')
      expect(store.valveStatus).toBe('calibration')
    })
  })

  describe('refreshMeasureUnit', () => {
    it('updates measureUnit', async () => {
      vi.mocked(sessionApi.readMeasureUnit).mockResolvedValue('MPa')
      const store = useMeasurementStore()
      await store.refreshMeasureUnit()
      expect(store.measureUnit).toBe('MPa')
    })
  })

  describe('refreshDeviceInfo', () => {
    it('updates deviceInfo', async () => {
      vi.mocked(sessionApi.readDeviceInfo).mockResolvedValue({ model: 'WTN1604' })
      const store = useMeasurementStore()
      await store.refreshDeviceInfo()
      expect(store.deviceInfo).toEqual({ model: 'WTN1604' })
    })
  })

  // ── 采集工作流 ──

  describe('start', () => {
    it('fails with warning when no device bound', async () => {
      const { ElMessage } = await import('element-plus')
      const store = useMeasurementStore()
      await store.start([1, 2])
      expect(measurementApi.startMeasurement).not.toHaveBeenCalled()
      expect(ElMessage.warning).toHaveBeenCalledWith('请先绑定计量设备')
    })

    it('calls API, updates state, clears rows on success', async () => {
      vi.mocked(sessionApi.bindMeasureDevice).mockResolvedValue(undefined)
      vi.mocked(measurementApi.saveMeasurementParamsConfig).mockResolvedValue(undefined)
      vi.mocked(measurementApi.generateMeasurementPoints).mockResolvedValue([])
      vi.mocked(measurementApi.startMeasurement).mockResolvedValue('collecting')
      const store = useMeasurementStore()
      await store.bindMeasureDevice('m1')
      store.rows = [{ timestamp: 'old', channels: { '1': 0 } }]

      await store.start([1, 2, 3])

      expect(measurementApi.startMeasurement).toHaveBeenCalledWith([1, 2, 3])
      expect(store.state).toBe('collecting')
      expect(store.channels).toEqual([1, 2, 3])
      expect(store.rows).toEqual([])
    })

    it('shows error on API failure', async () => {
      const { ElMessage } = await import('element-plus')
      vi.mocked(sessionApi.bindMeasureDevice).mockResolvedValue(undefined)
      vi.mocked(measurementApi.saveMeasurementParamsConfig).mockResolvedValue(undefined)
      vi.mocked(measurementApi.generateMeasurementPoints).mockResolvedValue([])
      vi.mocked(measurementApi.startMeasurement).mockRejectedValue(new Error('transition denied'))
      const store = useMeasurementStore()
      await store.bindMeasureDevice('m1')

      await store.start([1])

      expect(ElMessage.error).toHaveBeenCalled()
      expect(store.state).toBe('idle')
    })
  })

  describe('pause', () => {
    it('updates state on success', async () => {
      vi.mocked(measurementApi.pauseMeasurement).mockResolvedValue('paused')
      const store = useMeasurementStore()
      await store.pause()
      expect(store.state).toBe('paused')
    })
  })

  describe('stop', () => {
    it('updates state on success', async () => {
      vi.mocked(measurementApi.stopMeasurement).mockResolvedValue('idle')
      const store = useMeasurementStore()
      await store.stop()
      expect(store.state).toBe('idle')
    })
  })

  describe('refreshData', () => {
    it('updates rows from API response', async () => {
      const mockRows = [
        { timestamp: '2026-04-21T10:00:00Z', channels: { '1': 1.1, '2': 2.2 } }
      ]
      vi.mocked(measurementApi.fetchMeasurementData).mockResolvedValue({
        rows: mockRows,
        total: 1
      })
      const store = useMeasurementStore()
      await store.refreshData()
      expect(store.rows).toEqual(mockRows)
    })
  })

  describe('fetchCurrentState', () => {
    it('syncs state from API', async () => {
      vi.mocked(measurementApi.fetchMeasurementState).mockResolvedValue('collecting')
      const store = useMeasurementStore()
      await store.fetchCurrentState()
      expect(store.state).toBe('collecting')
    })

    it('silently ignores errors', async () => {
      vi.mocked(measurementApi.fetchMeasurementState).mockRejectedValue(new Error('fail'))
      const store = useMeasurementStore()
      await expect(store.fetchCurrentState()).resolves.toBeUndefined()
      expect(store.state).toBe('idle')
    })
  })

  describe('syncState', () => {
    it('directly sets state', () => {
      const store = useMeasurementStore()
      store.syncState('paused')
      expect(store.state).toBe('paused')
      expect(store.isPaused).toBe(true)
    })
  })

  describe('totalRows', () => {
    it('counts rows', () => {
      const store = useMeasurementStore()
      expect(store.totalRows).toBe(0)
      store.rows = [
        { timestamp: 't1', channels: {} },
        { timestamp: 't2', channels: {} }
      ]
      expect(store.totalRows).toBe(2)
    })
  })
})
