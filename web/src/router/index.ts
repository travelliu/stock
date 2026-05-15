import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('@/views/LoginView.vue') },
    { path: '/', redirect: '/stocks' },
    {
      path: '/stocks',
      component: () => import('@/views/StockListView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/stocks/:code',
      component: () => import('@/views/stock/StockDetailView.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', name: 'StockPrediction', component: () => import('@/views/stock/PredictionTab.vue') },
        { path: 'bars', name: 'StockBars', component: () => import('@/views/stock/DailyBarsTab.vue') },
        { path: 'predictions', name: 'StockPredictions', component: () => import('@/views/stock/PredictionRecordsTab.vue') },
      ],
    },
    {
      path: '/profile',
      component: () => import('@/views/profile/ProfileView.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', component: () => import('@/views/profile/ProfileInfo.vue') },
        { path: 'password', component: () => import('@/views/profile/ChangePassword.vue') },
        { path: 'token', component: () => import('@/views/profile/TushareToken.vue') },
        { path: 'api-tokens', component: () => import('@/views/profile/ApiTokens.vue') },
      ],
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
    { path: '/:pathMatch(.*)*', component: () => import('@/views/NotFound.vue') },
  ],
})

router.beforeEach(async (to, _from, next) => {
  const auth = useAuthStore()
  if (!auth.initialized) {
    await auth.fetchMe()
  }
  if (to.meta.requiresAuth && !auth.user) {
    next('/login')
  } else if (to.meta.requiresAdmin && auth.user?.role !== 'admin') {
    next('/stocks')
  } else {
    next()
  }
})

export default router
