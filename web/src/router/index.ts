import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('@/views/LoginView.vue') },
    {
      path: '/portfolio',
      component: () => import('@/views/PortfolioView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/stock/:tsCode',
      component: () => import('@/views/StockDetailView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/settings',
      component: () => import('@/views/SettingsView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/admin/users',
      component: () => import('@/views/admin/UsersView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
    },
    {
      path: '/admin/sync',
      component: () => import('@/views/admin/SyncView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
    },
    { path: '/', redirect: '/portfolio' },
  ],
})

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.user) {
    next('/login')
  } else if (to.meta.requiresAdmin && auth.user?.role !== 'admin') {
    next('/portfolio')
  } else {
    next()
  }
})

export default router
