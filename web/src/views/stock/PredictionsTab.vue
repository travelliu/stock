<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { wMessage } from '@/utils/message'
import { getPredictions, recalcPredictions } from '@/apis/analysis'
import type { AnalysisPrediction } from '@/types/api'

const props = defineProps<{ tsCode: string }>()

const preds = ref<AnalysisPrediction[]>([])
const loading = ref(false)
const recalcing = ref(false)

async function load() {
  loading.value = true
  try {
    preds.value = await getPredictions(props.tsCode, { limit: 30 })
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '加载失败'
    wMessage('error', msg)
  } finally {
    loading.value = false
  }
}

async function handleRecalc() {
  recalcing.value = true
  try {
    const res = await recalcPredictions(props.tsCode)
    wMessage('success', `更新了 ${res.updated} 条记录`)
    await load()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '重算失败'
    wMessage('error', msg)
  } finally {
    recalcing.value = false
  }
}

function fmt(v: number): string {
  return v ? v.toFixed(2) : '-'
}

function devFmt(actual: number, predict: number): string {
  if (!actual || !predict) return '-'
  const d = actual - predict
  const prefix = d >= 0 ? '+' : ''
  return prefix + d.toFixed(2)
}

onMounted(load)
</script>

<template>
  <div v-loading="loading">
    <div style="margin-bottom: 12px; display: flex; justify-content: flex-end">
      <el-button type="primary" size="small" :loading="recalcing" @click="handleRecalc">
        {{ $t('stockDetail.recalc') }}
      </el-button>
    </div>
    <el-table :data="preds" size="small" border>
      <el-table-column :label="$t('stockDetail.date')" prop="tradeDate" width="110" />
      <el-table-column :label="$t('stockDetail.predictHigh')" align="right">
        <template #default="{ row }">{{ fmt(row.predictHigh) }}</template>
      </el-table-column>
      <el-table-column :label="$t('stockDetail.actualHigh')" align="right">
        <template #default="{ row }">{{ fmt(row.actualHigh) }}</template>
      </el-table-column>
      <el-table-column :label="$t('stockDetail.devHigh')" align="right">
        <template #default="{ row }">{{ devFmt(row.actualHigh, row.predictHigh) }}</template>
      </el-table-column>
      <el-table-column :label="$t('stockDetail.predictLow')" align="right">
        <template #default="{ row }">{{ fmt(row.predictLow) }}</template>
      </el-table-column>
      <el-table-column :label="$t('stockDetail.actualLow')" align="right">
        <template #default="{ row }">{{ fmt(row.actualLow) }}</template>
      </el-table-column>
      <el-table-column :label="$t('stockDetail.devLow')" align="right">
        <template #default="{ row }">{{ devFmt(row.actualLow, row.predictLow) }}</template>
      </el-table-column>
    </el-table>
  </div>
</template>
