import { createRouter, createWebHistory } from 'vue-router'
import { userRoutes, userRouteGuard } from './user'

/**
 * 路由配置
 * 使用动态导入实现代码分割和懒加载
 * 按功能模块分组，优化加载性能
 */

// 布局组件 - 预加载
const MainLayout = () => import(/* webpackChunkName: "layout" */ '../layouts/MainLayout.vue')

// 认证相关 - 独立 chunk
const Login = () => import(/* webpackChunkName: "auth" */ '../views/Login.vue')
const Register = () => import(/* webpackChunkName: "auth" */ '../views/Register.vue')

// 核心页面 - 优先加载
const Dashboard = () => import(/* webpackChunkName: "core" */ '../views/Dashboard.vue')
const Profile = () => import(/* webpackChunkName: "core" */ '../views/Profile.vue')
const ChangePassword = () => import(/* webpackChunkName: "core" */ '../views/ChangePassword.vue')

// 代理管理 - 按需加载
const Inbounds = () => import(/* webpackChunkName: "proxy" */ '../views/Inbounds.vue')

// 用户管理 - 按需加载
const Users = () => import(/* webpackChunkName: "users" */ '../views/Users.vue')
const Roles = () => import(/* webpackChunkName: "users" */ '../views/Roles.vue')

// 监控统计 - 按需加载
const SystemMonitor = () => import(/* webpackChunkName: "monitor" */ '../views/SystemMonitor.vue')
const TrafficMonitor = () => import(/* webpackChunkName: "monitor" */ '../views/TrafficMonitor.vue')
const Stats = () => import(/* webpackChunkName: "monitor" */ '../views/Stats.vue')

// 系统管理 - 按需加载
const Settings = () => import(/* webpackChunkName: "system" */ '../views/Settings.vue')
const Certificates = () => import(/* webpackChunkName: "system" */ '../views/Certificates.vue')
const Logs = () => import(/* webpackChunkName: "system" */ '../views/Logs.vue')
const IPRestriction = () => import(/* webpackChunkName: "system" */ '../views/IPRestriction.vue')

// 订阅管理 - 按需加载
const Subscription = () => import(/* webpackChunkName: "subscription" */ '../views/Subscription.vue')
const AdminSubscriptions = () => import(/* webpackChunkName: "subscription" */ '../views/AdminSubscriptions.vue')

// 商业化管理 - 按需加载
const AdminPlans = () => import(/* webpackChunkName: "commercial-admin" */ '../views/AdminPlans.vue')
const AdminOrders = () => import(/* webpackChunkName: "commercial-admin" */ '../views/AdminOrders.vue')
const AdminCoupons = () => import(/* webpackChunkName: "commercial-admin" */ '../views/AdminCoupons.vue')
const AdminReports = () => import(/* webpackChunkName: "commercial-admin" */ '../views/AdminReports.vue')
const AdminGiftCards = () => import(/* webpackChunkName: "commercial-admin" */ '../views/AdminGiftCards.vue')
const AdminTrials = () => import(/* webpackChunkName: "commercial-admin" */ '../views/AdminTrials.vue')

// 错误页面
const NotFound = () => import(/* webpackChunkName: "error" */ '../views/NotFound.vue')

