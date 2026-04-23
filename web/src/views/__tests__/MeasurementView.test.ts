import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import MeasurementView from '../MeasurementView.vue'

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

describe('MeasurementView', () => {
  it('renders measurement module title and sidebar device panels', () => {
    setActivePinia(createPinia())

    const wrapper = mount(MeasurementView, {
      global: {
        stubs: {
          RouterLink: {
            template: '<a><slot /></a>'
          },
          MeasurementDevicePanel: {
            template: '<div>MeasurementDeviceStub</div>'
          },
          PressDevicePanel: {
            template: '<div>PressDeviceStub</div>'
          },
          ChannelMatrix: {
            template: '<div>ChannelMatrixStub</div>'
          },
          ElTable: ElTableStub,
          ElTableColumn: ElTableColumnStub
        }
      }
    })

    expect(wrapper.text()).toContain('计量工作台')
    expect(wrapper.text()).toContain('MeasurementDeviceStub')
    expect(wrapper.text()).toContain('PressDeviceStub')
    expect(wrapper.text()).toContain('启动条件')
    expect(wrapper.text()).toContain('ChannelMatrixStub')
    expect(wrapper.find('.sidebar').exists()).toBe(true)
    expect(wrapper.find('.workbench').exists()).toBe(true)
  })
})
