<script setup lang="ts">
import type { AnalysisResult } from '@/types/api'

const props = defineProps<{ result?: AnalysisResult | null }>()

function windowName(id: string): string {
  const map: Record<string, string> = { All: '历史', last_90: '近3月', last_30: '近1月', last_15: '近2周' }
  return map[id] || id
}


function fmt(v: number): string {
  return v ? v.toFixed(2) : '/'
}

function buildRows() {
  const t = props.result?.refTable
  if (!t) return []
  return [
    { label: '最高价预测', row: t.high },
    { label: '最低价预测', row: t.low },
    { label: '收盘价预测', row: t.close },
  ]
}
</script>

<template>
  <el-card v-if="result?.refTable" style="margin-top: 16px">
    <template #header>预测收盘价 (历史参考价)</template>
    <el-table :data="buildRows()" size="small" border>
      <el-table-column label="" width="100" prop="label" />
      <el-table-column
        v-for="w in result?.windows ?? []"
        :key="w.info.id"
        align="right"
      >
        <template #header>
          <el-tooltip placement="top" effect="dark">
            <template #content>
              {{ windowName(w.info.id) }}数据均值预测<br />
              最高价 = 开盘价 + {{ windowName(w.info.id) }}高开价差均值<br />
              最低价 = 开盘价 − {{ windowName(w.info.id) }}开低价差均值
            </template>
            <span class="tip-header">{{ windowName(w.info.id) }}</span>
          </el-tooltip>
        </template>
        <template #default="{ row }">{{ fmt(row.row.windows[w.info.id] ?? 0) }}</template>
      </el-table-column>
      <el-table-column align="right">
        <template #header>
          <el-tooltip placement="top" effect="dark">
            <template #content>
              由当日最低价反推<br />
              • 最高价预测：最低价 + 近2周高低价差均值<br />
              • 收盘价预测：最低价 + 近2周低收价差均值
            </template>
            <span class="tip-header">最低价反推</span>
          </el-tooltip>
        </template>
        <template #default="{ row }">{{ fmt(row.row.reverseLow) }}</template>
      </el-table-column>
      <el-table-column align="right">
        <template #header>
          <el-tooltip placement="top" effect="dark">
            <template #content>
              由当日最高价反推<br />
              • 最低价预测：最高价 − 近2周高低价差均值<br />
              • 收盘价预测：最高价 − 近2周高收价差均值
            </template>
            <span class="tip-header">最高价反推</span>
          </el-tooltip>
        </template>
        <template #default="{ row }">{{ fmt(row.row.reverseHigh) }}</template>
      </el-table-column>
      <el-table-column align="right">
        <template #header>
          <el-tooltip content="本行所有有效数值的算术平均" placement="top" effect="dark">
            <span class="tip-header">均值</span>
          </el-tooltip>
        </template>
        <template #default="{ row }">{{ fmt(row.row.mean) }}</template>
      </el-table-column>
      <el-table-column label="正负算一" width="70" align="center">
        <template #default="{ row }">{{ row.row.direction }}</template>
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
