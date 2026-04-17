import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import CalibrationView from '../CalibrationView.vue'

// Mock EventSource for jsdom environment
class MockEventSource {
  onmessage: ((event: MessageEvent) => void) | null = null
  onerror: ((event: Event) => void) | null = null
  close() {}
  addEventListener() {}
  removeEventListener() {}
  dispatchEvent() { return true }
  readonly readyState = 0
  readonly url = ''
  readonly withCredentials = false
  readonly CONNECTING = 0
  readonly OPEN = 1
  readonly CLOSED = 2
}

// @ts-expect-error - jsdom doesn't have EventSource
globalThis.EventSource = MockEventSource

// Stub el-table-column to avoid scoped slot destructuring errors in test
const ElTableStub = {
  template: '<div class="el-table-stub"><slot /></div>'
}
const ElTableColumnStub = {
  template: '<span />'
}

describe('CalibrationView', () => {
  it('renders calibration controls and selector', () => {
    setActivePinia(createPinia())

    const wrapper = mount(CalibrationView, {
      global: {
        stubs: {
          RouterLink: {
            template: '<a><slot /></a>'
          },
          Device1604Panel: {
            template: '<div>Device1604Stub</div>'
          },
          PressDevicePanel: {
            template: '<div>PressDeviceStub</div>'
          },
          ChannelMatrix: {
            template: '<div>ChannelMatrixStub</div>'
          },
          ProgressIndicator: {
            template: '<div>ProgressStub</div>'
          },
          ElTable: ElTableStub,
          ElTableColumn: ElTableColumnStub
        }
      }
    })

    expect(wrapper.text()).toContain('标定模块')
    expect(wrapper.text()).toContain('Device1604Stub')
    expect(wrapper.text()).toContain('ChannelMatrixStub')
    expect(wrapper.text()).toContain('生成测点')
    expect(wrapper.text()).toContain('标定数据')
  })
})
