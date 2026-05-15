<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { wMessage } from '@/utils/message'
import { getStock, queryBars } from '@/apis/stocks'
import type { Stock, DailyBar } from '@/types/api'
import StockBasicCard from '@/components/StockBasicCard.vue'
import SpreadHistogram from '@/components/SpreadHistogram.vue'
import DailyBarTable from '@/components/DailyBarTable.vue'

const props = defineProps<{ tsCode: string }>()

const stock = ref<Stock | null>(null)
const bars = ref<DailyBar[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const [s, b] = await Promise.all([
      getStock(props.tsCode),
      queryBars(props.tsCode),
    ])
    stock.value = s
    bars.value = b.slice(-30).reverse()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '加载失败'
    wMessage('error', msg)
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div v-loading="loading">
    <StockBasicCard v-if="stock" :stock="stock" :last-bar="bars[0]" />
    <div style="margin-top: 16px">
      <SpreadHistogram :bars="bars" />
    </div>
    <div style="margin-top: 16px">
      <DailyBarTable :bars="bars" />
    </div>
  </div>
</template>
