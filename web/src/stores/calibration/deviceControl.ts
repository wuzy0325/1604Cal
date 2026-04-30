import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import {
  bindMeasureDevice,
  bindDevices,
  readPressure,
  readStability,
  readMeasureData,
  readValveStatus,
  setValveStatus as apiSetValveStatus,
  readMeasureUnit,
  setMeasureUnit as apiSetMeasureUnit,
  readDeviceInfo,
  resetDevice as apiResetDevice
} from "@/api/session"
import {
  multipressRegister,
  multipressUnregister,
  multipressListDevices
} from "@/api/multipress"
import { useDeviceInventoryStore } from '@/stores/device/inventoryStore'

export const useDeviceControlStore = defineStore('deviceControl', () => {
  const deviceStore = useDeviceInventoryStore()

  // State
  const currentPressure = ref(0)
  const isStable = ref(false)
  const channelData = ref<number[]>([])
  const valveStatus = ref<string>('')
  const measureUnit = ref<string>('')
  const deviceInfo = ref<Record<string, string>>({})

  // Getters
  const device1604Connected = computed(() => {
    return deviceStore.measureDevices.some(d => d.status === 'connected')
  })

  const pressDeviceConnected = computed(() => {
    return deviceStore.pressureDevices.some(d => d.status === 'connected')
  })

  // Actions
  const sleep = (ms: number) => new Promise<void>((resolve) => {
    setTimeout(resolve, ms)
  })

  // 连接1604设备
  const connectDevice1604 = async (deviceId: string) => {
    try {
      const connected = await deviceStore.connectMeasureDevice(deviceId)
      if (!connected) {
        return
      }

      // 连接成功后立即绑定到设备会话，使其能读取阀门/单位/设备信息
      try {
        await bindMeasureDevice(deviceId)
      } catch (err) {
        ElMessage.error('绑定计量设备会话失败，无法读取阀门/单位信息')
        console.error('bindMeasureDevice failed:', err)
        return
      }

      // 读取设备信息、阀门状态和单位（连接后增加重试，避免设备刚建链时读数失败）
      const loaded = await refreshDeviceInfo({ retries: 3, retryDelayMs: 500 })
      if (!loaded) {
        ElMessage.warning('设备已连接，但阀门/单位信息读取失败，请稍后重试')
      }

      return true
    } catch (error) {
      console.error('连接1604设备失败:', error)
      return false
    }
  }

  // 断开1604设备
  const disconnectDevice1604 = async (deviceId: string) => {
    try {
      await deviceStore.disconnectMeasureDevice(deviceId)
    } catch (error) {
      console.error('断开1604设备失败:', error)
    }
  }

  // 连接打压设备：通过 multipress 服务注册（创建驱动 + TCP连接 + 注册到压力控制模块）
  const connectPressDevice = async (deviceId: string) => {
    try {
      await multipressRegister(deviceId)
      // multipress 服务不更新 DeviceManager 状态，需手动同步前端 store
      deviceStore.updateDeviceStatus(deviceId, 'connected')
      // 从 multipress 服务拉取实际读取到的单位与压力，覆盖配置中的默认值
      try {
        const states = await multipressListDevices()
        const state = states.find(s => s.deviceId === deviceId)
        const dev = deviceStore.pressureDevices.find(d => d.id === deviceId)
        if (state && dev) {
          if (state.unit) dev.unit = state.unit
          dev.currentPressure = state.currentPressure
        }
      } catch {
        // 静默失败，使用设备配置中的单位
      }
    } catch (error) {
      console.error('连接打压设备失败:', error)
      ElMessage.error('连接打压设备失败')
    }
  }

  // 断开打压设备：通过 multipress 服务注销（停止控制 + 断开TCP + 移除注册）
  const disconnectPressDevice = async (deviceId: string) => {
    try {
      await multipressUnregister(deviceId)
      deviceStore.updateDeviceStatus(deviceId, 'disconnected')
    } catch (error) {
      console.error('断开打压设备失败:', error)
      ElMessage.error('断开打压设备失败')
    }
  }

  // 设置校准设备（通知后端）
  const setDevices = async (measureDeviceId: string, pressureDeviceId: string) => {
    try {
      await bindDevices(measureDeviceId, pressureDeviceId)
    } catch (error) {
      console.error('绑定设备会话失败:', error)
      ElMessage.error('绑定设备会话失败')
    }
  }

  // 读取实时压力
  const refreshPressure = async () => {
    try {
      currentPressure.value = await readPressure()
    } catch {
      // 静默失败
    }
  }

  // 读取稳定状态
  const refreshStability = async () => {
    try {
      isStable.value = await readStability()
    } catch {
      // 静默失败
    }
  }

  // 读取计量设备数据
  const refreshMeasureData = async () => {
    try {
      channelData.value = await readMeasureData()
    } catch {
      // 静默失败
    }
  }

  // 刷新设备信息（阀门状态、单位、设备信息）
  const refreshDeviceInfo = async (
    options: { retries?: number; retryDelayMs?: number } = {}
  ): Promise<boolean> => {
    const retries = Math.max(1, options.retries ?? 1)
    const retryDelayMs = Math.max(0, options.retryDelayMs ?? 0)
    let valveReady = false
    let unitReady = false

    const tryRead = async <T>(reader: () => Promise<T>): Promise<PromiseSettledResult<T>> => {
      try {
        return { status: 'fulfilled', value: await reader() }
      } catch (error) {
        return { status: 'rejected', reason: error }
      }
    }

    for (let attempt = 1; attempt <= retries; attempt++) {
      const [valve, unit] = await Promise.allSettled([
        readValveStatus(),
        readMeasureUnit()
      ])

      if (valve.status === 'fulfilled') {
        valveStatus.value = valve.value
        valveReady = true
      }
      if (unit.status === 'fulfilled') {
        measureUnit.value = unit.value
        unitReady = true
      }
      // 连接阶段只要求阀门和单位可读；且允许在不同重试轮次分别成功。
      if (valveReady && unitReady) {
        // 设备信息是增强信息，避免其失败拖累阀门/单位可用性判定。
        const info = await tryRead(readDeviceInfo)
        if (info.status === 'fulfilled') {
          deviceInfo.value = info.value
        }
        return true
      }

      if (attempt < retries && retryDelayMs > 0) {
        await sleep(retryDelayMs)
      }
    }

    return false
  }

  // 读取阀门状态
  const refreshValveStatus = async () => {
    try {
      valveStatus.value = await readValveStatus()
    } catch {
      // 静默失败
    }
  }

  // 设置阀门状态
  const setValveStatus = async (status: string) => {
    try {
      await apiSetValveStatus(status)
      valveStatus.value = status
      ElMessage.success(status === 'calibration' ? '阀门已切换到校准模式' : '阀门已切换到测量模式')
    } catch (error) {
      console.error('设置阀门状态失败:', error)
      ElMessage.error('设置阀门状态失败')
    }
  }

  // 读取计量设备单位
  const refreshMeasureUnit = async () => {
    try {
      measureUnit.value = await readMeasureUnit()
    } catch {
      // 静默失败
    }
  }

  // 设置计量设备单位
  const setMeasureUnit = async (unit: string) => {
    try {
      await apiSetMeasureUnit(unit)
      measureUnit.value = await readMeasureUnit()
      ElMessage.success(`单位已切换为 ${measureUnit.value || unit}`)
    } catch (error) {
      console.error('设置计量设备单位失败:', error)
      ElMessage.error('设置计量设备单位失败')
    }
  }

  // 复位计量设备
  const resetDevice = async () => {
    try {
      await apiResetDevice()
      ElMessage.success('设备已复位')
    } catch (error) {
      console.error('复位设备失败:', error)
      ElMessage.error('复位设备失败')
    }
  }

  return {
    // State
    currentPressure,
    isStable,
    channelData,
    valveStatus,
    measureUnit,
    deviceInfo,
    // Getters
    device1604Connected,
    pressDeviceConnected,
    // Actions
    connectDevice1604,
    disconnectDevice1604,
    connectPressDevice,
    disconnectPressDevice,
    setDevices,
    refreshPressure,
    refreshStability,
    refreshMeasureData,
    refreshDeviceInfo,
    refreshValveStatus,
    setValveStatus,
    refreshMeasureUnit,
    setMeasureUnit,
    resetDevice
  }
})
