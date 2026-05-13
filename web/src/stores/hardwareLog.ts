import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export type HwLogKind = 'hw-cmd' | 'hw-res'

export interface HwLogEntry {
  id: number
  kind: HwLogKind
  timestamp: number
  model: string
  proto: string
  detail: string
}

const MAX_LOG_ENTRIES = 500

export const useHardwareLogStore = defineStore('hardwareLog', () => {
  const entries = ref<HwLogEntry[]>([])
  let nextId = 1

  function addEntry(kind: HwLogKind, model: string, proto: string, detail: string) {
    entries.value.push({
      id: nextId++,
      kind,
      timestamp: Date.now(),
      model,
      proto,
      detail
    })
    if (entries.value.length > MAX_LOG_ENTRIES) {
      entries.value = entries.value.slice(-MAX_LOG_ENTRIES)
    }
  }

  function clear() {
    entries.value = []
  }

  const count = computed(() => entries.value.length)

  return { entries, addEntry, clear, count }
})
