<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { wMessage } from '@/utils/message'
import { queryBars } from '@/apis/stocks'
import { getAnalysis, getPredictions, recalcPredictions } from '@/apis/analysis'
import type { AnalysisResult, AnalysisPrediction, DailyBar } from '@/types/api'
import SpreadModelTable from '@/components/SpreadModelTable.vue'
import TradePlanTable from '@/components/TradePlanTable.vue'
import SpreadAnalysisTable from '@/components/SpreadAnalysisTable.vue'
import SpreadHistogram from '@/components/SpreadHistogram.vue'

const route = useRoute()
const code = route.params.code as string

const analysis = ref<AnalysisResult | null>(null)
const latestPred = ref<AnalysisPrediction | null>(null)
const bars = ref<DailyBar[]>([])
const latestBar = ref<DailyBar | null>(null)
const openPrice = ref(0)
const loading = ref(false)
const recalcing = ref(false)

async function load() {
  loading.value = true
  try {
    const [analysisRes, predsPage, barsPage] = await Promise.all([
      getAnalysis(code),
      getPredictions(code, { limit: 1 }),
      queryBars(code, { limit: 30 }),
    ])
    analysis.value = analysisRes
    latestPred.value = predsPage.items[0] ?? null
    latestBar.value = barsPage.items[0] ?? null
    bars.value = barsPage.items
    if (latestBar.value?.open) {
      openPrice.value = latestBar.value.open
    }
  } catch (e: unknown) {
    wMessage('error', e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

async function handleRecalc() {
  recalcing.value = true
  try {
    const res = await recalcPredictions(code)
    wMessage('success', `更新了 ${res.updated} 条记录`)
    await load()
  } catch (e: unknown) {
    wMessage('error', e instanceof Error ? e.message : '重算失败')
  } finally {
    recalcing.value = false
  }
}

function fmt(v: number): string {
  return v ? v.toFixed(2) : '-'
}

onMounted(load)
</script>

<template>
  <div v-loading="loading">
    <el-card style="margin-bottom: 16px">
      <template #header>实时价格与预测</template>
      <table class="price-table">
        <thead>
          <tr>
            <th></th>
            <th>开盘价</th>
            <th>最高价</th>
            <th>最低价</th>
            <th>收盘价</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td class="row-label">实际</td>
            <td>
              <el-input-number
                v-model="openPrice"
                :precision="2"
                :step="0.01"
                size="small"
                style="width: 120px"
              />
            </td>
            <td>{{ fmt(latestBar?.high ?? 0) }}</td>
            <td>{{ fmt(latestBar?.low ?? 0) }}</td>
            <td>{{ fmt(latestBar?.close ?? 0) }}</td>
            <td>
              <el-button type="primary" size="small" :loading="recalcing" @click="handleRecalc">
                计算
              </el-button>
            </td>
          </tr>
          <tr>
            <td class="row-label">预测</td>
            <td>{{ fmt(openPrice) }}</td>
            <td>{{ fmt(latestPred?.predictHigh ?? 0) }}</td>
            <td>{{ fmt(latestPred?.predictLow ?? 0) }}</td>
            <td>{{ fmt(latestPred?.predictClose ?? 0) }}</td>
            <td></td>
          </tr>
        </tbody>
      </table>
    </el-card>

    <SpreadModelTable :result="analysis" />

    <TradePlanTable :result="analysis" />

    <SpreadAnalysisTable :result="analysis" />

    <div style="margin-top: 16px">
      <SpreadHistogram :bars="bars" />
    </div>
  </div>
</template>

<style scoped lang="scss">
.price-table {
  border-collapse: collapse;
  font-size: 13px;
  th, td {
    padding: 6px 16px;
    text-align: right;
    border-bottom: 1px solid var(--el-border-color-lighter);
    white-space: nowrap;
  }
  th {
    color: var(--el-text-color-secondary);
    font-weight: 500;
    border-bottom: 1px solid var(--el-border-color);
  }
  td:first-child, th:first-child {
    text-align: left;
  }
  .row-label {
    font-weight: 500;
    color: var(--el-text-color-regular);
    width: 48px;
  }
}
</style>
