import { mount } from '@vue/test-utils'

import CalibrationView from '../CalibrationView.vue'

describe('CalibrationView', () => {
  it('renders calibration controls and selector', () => {
    const wrapper = mount(CalibrationView, {
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

    expect(wrapper.text()).toContain('标定模块')
    expect(wrapper.text()).toContain('DeviceSelectionStub')
    expect(wrapper.text()).toContain('标定流程控制')
    expect(wrapper.text()).toContain('报告模板选择')
  })
})
