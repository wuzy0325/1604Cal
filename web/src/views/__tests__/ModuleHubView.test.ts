import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'

import ModuleHubView from '../ModuleHubView.vue'

describe('ModuleHubView', () => {
  it('renders four module entries', async () => {
    const router = createRouter({
      history: createWebHistory(),
      routes: [{ path: '/', component: { template: '<div />' } }]
    })
    // ensure router is ready before mounting
    await router.push('/')

    const wrapper = mount(ModuleHubView, {
      global: {
        plugins: [router],
        stubs: {
          RouterLink: {
            template: '<a><slot /></a>'
          }
        }
      }
    })

    expect(wrapper.text()).toContain('设备管理')
    expect(wrapper.text()).toContain('计量工作台')
    expect(wrapper.text()).toContain('标定工作台')
    expect(wrapper.text()).toContain('多设备打压')
    expect(wrapper.text()).toContain('进入模块')
  })
})
