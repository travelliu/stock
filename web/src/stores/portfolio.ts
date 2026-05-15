import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Portfolio } from '@/types/api'
import { listPortfolio, addPortfolio, removePortfolio } from '@/apis/portfolio'

export const usePortfolioStore = defineStore('portfolio', () => {
  const items = ref<Portfolio[]>([])

  async function fetch() {
    items.value = await listPortfolio()
  }

  async function add(tsCode: string, note: string) {
    await addPortfolio({ tsCode, note })
    await fetch()
  }

  async function remove(tsCode: string) {
    await removePortfolio(tsCode)
    await fetch()
  }

  return { items, fetch, add, remove }
})
