<script setup lang="ts">
import type { AnalysisResult, PredictBreakdown } from '@/types/api'

const props = defineProps<{ result?: AnalysisResult | null }>()

function windowName(id: string): string {
  const map: Record<string, string> = { All: '历史', last_90: '近3月', last_30: '近1月', last_15: '近2周' }
  return map[id] || id
}

function fmt(v: number | null | undefined): string {
  return v ? v.toFixed(2) : '/'
}

interface MethodRow {
  label: string
  get: (b: PredictBreakdown) => number
}

const methods: MethodRow[] = [
  { label: '均值',     get: b => b.byMean },
  { label: '中位数',   get: b => b.byMedian },
  { label: 'EWMA',    get: b => b.byEwma },
  { label: '比率',     get: b => b.byRatio },
  { label: '最低反推', get: b => b.reverseLow },
  { label: '最高反推', get: b => b.reverseHigh },
  { label: '综合均值', get: b => b.mean },
]

interface DataRow {
  label: string
  high: (PredictBreakdown | null)[]   // one entry per window
  low: (PredictBreakdown | null)[]
  close: (PredictBreakdown | null)[]
}

// Keep a reference to the getter alongside each row for template use.
interface FullRow extends DataRow {
  _get: (b: PredictBreakdown) => number
}

function buildFullRows(): FullRow[] {
  const windows = props.result?.windows ?? []
  return methods.map(m => ({
    label: m.label,
    high:  windows.map(w => w.predict?.high  ?? null),
    low:   windows.map(w => w.predict?.low   ?? null),
    close: windows.map(w => w.predict?.close ?? null),
    _get:  m.get,
  }))
}
</script>

<template>
  <el-card v-if="result?.refTable" style="margin-top: 10px">
    <template #header>预测参考价</template>

    <el-table :data="buildFullRows()" size="small" border>
      <el-table-column label="方法" prop="label" width="80" fixed />

      <!-- 最高价预测 group -->
      <el-table-column label="最高价预测" header-align="center" align="center">
        <el-table-column
          v-for="(w, i) in (result?.windows ?? [])"
          :key="'h-' + w.info.id"
          :label="windowName(w.info.id)"
          align="right"
          width="72"
        >
          <template #default="{ row }">{{ fmt(row._get(row.high[i])) }}</template>
        </el-table-column>
      </el-table-column>

      <!-- 最低价预测 group -->
      <el-table-column label="最低价预测" header-align="center" align="center">
        <el-table-column
          v-for="(w, i) in (result?.windows ?? [])"
          :key="'l-' + w.info.id"
          :label="windowName(w.info.id)"
          align="right"
          width="72"
        >
          <template #default="{ row }">{{ fmt(row._get(row.low[i])) }}</template>
        </el-table-column>
      </el-table-column>

      <!-- 收盘价预测 group -->
      <el-table-column label="收盘价预测" header-align="center" align="center">
        <el-table-column
          v-for="(w, i) in (result?.windows ?? [])"
          :key="'c-' + w.info.id"
          :label="windowName(w.info.id)"
          align="right"
          width="72"
        >
          <template #default="{ row }">{{ fmt(row._get(row.close[i])) }}</template>
        </el-table-column>
      </el-table-column>
    </el-table>

    <!-- 跨窗口综合均值 -->
    <el-table
      :data="[
        { label: '跨窗口综合均值', high: result.refTable.high.mean, low: result.refTable.low.mean, close: result.refTable.close.mean }
      ]"
      size="small"
      border
      style="margin-top: 6px"
    >
      <el-table-column label="" prop="label" width="80" />
      <el-table-column label="最高价" align="right" width="72">
        <template #default="{ row }">{{ fmt(row.high) }}</template>
      </el-table-column>
      <el-table-column label="最低价" align="right" width="72">
        <template #default="{ row }">{{ fmt(row.low) }}</template>
      </el-table-column>
      <el-table-column label="收盘价" align="right" width="72">
        <template #default="{ row }">{{ fmt(row.close) }}</template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<style scoped>
.tip-header {
  cursor: help;
  border-bottom: 1px dashed var(--el-text-color-secondary);
}
</style>
