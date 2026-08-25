import { defineStore } from 'pinia'

export type ModuleKey = 'measurement' | 'calibration'

export interface ModuleDeviceSelection {
  /** 单设备兼容字段：始终等于 measureDeviceIds[0] */
  measureDeviceId: string
  /** 多设备勾选列表（保持用户勾选顺序，后端绑定顺序与此一致） */
  measureDeviceIds: string[]
  pressureDeviceId: string
}

export interface DeviceState {
  connectedCount: number
  selections: Record<ModuleKey, ModuleDeviceSelection>
}

function emptySelection(): ModuleDeviceSelection {
  return { measureDeviceId: '', measureDeviceIds: [], pressureDeviceId: '' }
}

export const useDeviceStore = defineStore('device', {
  state: (): DeviceState => ({
    connectedCount: 0,
    selections: {
      measurement: emptySelection(),
      calibration: emptySelection()
    }
  }),

  getters: {
    selectionByModule: (state) => {
      return (module: ModuleKey): ModuleDeviceSelection => state.selections[module]
    }
  },

  actions: {
    setConnectedCount(count: number) {
      this.connectedCount = count
    },

    // measureDeviceIds 与 measureDeviceId 双向同步：
    // 传数组时单设备字段取首元素，传单设备字段时包装为数组，保证两个视图读取一致。
    setModuleSelection(module: ModuleKey, selection: Partial<ModuleDeviceSelection>) {
      const merged: ModuleDeviceSelection = {
        ...this.selections[module],
        ...selection
      }
      if (selection.measureDeviceIds !== undefined) {
        merged.measureDeviceId = selection.measureDeviceIds[0] ?? ''
      } else if (selection.measureDeviceId !== undefined) {
        merged.measureDeviceIds = selection.measureDeviceId ? [selection.measureDeviceId] : []
      }
      this.selections[module] = merged
    }
  }
})
