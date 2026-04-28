import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

import MeasurementControl from '../MeasurementControl.vue'
import { useMeasurementStore } from '@/stores/measurement'

const ElIconStub = defineComponent({
  name: 'ElIcon',
  template: '<i><slot /></i>'
})

function mountControl() {
  return mount(MeasurementControl, {
    props: {
      channels: [1, 2],
      isStable: false,
      stableSeconds: 0,
      selectedChannelCount: 2
    },
    global: {
      stubs: {
        ElIcon: ElIconStub
      }
    }
  })
}

describe('MeasurementControl', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('shows pressurizing state text and allows pause while running', () => {
    const store = useMeasurementStore()
    store.measureDeviceId = 'measure-1'
    store.syncState('pressurizing')

    const wrapper = mountControl()

    expect(wrapper.text()).toContain('打压中')
    const pauseButton = wrapper.findAll('button').find(btn => btn.text().includes('暂停'))
    expect(pauseButton).toBeTruthy()
    expect(pauseButton?.attributes('disabled')).toBeUndefined()
  })

  it('allows starting again from completed state when device is bound', () => {
    const store = useMeasurementStore()
    store.measureDeviceId = 'measure-1'
    store.syncState('completed')

    const wrapper = mountControl()

    const startButton = wrapper.findAll('button').find(btn => btn.text().includes('开始采样'))
    expect(startButton).toBeTruthy()
    expect(startButton?.attributes('disabled')).toBeUndefined()
  })

  it('shows a sampling-focused toolbar without pseudo-workbench controls', () => {
    const wrapper = mountControl()

    expect(wrapper.text()).toContain('已选通道')
    expect(wrapper.text()).not.toContain('控制模式')
    expect(wrapper.text()).not.toContain('打压模式')
    expect(wrapper.text()).not.toContain('报警设置')
    expect(wrapper.text()).not.toContain('进度')
  })
})
