<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { wMessage } from '@/utils/message'
import { getStock, queryBars } from '@/apis/stocks'
import type { Stock, DailyBar } from '@/types/api'
import StockBasicCard from '@/components/StockBasicCard.vue'

const route = useRoute()
const router = useRouter()
const code = route.params.code as string

const stock = ref<Stock | null>(null)
const lastBar = ref<DailyBar | undefined>(undefined)
const loading = ref(false)

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
    const [s, barsPage] = await Promise.all([
      getStock(code),
      queryBars(code, { limit: 1 }),
    ])
    stock.value = s
    lastBar.value = barsPage.items[0]
  } catch (e: unknown) {
    wMessage('error', e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div v-loading="loading">
    <div class="detail-header">
      <el-button link @click="router.push('/stocks')">← 返回列表</el-button>
    </div>
    <div v-if="stock" style="margin-top: 8px">
      <StockBasicCard :stock="stock" :last-bar="lastBar" />
    </div>
    <el-tabs
      :model-value="activeTab"
      style="margin-top: 16px"
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
