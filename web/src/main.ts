import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import 'element-plus/dist/index.css'

import App from './App.vue'
import router from './router'
import { initDesktopApiBase } from './api/client'
import './styles/variables.scss'
import './styles/button-override.css'

async function bootstrap() {
  // 在桌面模式下，初始化 API 基础路径指向内嵌 HTTP 服务器。
  await initDesktopApiBase()

  const app = createApp(App)

  // 注册所有图标
  for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component)
  }

  app.use(createPinia())
  app.use(router)
  app.use(ElementPlus)
  app.mount('#app')
}

bootstrap()
