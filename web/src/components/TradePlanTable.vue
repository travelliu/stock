<script setup lang="ts">
import type { AnalysisResult, MeansData } from '@/types/api'

const props = defineProps<{ result?: AnalysisResult | null }>()

function windowName(id: string): string {
  const map: Record<string, string> = { All: '历史', last_90: '近3月', last_30: '近1月', last_15: '近2周' }
  return map[id] || id
}

function getSpreadMean(means: MeansData | null, key: string): number {
  if (!means) return 0
  const m = (means as unknown as Record<string, { mean: number } | null>)[key]
  return m ? m.mean : 0
}

function fmt(v: number): string {
  return v ? v.toFixed(2) : '/'
}

interface PredictRow {
  label: string
  values: string[]
  reverse: string
  mean: string
  direction: string
}

function buildRows(): PredictRow[] {
  const r = props.result
  if (!r || !r.openPrice) return []

  const winMap = new Map<string, MeansData | null>()
  for (const w of r.windows) winMap.set(w.info.id, w.means)

  const lastWin = r.windows.length > 0 ? r.windows[r.windows.length - 1] : null
  const lastMeans = lastWin?.means ?? null

  const rows: PredictRow[] = []

  // High prediction
  const highVals: number[] = []
  const highCells: string[] = []
  for (const w of r.windows) {
    const m = getSpreadMean(w.means, 'spreadOH')
    if (m !== 0) {
      const v = r.openPrice! + m
      highCells.push(fmt(v))
      highVals.push(v)
    } else {
      highCells.push('/')
    }
  }
  let highReverse = '/'
  if (r.actualLow && lastMeans) {
    const hl = getSpreadMean(lastMeans, 'spreadHL')
    if (hl !== 0) {
      const v = r.actualLow + hl
      highReverse = fmt(v)
      highVals.push(v)
    }
  }
  const highMean = highVals.length > 0 ? highVals.reduce((a, b) => a + b, 0) / highVals.length : 0
  rows.push({
    label: '最高价预测',
    values: highCells,
    reverse: highReverse,
    mean: highMean ? fmt(highMean) : '/',
    direction: '+',
  })

  // Low prediction
  const lowVals: number[] = []
  const lowCells: string[] = []
  for (const w of r.windows) {
    const m = getSpreadMean(w.means, 'spreadOL')
    if (m !== 0) {
      const v = r.openPrice! - m
      lowCells.push(fmt(v))
      lowVals.push(v)
    } else {
      lowCells.push('/')
    }
  }
  let lowReverse = '/'
  if (r.actualHigh && lastMeans) {
    const hl = getSpreadMean(lastMeans, 'spreadHL')
    if (hl !== 0) {
      const v = r.actualHigh - hl
      lowReverse = fmt(v)
      lowVals.push(v)
    }
  }
  const lowMean = lowVals.length > 0 ? lowVals.reduce((a, b) => a + b, 0) / lowVals.length : 0
  rows.push({
    label: '最低价预测',
    values: lowCells,
    reverse: lowReverse,
    mean: lowMean ? fmt(lowMean) : '/',
    direction: '-',
  })

  // Close prediction
  const closeVals: number[] = []
  let closeReverseLow = '/'
  let closeReverseHigh = '/'
  if (r.actualLow && lastMeans) {
    const lc = getSpreadMean(lastMeans, 'spreadLC')
    if (lc !== 0) {
      const v = r.actualLow + lc
      closeReverseLow = fmt(v)
      closeVals.push(v)
    }
  }
  if (r.actualHigh && lastMeans) {
    const hc = getSpreadMean(lastMeans, 'spreadHC')
    if (hc !== 0) {
      const v = r.actualHigh - hc
      closeReverseHigh = fmt(v)
      closeVals.push(v)
    }
  }
  const closeMean = closeVals.length > 0 ? closeVals.reduce((a, b) => a + b, 0) / closeVals.length : 0
  rows.push({
    label: '收盘价预测',
    values: r.windows.map(() => '/'),
    reverse: closeReverseLow,
    mean: closeMean ? fmt(closeMean) : '/',
    direction: '-',
    _reverseHigh: closeReverseHigh,
  } as PredictRow & { _reverseHigh: string })

  return rows
}
</script>

<template>
  <el-card v-if="result?.openPrice" style="margin-top: 16px">
    <template #header>{{ $t('stockDetail.tradePlan') }}</template>
    <el-table :data="buildRows()" size="small" border>
      <el-table-column :label="$t('stockDetail.timePeriod')" width="100" prop="label" />
      <el-table-column v-for="(w, i) in result?.windows ?? []" :key="i" :label="windowName(w.info.id)" align="right">
        <template #default="{ row }">
          {{ row.values[i] }}
        </template>
      </el-table-column>
      <el-table-column label="最低价反推" align="right" prop="reverse" />
      <el-table-column label="最高价反推" align="right">
        <template #default="{ row }">
          {{ (row as any)._reverseHigh ?? '/' }}
        </template>
      </el-table-column>
      <el-table-column label="均值" align="right" prop="mean" />
      <el-table-column label="正负算一" width="70" align="center" prop="direction" />
    </el-table>
  </el-card>
</template>
