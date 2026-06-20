import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  { path: '/login', name: 'login', component: () => import('@/pages/Login.vue'), meta: { guest: true, title: '登录' } },
  { path: '/', name: 'dashboard', component: () => import('@/pages/Dashboard.vue'), meta: { title: '运行驾驶舱' } },
  { path: '/areas', name: 'areas', component: () => import('@/pages/Areas.vue'), meta: { title: '台区管理' } },
  { path: '/areas/new', name: 'area-new', component: () => import('@/pages/AreaEdit.vue'), meta: { roles: ['station', 'admin'], title: '录入台区' } },
  { path: '/areas/:id', name: 'area-detail', component: () => import('@/pages/AreaDetail.vue'), meta: { title: '台区详情' } },
  { path: '/areas/:id/edit', name: 'area-edit', component: () => import('@/pages/AreaEdit.vue'), meta: { roles: ['station', 'admin'], title: '编辑台区' } },
  { path: '/declarations', name: 'declarations', component: () => import('@/pages/Declarations.vue'), meta: { title: '并网申报' } },
  {
    path: '/declarations/new',
    name: 'declaration-new',
    component: () => import('@/pages/DeclarationEdit.vue'),
    meta: { roles: ['owner', 'admin'], title: '提交申报' },
  },
  {
    path: '/declarations/:id/approve',
    name: 'declaration-approve',
    component: () => import('@/pages/DeclarationApprove.vue'),
    meta: { roles: ['station', 'admin'], title: '并网审批' },
  },
  { path: '/alarms', name: 'alarms', component: () => import('@/pages/Alarms.vue'), meta: { title: '反送电告警' } },
  { path: '/alarms/:id', name: 'alarm-detail', component: () => import('@/pages/AlarmDetail.vue'), meta: { title: '告警处理' } },
  { path: '/limits', name: 'limits', component: () => import('@/pages/Limits.vue'), meta: { title: '限发执行' } },
  {
    path: '/limits/new',
    name: 'limit-new',
    component: () => import('@/pages/LimitEdit.vue'),
    meta: { roles: ['dispatcher', 'admin'], title: '发布限发' },
  },
  { path: '/limits/:id', name: 'limit-detail', component: () => import('@/pages/LimitDetail.vue'), meta: { title: '限发详情' } },
  { path: '/analysis', name: 'analysis', component: () => import('@/pages/Analysis.vue'), meta: { title: '消纳分析' } },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({ history: createWebHistory(), routes })

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.guest) {
    if (auth.isLoggedIn) return { name: 'dashboard' }
    return true
  }
  if (!auth.isLoggedIn) return { name: 'login', query: { redirect: to.fullPath } }
  const roles = to.meta.roles as string[] | undefined
  if (roles && !roles.includes(auth.role)) return { name: 'dashboard' }
  return true
})

export default router
