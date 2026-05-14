import { defineStore } from 'pinia'
import { ref } from 'vue'
import client from '@/api/client'

interface PortfolioEntry {
  id: number
  ts_code: string
  note: string
  added_at: string
}

export const usePortfolioStore = defineStore('portfolio', () => {
  const items = ref<PortfolioEntry[]>([])

  async function fetch() {
    const { data } = await client.get('/portfolio')
    items.value = data
  }

  async function add(tsCode: string, note: string) {
    await client.post('/portfolio', { ts_code: tsCode, note })
    await fetch()
  }

  async function remove(tsCode: string) {
    await client.delete(`/portfolio/${tsCode}`)
    await fetch()
  }

  return { items, fetch, add, remove }
})