const routes = [
  // 用户前台路由
  ...userRoutes,
  
  {
    path: '/',
    component: MainLayout,
    children: [
      // 核心页面
      {
        path: '',
        name: 'Dashboard',
        component: Dashboard,
        meta: { 
          requiresAuth: true,
          title: '仪表盘'
        }
      },
      {
        path: 'profile',
        name: 'Profile',
        component: Profile,
        meta: { 
          requiresAuth: true,
          title: '个人资料'
        }
      },
      {
        path: 'change-password',
        name: 'ChangePassword',
        component: ChangePassword,
        meta: { 
          requiresAuth: true,
          title: '修改密码'
        }
      },
      
      // 设备管理
      {
        path: 'devices',
        name: 'Devices',
        component: () => import(/* webpackChunkName: "user" */ '../views/Devices.vue'),
        meta: {
          requiresAuth: true,
          title: '我的设备'
        }
      },
      
      // 代理管理
      {
        path: 'inbounds',
        name: 'Inbounds',
        component: Inbounds,
        meta: { 
          requiresAuth: true,
          title: '入站管理'
        }
      },
      
      // 订阅管理
      {
        path: 'subscription',
        name: 'Subscription',
        component: Subscription,
        meta: { 
          requiresAuth: true,
          title: '订阅管理'
        }
      },
      {
        path: 'admin/subscriptions',
        name: 'AdminSubscriptions',
        component: AdminSubscriptions,
        meta: {
          requiresAuth: true,
          title: '订阅管理（管理员）',
          roles: ['admin']
        }
      },
      
      // 用户管理
      {
        path: 'users',
        name: 'Users',
        component: Users,
        meta: { 
          requiresAuth: true,
          title: '用户管理',
          roles: ['admin']
        }
      },
      {
        path: 'roles',
        name: 'Roles',
        component: Roles,
        meta: { 
          requiresAuth: true,
          title: '角色管理',
          roles: ['admin']
        }
      },
      
      // 监控统计
      {
        path: 'system-monitor',
        name: 'SystemMonitor',
        component: SystemMonitor,
        meta: { 
          requiresAuth: true,
          title: '系统监控'
        }
      },
      {
        path: 'traffic-monitor',
        name: 'TrafficMonitor',
        component: TrafficMonitor,
        meta: { 
          requiresAuth: true,
          title: '流量监控'
        }
      },
      {
        path: 'stats',
        name: 'Stats',
        component: Stats,
        meta: { 
          requiresAuth: true,
          title: '统计数据'
        }
      },
      
      // 系统管理
      {
        path: 'settings',
        name: 'Settings',
        component: Settings,
        meta: {
          requiresAuth: true,
          title: '系统设置',
          roles: ['admin']
        }
      },
      {
        path: 'certificates',
        name: 'Certificates',
        component: Certificates,
        meta: {
          requiresAuth: true,
          title: '证书管理',
          roles: ['admin']
        }
      },
      {
        path: 'logs',
        name: 'Logs',
        component: Logs,
        meta: {
          requiresAuth: true,
          title: '日志管理',
          roles: ['admin']
        }
      },
      {
        path: 'ip-restriction',
        name: 'IPRestriction',
        component: IPRestriction,
        meta: {
          requiresAuth: true,
          title: 'IP 限制管理',
          roles: ['admin']
        }
      },
      // 商业化管理
      {
        path: 'admin/plans',
        name: 'AdminPlans',
        component: AdminPlans,
        meta: {
          requiresAuth: true,
          title: '套餐管理',
          roles: ['admin']
        }
      },
      {
        path: 'admin/orders',
        name: 'AdminOrders',
        component: AdminOrders,
        meta: {
          requiresAuth: true,
          title: '订单管理',
          roles: ['admin']
        }
      },
      {
        path: 'admin/coupons',
        name: 'AdminCoupons',
        component: AdminCoupons,
        meta: {
          requiresAuth: true,
          title: '优惠券管理',
          roles: ['admin']
        }
      },
      {
        path: 'admin/reports',
        name: 'AdminReports',
        component: AdminReports,
        meta: {
          requiresAuth: true,
          title: '财务报表',
          roles: ['admin']
        }
      },
      {
        path: 'admin/gift-cards',
        name: 'AdminGiftCards',
        component: AdminGiftCards,
        meta: {
          requiresAuth: true,
          title: '礼品卡管理',
          roles: ['admin']
        }
      },
      {
        path: 'admin/trials',
        name: 'AdminTrials',
        component: AdminTrials,
        meta: {
          requiresAuth: true,
          title: '试用管理',
          roles: ['admin']
        }
      }
    ]
  },
  
  // 认证页面
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { 
      title: '登录',
      guest: true
    }
  },
  {
    path: '/register',
    name: 'Register',
    component: Register,
    meta: { 
      title: '注册',
      guest: true
    }
  },
  
  // 404 页面
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: NotFound,
    meta: { title: '页面未找到' }
  }
]

// 创建路由实例
const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    }
    return { top: 0 }
  }
})

// 全局前置守卫
router.beforeEach((to, from, next) => {
  // 用户前台路由使用专门的守卫
  if (to.path.startsWith('/user')) {
    userRouteGuard(to, from, next)
    return
  }
  
  const isAuthenticated = localStorage.getItem('token')
  const userRole = localStorage.getItem('userRole') || 'user'
  
  // 更新页面标题
  if (to.meta.title) {
    document.title = `${to.meta.title} - V Panel`
  }
  
  // 需要认证的页面
  if (to.meta.requiresAuth && !isAuthenticated) {
    next({ name: 'Login', query: { redirect: to.fullPath } })
    return
  }
  
  // 已登录用户访问登录/注册页面
  if (to.meta.guest && isAuthenticated) {
    next('/')
    return
  }
  
  // 角色权限检查
  if (to.meta.roles && !to.meta.roles.includes(userRole)) {
    next({ name: 'Dashboard' })
    return
  }
  
  next()
})

// 全局后置钩子 - 用于预加载
router.afterEach((to) => {
  // 预加载可能访问的下一个页面
  if (to.name === 'Dashboard') {
    // 预加载常用页面
    import(/* webpackChunkName: "proxy" */ '../views/Inbounds.vue')
    import(/* webpackChunkName: "monitor" */ '../views/SystemMonitor.vue')
  }
})

export default router
