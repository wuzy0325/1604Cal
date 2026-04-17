import { mount } from '@vue/test-utils'

import ModuleHubView from '../ModuleHubView.vue'

describe('ModuleHubView', () => {
  it('renders four module entries', () => {
    const wrapper = mount(ModuleHubView, {
      global: {
        stubs: {
          RouterLink: {
            template: '<a><slot /></a>'
          }
        }
      }
    })

    expect(wrapper.text()).toContain('设备管理模块')
    expect(wrapper.text()).toContain('计量模块')
    expect(wrapper.text()).toContain('标定模块')
    expect(wrapper.text()).toContain('多设备打压模块')
    expect(wrapper.text()).toContain('进入设备管理')
    expect(wrapper.text()).toContain('进入计量模块')
    expect(wrapper.text()).toContain('进入标定模块')
    expect(wrapper.text()).toContain('进入打压模块')
  })
})
