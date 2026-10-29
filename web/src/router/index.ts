import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'
import type { Role } from '@/api/types'
import MainLayout from '@/layout/MainLayout.vue'

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    icon?: string
    /** 页头下方的说明文字 */
    desc?: string
    /** 侧边栏分组，不填则作为一级菜单 */
    group?: string
    /** 访问该页面所需的最低角色 */
    role?: Role
  }
}

export const menuRoutes: RouteRecordRaw[] = [
  { path: 'dashboard', name: 'dashboard', component: () => import('@/views/Dashboard.vue'), meta: { title: '概览', icon: 'DataBoard' } },
  { path: 'players', name: 'players', component: () => import('@/views/Players.vue'), meta: { title: '玩家管理', desc: '按区服查询玩家，进行封禁、踢下线和资源调整', icon: 'User', group: '运营管理' } },
  { path: 'mails', name: 'mails', component: () => import('@/views/Mails.vue'), meta: { title: '邮件发放', desc: '向指定玩家或区服发送系统邮件，可附带道具奖励', icon: 'Message', group: '运营管理' } },
  { path: 'announcements', name: 'announcements', component: () => import('@/views/Announcements.vue'), meta: { title: '公告管理', desc: '编辑跑马灯公告，并推送到指定区服', icon: 'Bell', group: '运营管理' } },
  { path: 'servers', name: 'servers', component: () => import('@/views/Servers.vue'), meta: { title: '区服管理', desc: '维护区服信息与 GM 接口地址，查看实时在线人数', icon: 'Monitor', group: '服务管理' } },
  { path: 'admins', name: 'admins', component: () => import('@/views/Admins.vue'), meta: { title: '账号管理', desc: '管理后台账号、角色权限与登录状态', icon: 'Avatar', group: '系统管理', role: 'super_admin' } },
  { path: 'oplogs', name: 'oplogs', component: () => import('@/views/OpLogs.vue'), meta: { title: '操作日志', desc: '所有写操作的审计记录，保留 180 天', icon: 'Document', group: '系统管理', role: 'admin' } },
]

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: () => import('@/views/Login.vue'), meta: { title: '登录' } },
    { path: '/', component: MainLayout, redirect: '/dashboard', children: menuRoutes },
    { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
  ],
})

router.beforeEach(async (to) => {
  const store = useUserStore()
  if (to.name === 'login') return true
  if (!store.token) return { name: 'login', query: { redirect: to.fullPath } }
  if (!store.user) {
    try {
      await store.fetchProfile()
    } catch {
      store.logout()
      return { name: 'login', query: { redirect: to.fullPath } }
    }
  }
  if (to.meta.role && !store.hasRole(to.meta.role)) return { name: 'dashboard' }
  return true
})

router.afterEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} - 游戏管理后台` : '游戏管理后台'
})

export default router
