/* eslint-disable @typescript-eslint/no-explicit-any */
import { computed, defineComponent, h, inject, provide, ref } from 'vue'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import CalibrationDataView from '../CalibrationDataView.vue'
import { useCalibrationStore } from '@/stores/calibration'
import { usePressurePointStore } from '@/stores/calibration/pressurePoints'

type RowData = Record<string, unknown>

const tableRowsKey = Symbol('tableRows')

const ElTableStub = defineComponent({
  name: 'ElTable',
  props: {
    data: {
      type: Array,
      default: () => []
    }
  },
  setup(props: any, { slots }: any) {
    provide(tableRowsKey, computed(() => props.data as RowData[]))
    return () => h('div', { class: 'el-table-stub' }, slots.default ? slots.default() : [])
  }
})

const ElTableColumnStub = defineComponent({
  name: 'ElTableColumn',
  props: {
    label: {
      type: String,
      default: ''
    },
    prop: {
      type: String,
      default: ''
    }
  },
  setup(props: any, { slots }: any) {
    const rows = inject<any>(tableRowsKey, ref<RowData[]>([]))

    return () => h('div', { class: 'el-table-column-stub', 'data-label': props.label }, [
      h('div', { class: 'column-label' }, props.label),
      ...rows.value.map((row: RowData, index: number) => h(
        'div',
        { class: 'column-row', 'data-row-index': String(index) },
        slots.default
          ? slots.default({ row, $index: index })
          : props.prop
            ? String(row[props.prop] ?? '')
            : ''
      ))
    ])
  }
})

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

describe('CalibrationDataView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())

    const calibrationStore = useCalibrationStore()
    const pressurePointStore = usePressurePointStore()

    calibrationStore.selectedChannels = [1, 2]
    pressurePointStore.pressurePoints = [
      {
        id: 'point-1',
        index: 1,
        targetPressure: 10,
        status: 'stabilizing',
        collectedData: [10.12, 10.34],
        actualPressure: 10.2
      },
      {
        id: 'point-2',
        index: 2,
        targetPressure: 20,
        status: 'completed',
        collectedData: [20.12, 20.34],
        actualPressure: 20.2
      }
    ]
  })

  it('renders separate point-operation and collected-data tables', () => {
    const wrapper = mount(CalibrationDataView, {
      global: {
        stubs: {
          ElTable: ElTableStub,
          ElTableColumn: ElTableColumnStub,
          ElButton: ElButtonStub,
          ElTag: { template: '<span><slot /></span>' },
          ElIcon: { template: '<i><slot /></i>' }
        }
      }
    })

    expect(wrapper.find('[data-testid="pressure-point-operation-table"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="collected-data-table"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('压力点设置')
    expect(wrapper.text()).toContain('采集数据')
  })

  it('keeps confirm and collect actions in point-operation rows', async () => {
    const calibrationStore = useCalibrationStore()
    calibrationStore.sessionState = 'ready'
    const confirmSpy = vi.spyOn(calibrationStore as any, 'confirmPressure').mockImplementation(() => {})
    const collectSpy = vi.spyOn(calibrationStore as any, 'collectData').mockResolvedValue(undefined)

    const wrapper = mount(CalibrationDataView, {
      global: {
        stubs: {
          ElTable: ElTableStub,
          ElTableColumn: ElTableColumnStub,
          ElButton: ElButtonStub,
          ElTag: { template: '<span><slot /></span>' },
          ElIcon: { template: '<i><slot /></i>' }
        }
      }
    })

    await wrapper.get('[data-testid="confirm-btn-point-1"]').trigger('click')
    await wrapper.get('[data-testid="collect-btn-point-1"]').trigger('click')

    expect(confirmSpy).toHaveBeenCalledWith('point-1')
    expect(collectSpy).toHaveBeenCalledWith('point-1')
  })

  it('before start, manual mode keeps confirm disabled and hides pressurize action', () => {
    const calibrationStore = useCalibrationStore()
    const pressurePointStore = usePressurePointStore()

    pressurePointStore.pressurePoints = [
      {
        id: 'point-1',
        index: 1,
        targetPressure: 10,
        status: 'pending',
        collectedData: undefined,
        actualPressure: undefined
      }
    ]

    ;(calibrationStore as any).controlMode = 'manual'

    const wrapper = mount(CalibrationDataView, {
      global: {
        stubs: {
          ElTable: ElTableStub,
          ElTableColumn: ElTableColumnStub,
          ElButton: ElButtonStub,
          ElTag: { template: '<span><slot /></span>' },
          ElIcon: { template: '<i><slot /></i>' }
        }
      }
    })

    const confirmButton = wrapper.get('[data-testid="confirm-btn-point-1"]')
    expect(confirmButton.attributes('disabled')).toBe('')
    expect(wrapper.text()).not.toContain('打压')
    expect(wrapper.text()).toContain('待确认')
  })

  it('after start, manual mode allows confirm then collect', async () => {
    const calibrationStore = useCalibrationStore()
    const pressurePointStore = usePressurePointStore()

    pressurePointStore.pressurePoints = [
      {
        id: 'point-1',
        index: 1,
        targetPressure: 10,
        status: 'pending',
        collectedData: undefined,
        actualPressure: undefined
      }
    ]

    ;(calibrationStore as any).controlMode = 'manual'
    calibrationStore.sessionState = 'ready'

    const wrapper = mount(CalibrationDataView, {
      global: {
        stubs: {
          ElTable: ElTableStub,
          ElTableColumn: ElTableColumnStub,
          ElButton: ElButtonStub,
          ElTag: { template: '<span><slot /></span>' },
          ElIcon: { template: '<i><slot /></i>' }
        }
      }
    })

    const collectButtonBeforeConfirm = wrapper.get('[data-testid="collect-btn-point-1"]')
    expect(collectButtonBeforeConfirm.attributes('disabled')).toBe('')

    await wrapper.get('[data-testid="confirm-btn-point-1"]').trigger('click')

    expect(pressurePointStore.pressurePoints[0].status).toBe('stabilizing')
    expect(wrapper.text()).toContain('待采集')
    expect(wrapper.text()).not.toContain('稳定中')

    const collectButtonAfterConfirm = wrapper.get('[data-testid="collect-btn-point-1"]')
    expect(collectButtonAfterConfirm.attributes('disabled')).toBeUndefined()
  })

  it('keeps collect enabled for completed point without retry action', () => {
    const calibrationStore = useCalibrationStore()
    calibrationStore.sessionState = 'ready'

    const wrapper = mount(CalibrationDataView, {
      global: {
        stubs: {
          ElTable: ElTableStub,
          ElTableColumn: ElTableColumnStub,
          ElButton: ElButtonStub,
          ElTag: { template: '<span><slot /></span>' },
          ElIcon: { template: '<i><slot /></i>' }
        }
      }
    })

    const collectButton = wrapper.get('[data-testid="collect-btn-point-2"]')
    expect(collectButton.attributes('disabled')).toBeUndefined()
    expect(wrapper.text()).not.toContain('重采')
  })
})
