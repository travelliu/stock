import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User } from '@/types/api'
import * as authApi from '@/apis/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const initialized = ref(false)

  async function fetchMe() {
    try {
      user.value = await authApi.me()
    } catch {
      user.value = null
    } finally {
      initialized.value = true
    }
  }

  async function login(username: string, password: string) {
    user.value = await authApi.login({ username, password })
    return user.value
  }

  async function logout() {
    try {
      await authApi.logout()
    } catch {
      // ignore
    }
    user.value = null
  }

  return { user, initialized, fetchMe, login, logout }
})
