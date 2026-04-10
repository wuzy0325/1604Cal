import { defineStore } from 'pinia'

export type ModuleKey = 'measurement' | 'calibration'

export interface ModuleDeviceSelection {
  measureDeviceId: string
  pressureDeviceId: string
}

export interface DeviceState {
  connectedCount: number
  selections: Record<ModuleKey, ModuleDeviceSelection>
}

export const useDeviceStore = defineStore('device', {
  state: (): DeviceState => ({
    connectedCount: 0,
    selections: {
      measurement: {
        measureDeviceId: '',
        pressureDeviceId: ''
      },
      calibration: {
        measureDeviceId: '',
        pressureDeviceId: ''
      }
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

    setModuleSelection(module: ModuleKey, selection: Partial<ModuleDeviceSelection>) {
      this.selections[module] = {
        ...this.selections[module],
        ...selection
      }
    }
  }
})
