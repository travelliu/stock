<script setup lang="ts">
import { ref, inject, onMounted, computed } from 'vue'
import type { Ref } from 'vue'
import { useRoute } from 'vue-router'
import { wMessage } from '@/utils/message'
import { getAnalysis, getPredictions } from '@/apis/analysis'
import type { AnalysisParams } from '@/apis/analysis'
import type { AnalysisResult, AnalysisPrediction, RealtimeQuote } from '@/types/api'
import SpreadModelTable from '@/components/SpreadModelTable.vue'
import TradePlanTable from '@/components/TradePlanTable.vue'
import SpreadAnalysisTable from '@/components/SpreadAnalysisTable.vue'
import SpreadHistogram from '@/components/SpreadHistogram.vue'

const route = useRoute()
const code = route.params.code as string
const quote = inject<Ref<RealtimeQuote | undefined>>('stockQuote')

const analysis = ref<AnalysisResult | null>(null)
const latestPred = ref<AnalysisPrediction | null>(null)
const openPrice = ref(0)
const actualHigh = ref(0)
const actualLow = ref(0)
const actualClose = ref(0)
const loading = ref(false)
const recalcing = ref(false)

async function load() {
  loading.value = true
  try {
    const predsPage = await getPredictions(code, { limit: 1 })
    latestPred.value = predsPage.items[0] ?? null

    const initParams: AnalysisParams = {}
    const q = quote?.value
    if (q) {
      openPrice.value = q.open || q.price || 0
      actualHigh.value = q.high || 0
      actualLow.value = q.low || 0
      actualClose.value = q.price || 0
      if (openPrice.value) initParams.actualOpen = openPrice.value
      if (q.high) initParams.actualHigh = q.high
      if (q.low) initParams.actualLow = q.low
      if (q.price) initParams.actualClose = q.price
    }
    analysis.value = await getAnalysis(code, initParams)
  } catch (e: unknown) {
    wMessage('error', e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

async function handleRecalc() {
  recalcing.value = true
  try {
    const params = {
      actualOpen: openPrice.value || undefined,
      actualHigh: actualHigh.value || undefined,
      actualLow: actualLow.value || undefined,
      actualClose: actualClose.value || undefined,
    }
    analysis.value = await getAnalysis(code, params)
  } catch (e: unknown) {
    wMessage('error', e instanceof Error ? e.message : '计算失败')
  } finally {
    recalcing.value = false
  }
}

const displayPredHigh = computed(() =>
  analysis.value?.refTable?.high.mean || latestPred.value?.predictHigh || 0
)
const displayPredLow = computed(() =>
  analysis.value?.refTable?.low.mean || latestPred.value?.predictLow || 0
)
const displayPredClose = computed(() =>
  analysis.value?.refTable?.close.mean || latestPred.value?.predictClose || 0
)

function fmt(v: number): string {
  return v ? v.toFixed(2) : '-'
}

onMounted(load)
</script>

<template>
  <div v-loading="loading">
    <el-card style="margin-bottom: 10px">
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
            <td><el-input-number v-model="openPrice" :precision="2" :step="0.01" size="small" class="price-input" /></td>
            <td><el-input-number v-model="actualHigh" :precision="2" :step="0.01" size="small" class="price-input" /></td>
            <td><el-input-number v-model="actualLow" :precision="2" :step="0.01" size="small" class="price-input" /></td>
            <td><el-input-number v-model="actualClose" :precision="2" :step="0.01" size="small" class="price-input" /></td>
            <td>
              <el-button type="primary" size="small" :loading="recalcing" @click="handleRecalc">计算</el-button>
            </td>
          </tr>
          <tr>
            <td class="row-label">预测</td>
            <td>{{ fmt(openPrice) }}</td>
            <td>{{ fmt(displayPredHigh) }}</td>
            <td>{{ fmt(displayPredLow) }}</td>
            <td>{{ fmt(displayPredClose) }}</td>
            <td></td>
          </tr>
        </tbody>
      </table>
    </el-card>

    <TradePlanTable :result="analysis" />

    <SpreadModelTable :result="analysis" />

    <SpreadAnalysisTable :result="analysis" />

    <div style="margin-top: 10px">
      <SpreadHistogram :result="analysis" />
    </div>
  </div>
</template>

<style scoped lang="scss">
/* el-card header/body padding is set globally in element-reset.scss */
.price-table {
  border-collapse: collapse;
  font-size: 13px;
  th, td {
    padding: 6px 12px;
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
:deep(.price-input) {
  width: 110px;
  .el-input__inner {
    text-align: right;
  }
}
</style>
