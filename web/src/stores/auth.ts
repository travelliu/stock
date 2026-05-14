import { defineStore } from 'pinia'
import { ref } from 'vue'
import client from '@/api/client'

interface User {
  id: number
  username: string
  role: string
  tushare_token: string
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)

  async function fetchMe() {
    try {
      const { data } = await client.get('/auth/me')
      user.value = data
    } catch {
      user.value = null
    }
  }

  async function login(username: string, password: string) {
    const { data } = await client.post('/auth/login', { username, password })
    user.value = data
    return data
  }

  async function logout() {
    try {
      await client.post('/auth/logout')
    } catch {
      // ignore
    }
    user.value = null
    window.location.href = '/login'
  }

  return { user, fetchMe, login, logout }
})
