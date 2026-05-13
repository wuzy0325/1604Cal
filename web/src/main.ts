import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import 'element-plus/dist/index.css'

import App from './App.vue'
import router from './router'
import { initDesktopApiBase, createEventStream } from './api/client'
import { EVENT_HARDWARE_COMMAND, EVENT_HARDWARE_RESPONSE } from './shared/events'
import { useHardwareLogStore } from './stores/hardwareLog'
import type { StreamEventPayload } from './types/api'
import './styles/global.scss'

async function bootstrap() {
  // 在桌面模式下，初始化 API 基础路径指向内嵌 HTTP 服务器。
  await initDesktopApiBase()

  const app = createApp(App)

  // 注册所有图标
  for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component)
  }

  const pinia = createPinia()
  app.use(pinia)
  app.use(router)
  app.use(ElementPlus)

  // 全局监听硬件通讯事件
  const hardwareLog = useHardwareLogStore()
  createEventStream((payload: StreamEventPayload) => {
    if (payload.type === EVENT_HARDWARE_COMMAND) {
      const data = payload.data as { model?: string; proto?: string; cmd?: string }
      hardwareLog.addEntry('hw-cmd', data?.model ?? '', data?.proto ?? '', data?.cmd ?? '')
    }
    if (payload.type === EVENT_HARDWARE_RESPONSE) {
      const data = payload.data as { model?: string; proto?: string; resp?: string; cmd?: string }
      const detail = data?.resp ?? ''
      hardwareLog.addEntry('hw-res', data?.model ?? '', data?.proto ?? '', detail.length > 200 ? detail.slice(0, 200) + '...' : detail)
    }
  })

  app.mount('#app')
}

bootstrap()
