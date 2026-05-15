<script setup lang="ts">
import type { AnalysisResult, MeansAvgData, RecommendRangeResult } from '@/types/api'

const props = defineProps<{ result?: AnalysisResult | null }>()

function windowName(id: string): string {
  const map: Record<string, string> = { All: '历史', last_90: '近3月', last_30: '近1月', last_15: '近2周' }
  return map[id] || id
}

function fmt(v: number | null | undefined): string {
  return v != null && v !== 0 ? v.toFixed(2) : '-'
}

function fmtRec(r: RecommendRangeResult | null | undefined): string {
  if (!r) return '-'
  return `${r.low.toFixed(2)}~${r.high.toFixed(2)} (${r.cumPct.toFixed(1)}%)`
}

function stats(m: MeansAvgData | null | undefined) {
  return {
    count: m?.count ?? 0,
    avg: fmt(m?.avg),
    median: fmt(m?.median),
    mean: fmt(m?.mean),
    rec: fmtRec(m?.recommend),
  }
}

interface Row {
  label: string
  oh: ReturnType<typeof stats>
  ol: ReturnType<typeof stats>
}

function buildRows(): Row[] {
  if (!props.result?.windows) return []
  return [...props.result.windows].reverse().map(w => ({
    label: windowName(w.info.id),
    oh: stats(w.means?.spreadOH),
    ol: stats(w.means?.spreadOL),
  }))
}
</script>

<template>
  <el-card v-if="result" style="margin-top: 16px">
    <template #header>价差分析</template>
    <el-table :data="buildRows()" size="small" border>
      <el-table-column label="时段" prop="label" width="70" />
      <el-table-column label="── 最高-开盘 ──">
        <el-table-column label="样本" align="right" width="55">
          <template #default="{ row }">{{ row.oh.count }}</template>
        </el-table-column>
        <el-table-column label="均值" align="right" width="70">
          <template #default="{ row }">{{ row.oh.mean }}</template>
        </el-table-column>
        <el-table-column label="中位数" align="right" width="70">
          <template #default="{ row }">{{ row.oh.median }}</template>
        </el-table-column>
        <el-table-column label="平均值" align="right" width="70">
          <template #default="{ row }">{{ row.oh.avg }}</template>
        </el-table-column>
      </el-table-column>
      <el-table-column label="── 开盘-最低 ──">
        <el-table-column label="样本" align="right" width="55">
          <template #default="{ row }">{{ row.ol.count }}</template>
        </el-table-column>
        <el-table-column label="均值" align="right" width="70">
          <template #default="{ row }">{{ row.ol.mean }}</template>
        </el-table-column>
        <el-table-column label="中位数" align="right" width="70">
          <template #default="{ row }">{{ row.ol.median }}</template>
        </el-table-column>
        <el-table-column label="平均值" align="right" width="70">
          <template #default="{ row }">{{ row.ol.avg }}</template>
        </el-table-column>
      </el-table-column>
      <el-table-column label="高抛推荐 (累计占比)" align="right">
        <template #default="{ row }">{{ row.oh.rec }}</template>
      </el-table-column>
      <el-table-column label="低吸推荐 (累计占比)" align="right">
        <template #default="{ row }">{{ row.ol.rec }}</template>
      </el-table-column>
    </el-table>
  </el-card>
</template>
