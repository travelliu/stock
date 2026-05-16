<script setup lang="ts">
import { ref, computed, provide, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { wMessage } from '@/utils/message'
import { getStock, getQuote } from '@/apis/stocks'
import type { Stock, RealtimeQuote } from '@/types/api'
import StockBasicCard from '@/components/StockBasicCard.vue'

const route = useRoute()
const router = useRouter()
const code = route.params.code as string

const stock = ref<Stock | null>(null)
const quote = ref<RealtimeQuote | undefined>(undefined)
const loading = ref(false)

provide('stockQuote', quote)

const activeTab = computed(() => {
  if (route.path.endsWith('/bars')) return 'bars'
  if (route.path.endsWith('/predictions')) return 'predictions'
  return 'prediction'
})

function onTabClick(paneName: string) {
  if (paneName === 'bars') router.push(`/stocks/${code}/bars`)
  else if (paneName === 'predictions') router.push(`/stocks/${code}/predictions`)
  else router.push(`/stocks/${code}`)
}

onMounted(async () => {
  loading.value = true
  try {
    const [s, q] = await Promise.all([getStock(code), getQuote(code).catch(() => undefined)])
    stock.value = s
    quote.value = q
  } catch (e: unknown) {
    wMessage('error', e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div v-loading="loading">
    <div v-if="stock">
      <StockBasicCard :stock="stock" :quote="quote" @back="router.push('/stocks')" />
    </div>
    <el-tabs
      :model-value="activeTab"
      style="margin-top: 8px"
      @tab-click="(tab: any) => onTabClick(tab.paneName)"
    >
      <el-tab-pane label="预测" name="prediction" />
      <el-tab-pane label="日K数据" name="bars" />
      <el-tab-pane label="预测记录" name="predictions" />
    </el-tabs>
    <router-view />
  </div>
</template>

<style scoped lang="scss">
.detail-header {
  display: flex;
  align-items: center;
}
</style>
