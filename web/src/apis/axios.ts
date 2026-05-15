import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import { useLangStore } from '@/stores/lang'
import router from '@/router'
import { wMessage } from '@/utils/message'

const $http = axios.create({
  baseURL: '/api',
  headers: { 'Content-Type': 'application/json' },
})

$http.interceptors.request.use((cfg) => {
  const lang = useLangStore().lang
  cfg.headers.lang = lang
  return cfg
})

$http.interceptors.response.use(
  (res) => {
    if (res.data && typeof res.data.code === 'number') {
      if (res.data.code !== 200) {
        wMessage('error', res.data.message || 'unknown error')
        return Promise.reject(new Error(res.data.message || 'unknown error'))
      }
      return res.data.data
    }
    return res.data
  },
  (err) => {
    if (err.response?.status === 401) {
      useAuthStore().logout()
      router.push('/login')
    } else {
      wMessage('error', err.message || '网络错误')
    }
    return Promise.reject(err)
  }
)

export { $http }
