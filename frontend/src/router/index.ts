import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useTerminalStore } from '@/stores/terminal'
import Layout from '@/views/Layout.vue'
import Login from '@/views/Login.vue'
import HostList from '@/views/HostList.vue'
import HostDetail from '@/views/HostDetail.vue'
import Terminal from '@/views/Terminal.vue'
import TerminalManager from '@/views/TerminalManager.vue'
import UserManagement from '@/views/UserManagement.vue'
import AuditManagement from '@/views/AuditManagement.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: Layout,
    redirect: '/hosts',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'hosts',
        name: 'HostList',
        component: HostList
      },
      {
        path: 'host/:id',
        name: 'HostDetail',
        component: HostDetail
      },
      {
        path: 'terminal/:hostId',
        name: 'Terminal',
        component: Terminal
      },
      {
        path: 'terminals',
        name: 'TerminalManager',
        component: TerminalManager,
        meta: { keepAlive: true }
      },
      {
        path: 'users',
        name: 'UserManagement',
        component: UserManagement
      },
      {
        path: 'audit',
        name: 'AuditManagement',
        component: AuditManagement
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  console.log('路由守卫:', { to: to.path, from: from.path })
  
  const authStore = useAuthStore()
  
  // 初始化认证状态
  authStore.initAuth()
  
  // 如果离开终端管理页面，设置所有终端为非活动状态
  if (from.path === '/terminals' && to.path !== '/terminals') {
    const terminalStore = useTerminalStore()
    terminalStore.terminals.forEach(terminal => {
      terminalStore.deactivateTerminal(terminal.id)
    })
  }
  
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth !== false)
  const isAuthenticated = authStore.isAuthenticated()
  
  console.log('认证检查:', { requiresAuth, isAuthenticated, path: to.path })
  
  if (requiresAuth && !isAuthenticated) {
    // 需要认证但未登录，跳转到登录页
    console.log('重定向到登录页')
    next('/login')
  } else if (to.path === '/login' && isAuthenticated) {
    // 已登录用户访问登录页，跳转到主页
    console.log('重定向到主页')
    next('/hosts')
  } else {
    console.log('允许访问')
    next()
  }
})

export default router 