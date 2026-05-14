import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

import MeasurementDataView from '../MeasurementDataView.vue'

const ElIconStub = defineComponent({
  name: 'ElIcon',
  template: '<i><slot /></i>'
})

describe('MeasurementDataView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders raw sampling rows even when no pressure targets are configured', () => {
    const wrapper = mount(MeasurementDataView, {
      props: {
        rows: [
          {
            timestamp: '2026-04-23T10:00:01Z',
            channels: { '1': 10.1, '2': 10.2 }
          }
        ],
        channels: [1, 2],
        targets: [],
        averageCount: 1,
        state: 'collecting'
      },
      global: {
        stubs: {
          ElIcon: ElIconStub
        }
      }
    })

    expect(wrapper.text()).toContain('1 条采样')
    expect(wrapper.text()).toContain('10.100')
    expect(wrapper.text()).toContain('10.200')
    expect(wrapper.text()).not.toContain('请先配置参数并生成压力表')
  })
})
