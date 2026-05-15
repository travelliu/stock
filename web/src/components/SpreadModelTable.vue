<script setup lang="ts">
import type { AnalysisResult, MeansData } from '@/types/api'

const props = defineProps<{ result?: AnalysisResult | null }>()

const spreadKeys = ['spreadOH', 'spreadOL', 'spreadHL', 'spreadHC', 'spreadLC', 'spreadOC'] as const

const spreadLabels: Record<string, string> = {
  spreadOH: '开盘与最高价',
  spreadOL: '开盘与最低价',
  spreadHL: '最高与最低价',
  spreadHC: '最高与收盘价',
  spreadLC: '最低与收盘价',
  spreadOC: '开盘与收盘价',
}

function windowName(id: string): string {
  const map: Record<string, string> = { All: '历史', last_90: '近3月', last_30: '近1月', last_15: '近2周' }
  return map[id] || id
}

function getMean(means: MeansData | null, key: string): string {
  if (!means) return '-'
  const m = (means as unknown as Record<string, { mean: number } | null>)[key]
  return m && m.mean !== 0 ? m.mean.toFixed(2) : '-'
}

const camelToSnake: Record<string, string> = {
  spreadOH: 'spread_oh',
  spreadOL: 'spread_ol',
  spreadHL: 'spread_hl',
  spreadHC: 'spread_hc',
  spreadLC: 'spread_lc',
  spreadOC: 'spread_oc',
}

function compositeVal(key: string): string {
  const cm = props.result?.compositeMeans
  if (!cm) return '-'
  const v = cm[camelToSnake[key] ?? key]
  return v ? v.toFixed(2) : '-'
}
</script>

<template>
  <el-card v-if="result">
    <template #header>{{ $t('stockDetail.modelTable') }}</template>
    <el-table :data="[...result.windows, { info: { id: 'composite' }, means: null }]" size="small" border>
      <el-table-column :label="$t('stockDetail.timePeriod')" width="100">
        <template #default="{ row }">
          {{ row.info.id === 'composite' ? '综合均值' : windowName(row.info.id) }}
        </template>
      </el-table-column>
      <el-table-column v-for="key in spreadKeys" :key="key" :label="spreadLabels[key]" align="right">
        <template #default="{ row }">
          {{ row.info.id === 'composite' ? compositeVal(key) : getMean(row.means, key) }}
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>
