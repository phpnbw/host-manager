import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import App from './App.vue'
import router from './router'
import { useAuthStore } from './stores/auth'

console.log('=== Vue.js 应用初始化开始 ===')

try {
  console.log('1. 创建 Vue 应用实例...')
  const app = createApp(App)
  
  console.log('2. 创建 Pinia 状态管理...')
  const pinia = createPinia()

  console.log('3. 注册 Element Plus 图标...')
  // 注册Element Plus图标
  for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component)
  }

  console.log('4. 安装插件...')
  app.use(pinia)
  app.use(router)
  app.use(ElementPlus)

  console.log('5. 初始化认证状态...')
  // 初始化认证状态
  const authStore = useAuthStore()
  authStore.initAuth()

  console.log('6. 挂载应用到 DOM...')
  app.mount('#app')
  
  console.log('✅ Vue.js 应用初始化完成')
} catch (error) {
  console.error('❌ Vue.js 应用初始化失败:', error)
} 