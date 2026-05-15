import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Portfolio } from '@/types/api'
import { listPortfolio, addPortfolio, removePortfolio } from '@/apis/portfolio'

export const usePortfolioStore = defineStore('portfolio', () => {
  const items = ref<Portfolio[]>([])

  async function fetch() {
    items.value = await listPortfolio()
  }

  async function add(code: string, note: string) {
    await addPortfolio({ code, note })
    await fetch()
  }

  async function remove(code: string) {
    await removePortfolio(code)
    await fetch()
  }

  return { items, fetch, add, remove }
})
