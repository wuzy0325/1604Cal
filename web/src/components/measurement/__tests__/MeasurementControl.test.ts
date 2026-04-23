import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

import MeasurementControl from '../MeasurementControl.vue'
import { useMeasurementStore } from '@/stores/measurement'

const ElButtonStub = defineComponent({
  name: 'ElButton',
  props: {
    disabled: {
      type: Boolean,
      default: false
    }
  },
  emits: ['click'],
  template: '<button :disabled="disabled" @click="$emit(\'click\')"><slot /></button>'
})

const ElInputNumberStub = defineComponent({
  name: 'ElInputNumber',
  template: '<div class="el-input-number-stub" />'
})

const ElIconStub = defineComponent({
  name: 'ElIcon',
  template: '<i><slot /></i>'
})

function mountControl() {
  return mount(MeasurementControl, {
    props: {
      channels: [1, 2]
    },
    global: {
      stubs: {
        ElButton: ElButtonStub,
        ElInputNumber: ElInputNumberStub,
        ElIcon: ElIconStub
      }
    }
  })
}

describe('MeasurementControl', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('shows pressuring state text and allows pause while running', () => {
    const store = useMeasurementStore()
    store.measureDeviceId = 'measure-1'
    store.syncState('pressuring')

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

    const startButton = wrapper.findAll('button').find(btn => btn.text().includes('开始采集'))
    expect(startButton).toBeTruthy()
    expect(startButton?.attributes('disabled')).toBeUndefined()
  })
})
