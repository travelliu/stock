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

const spreadTips: Record<string, string> = {
  spreadOH: '高开价差\n= 最高价 − 开盘价\n用于预测当日最高价\n\n表中均值 = (平均值 + 中位数) / 2',
  spreadOL: '开低价差\n= 开盘价 − 最低价\n用于预测当日最低价\n\n表中均值 = (平均值 + 中位数) / 2',
  spreadHL: '高低价差\n= 最高价 − 最低价\n用于由最低价反推最高价（或反之）\n\n表中均值 = (平均值 + 中位数) / 2',
  spreadHC: '高收价差\n= 最高价 − 收盘价\n用于由最高价反推收盘价\n\n表中均值 = (平均值 + 中位数) / 2',
  spreadLC: '低收价差\n= 收盘价 − 最低价\n用于由最低价反推收盘价\n\n表中均值 = (平均值 + 中位数) / 2',
  spreadOC: '开收价差\n= |收盘价 − 开盘价|\n反映当日波动幅度\n\n表中均值 = (平均值 + 中位数) / 2',
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
      <el-table-column v-for="key in spreadKeys" :key="key" align="right">
        <template #header>
          <el-tooltip placement="top" effect="dark">
            <template #content>
              <span style="white-space: pre-line">{{ spreadTips[key] }}</span>
            </template>
            <span class="tip-header">{{ spreadLabels[key] }}</span>
          </el-tooltip>
        </template>
        <template #default="{ row }">
          {{ row.info.id === 'composite' ? compositeVal(key) : getMean(row.means, key) }}
        </template>
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
