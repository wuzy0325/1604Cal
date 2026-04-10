import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import DeviceManagementPanel from '../DeviceManagementPanel.vue'
import * as apiClient from '@/services/apiClient'

vi.mock('@/services/apiClient', () => ({
  createEventStream: vi.fn(),
  fetchDevices: vi.fn(),
  fetchDeviceConnectConfig: vi.fn(),
  fetchUnitConsistency: vi.fn(),
  upsertDevice: vi.fn(),
  setDeviceStatus: vi.fn(),
  connectDevice: vi.fn(),
  disconnectDevice: vi.fn()
}))

const mockDevice = {
  id: 'm1',
  name: 'measure-1',
  type: 'measure' as const,
  model: 'WTN1604',
  host: '192.168.1.100',
  port: 9000,
  unit: 'kPa',
  status: 'connected' as const
}

describe('DeviceManagementPanel', () => {
  const closeSpy = vi.fn()
  let streamCallback: ((payload: apiClient.StreamEventPayload) => void) | null = null

  beforeEach(() => {
    vi.clearAllMocks()

    streamCallback = null

    vi.mocked(apiClient.createEventStream).mockImplementation((onEvent) => {
      streamCallback = onEvent
      return {
        close: closeSpy
      } as unknown as EventSource
    })

    vi.mocked(apiClient.fetchDevices).mockResolvedValue([mockDevice])
    vi.mocked(apiClient.fetchDeviceConnectConfig).mockResolvedValue({
      connectAttemptTimeoutMs: 600,
      connectMaxAttempts: 3,
      connectInitialBackoffMs: 80,
      connectMaxBackoffMs: 300,
      disconnectAttemptTimeoutMs: 400,
      disconnectMaxAttempts: 2,
      disconnectInitialBackoffMs: 40,
      disconnectMaxBackoffMs: 120
    })
    vi.mocked(apiClient.fetchUnitConsistency).mockResolvedValue({
      consistent: true,
      conflicts: []
    })
    vi.mocked(apiClient.upsertDevice).mockResolvedValue(mockDevice)
    vi.mocked(apiClient.setDeviceStatus).mockResolvedValue({
      id: mockDevice.id,
      status: 'disconnected'
    })
    vi.mocked(apiClient.connectDevice).mockResolvedValue(mockDevice)
    vi.mocked(apiClient.disconnectDevice).mockResolvedValue({
      ...mockDevice,
      status: 'disconnected'
    })
  })

  it('subscribes and unsubscribes event stream on lifecycle', async () => {
    const wrapper = mount(DeviceManagementPanel)
    await flushPromises()

    expect(apiClient.createEventStream).toHaveBeenCalled()

    wrapper.unmount()
    expect(closeSpy).toHaveBeenCalled()
  })

  it('renders device from API on mount', async () => {
    const wrapper = mount(DeviceManagementPanel)
    await flushPromises()

    expect(wrapper.text()).toContain('measure-1')
    expect(wrapper.text()).toContain('已连接')
    expect(wrapper.text()).toContain('连接重试策略')
  })

  it('opens create dialog and submits form', async () => {
    const wrapper = mount(DeviceManagementPanel)
    await flushPromises()

    await wrapper.get('[data-test="add-device"]').trigger('click')
    expect(wrapper.text()).toContain('新增设备配置')

    await wrapper.get('[data-test="form-id"]').setValue('p9')
    await wrapper.get('[data-test="form-name"]').setValue('pressure-9')
    await wrapper.get('[data-test="form-type"]').setValue('pressure')
    await wrapper.get('[data-test="form-model"]').setValue('ConST 820')
    await wrapper.get('[data-test="form-host"]').setValue('192.168.1.109')
    await wrapper.get('[data-test="form-port"]').setValue('7009')
    await wrapper.get('[data-test="form-unit"]').setValue('kPa')

    await wrapper.get('[data-test="submit-form"]').trigger('click')
    await flushPromises()

    expect(apiClient.upsertDevice).toHaveBeenCalled()
  })

  it('calls connect endpoint when clicking connect button', async () => {
    vi.mocked(apiClient.fetchDevices).mockResolvedValueOnce([
      {
        ...mockDevice,
        status: 'disconnected'
      }
    ])

    const wrapper = mount(DeviceManagementPanel)
    await flushPromises()

    await wrapper.findAll('button').find((btn) => btn.text() === '连接')?.trigger('click')
    await flushPromises()

    expect(apiClient.connectDevice).toHaveBeenCalledWith('m1')
  })

  it('shows duplicate id validation error in create mode', async () => {
    const wrapper = mount(DeviceManagementPanel)
    await flushPromises()

    await wrapper.get('[data-test="add-device"]').trigger('click')
    await wrapper.get('[data-test="form-id"]').setValue('m1')
    await wrapper.get('[data-test="form-host"]').setValue('192.168.1.109')
    await wrapper.get('[data-test="form-port"]').setValue('7009')
    await wrapper.get('[data-test="form-unit"]').setValue('kPa')
    await wrapper.get('[data-test="submit-form"]').trigger('click')

    expect(wrapper.text()).toContain('设备ID已存在')
    expect(apiClient.upsertDevice).not.toHaveBeenCalled()
  })

  it('shows ip and port validation errors', async () => {
    const wrapper = mount(DeviceManagementPanel)
    await flushPromises()

    await wrapper.get('[data-test="add-device"]').trigger('click')
    await wrapper.get('[data-test="form-id"]').setValue('m9')
    await wrapper.get('[data-test="form-host"]').setValue('invalid-ip')
    await wrapper.get('[data-test="form-port"]').setValue('7009')
    await wrapper.get('[data-test="form-unit"]').setValue('kPa')
    await wrapper.get('[data-test="submit-form"]').trigger('click')

    expect(wrapper.text()).toContain('IP地址格式不正确')

    await wrapper.get('[data-test="form-host"]').setValue('192.168.1.110')
    await wrapper.get('[data-test="form-port"]').setValue('70000')
    await wrapper.get('[data-test="submit-form"]').trigger('click')

    expect(wrapper.text()).toContain('端口必须在1-65535之间')
    expect(apiClient.upsertDevice).not.toHaveBeenCalled()
  })

  it('renders error reason and last error time from api data', async () => {
    vi.mocked(apiClient.fetchDevices).mockResolvedValueOnce([
      {
        ...mockDevice,
        status: 'error',
        lastErrorReason: 'tcp timeout',
        lastErrorAt: '2026-04-10T12:00:00Z'
      }
    ])

    const wrapper = mount(DeviceManagementPanel)
    await flushPromises()

    expect(wrapper.text()).toContain('错误原因：tcp timeout')
    expect(wrapper.text()).toContain('最近错误时间')
  })

  it('applies status error fields from sse payload', async () => {
    const wrapper = mount(DeviceManagementPanel)
    await flushPromises()

    streamCallback?.({
      type: 'device.status.changed',
      data: {
        id: 'm1',
        status: 'error',
        errorReason: 'tcp timeout',
        lastErrorAt: '2026-04-10T12:00:00Z'
      }
    })
    await flushPromises()

    expect(wrapper.text()).toContain('错误原因：tcp timeout')
    expect(wrapper.text()).toContain('最近错误时间')
    expect(wrapper.text()).toContain('异常')
  })
})
