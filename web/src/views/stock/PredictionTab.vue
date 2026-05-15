<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { wMessage } from '@/utils/message'
import { queryBars } from '@/apis/stocks'
import { getAnalysis, getPredictions, recalcPredictions } from '@/apis/analysis'
import type { AnalysisResult, AnalysisPrediction, DailyBar } from '@/types/api'
import SpreadModelTable from '@/components/SpreadModelTable.vue'
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
      <div class="price-grid">
        <div class="price-row">
          <span class="price-label">今日开盘</span>
          <el-input-number
            v-model="openPrice"
            :precision="2"
            :step="0.01"
            size="small"
            style="width: 140px"
          />
          <el-button type="primary" size="small" :loading="recalcing" @click="handleRecalc">
            计算预测
          </el-button>
        </div>
        <div class="price-row">
          <span class="price-label">实际 H / L / C</span>
          <span>
            {{ fmt(latestBar?.high ?? 0) }} /
            {{ fmt(latestBar?.low ?? 0) }} /
            {{ fmt(latestBar?.close ?? 0) }}
          </span>
        </div>
        <div class="price-row">
          <span class="price-label">预测 H / L / C</span>
          <span>
            {{ fmt(latestPred?.predictHigh ?? 0) }} /
            {{ fmt(latestPred?.predictLow ?? 0) }} /
            {{ fmt(latestPred?.predictClose ?? 0) }}
          </span>
        </div>
      </div>
    </el-card>

    <SpreadModelTable :result="analysis" />

    <div style="margin-top: 16px">
      <SpreadHistogram :bars="bars" />
    </div>
  </div>
</template>

<style scoped lang="scss">
.price-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.price-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.price-label {
  width: 120px;
  font-size: 13px;
  color: #606266;
  flex-shrink: 0;
}
</style>
