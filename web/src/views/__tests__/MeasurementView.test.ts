import { mount } from '@vue/test-utils'

import MeasurementView from '../MeasurementView.vue'

describe('MeasurementView', () => {
  it('renders measurement module title and device selector', () => {
    const wrapper = mount(MeasurementView, {
      global: {
        stubs: {
          RouterLink: {
            template: '<a><slot /></a>'
          },
          DeviceSelectionPanel: {
            template: '<div>DeviceSelectionStub</div>'
          }
        }
      }
    })

    expect(wrapper.text()).toContain('计量模块')
    expect(wrapper.text()).toContain('DeviceSelectionStub')
    expect(wrapper.text()).toContain('进入标定模块')
  })
})
