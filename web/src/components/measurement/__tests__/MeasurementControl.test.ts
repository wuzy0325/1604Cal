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

  it('shows pause button enabled while running', () => {
    const store = useMeasurementStore()
    store.measureDeviceId = 'measure-1'
    store.syncState('pressurizing')

    const wrapper = mountControl()

    const pauseButton = wrapper.findAll('button').find(btn => btn.text().includes('暂停'))
    expect(pauseButton).toBeTruthy()
    expect(pauseButton?.attributes('disabled')).toBeUndefined()
  })

  it('shows start button enabled when completed and device bound', () => {
    const store = useMeasurementStore()
    store.measureDeviceId = 'measure-1'
    store.syncState('completed')

    const wrapper = mountControl()

    const startButton = wrapper.findAll('button').find(btn => btn.text().includes('开始'))
    expect(startButton).toBeTruthy()
    expect(startButton?.attributes('disabled')).toBeUndefined()
  })

  it('renders mode toggles and channel selector', () => {
    const wrapper = mountControl()

    expect(wrapper.text()).toContain('采集通道')
    expect(wrapper.text()).toContain('模式')
    expect(wrapper.text()).toContain('报警设置')
    expect(wrapper.text()).toContain('任务进度')
  })
})
