import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const client = axios.create({
  baseURL: '/api',
  headers: { 'Content-Type': 'application/json' },
})

client.interceptors.response.use(
  (res) => {
    if (res.data && typeof res.data.code === 'number') {
      if (res.data.code !== 200) {
        return Promise.reject(new Error(res.data.message || 'unknown error'))
      }
      res.data = res.data.data
    }
    return res
  },
  (err) => {
    if (err.response?.status === 401) {
      const auth = useAuthStore()
      auth.logout()
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

export default client
