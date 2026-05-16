<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { wMessage } from '@/utils/message'
import { usePortfolioStore } from '@/stores/portfolio'
import { searchStocks, queryBars, getQuote } from '@/apis/stocks'
import type { Stock, Portfolio, DailyBar, RealtimeQuote } from '@/types/api'

const router = useRouter()
const portfolioStore = usePortfolioStore()
const showAdd = ref(false)
const selectedCode = ref('')
const note = ref('')
const stockOptions = ref<{ value: string; label: string }[]>([])
const loadingAdd = ref(false)

interface BarInfo {
  bar: DailyBar
  pctChg: number
}
const barMap = ref<Record<string, BarInfo>>({})
const quoteMap = ref<Record<string, RealtimeQuote>>({})
const loadingBars = ref(false)

async function loadBars(items: Portfolio[]) {
  if (!items.length) return
  loadingBars.value = true
  try {
    const results = await Promise.all(items.map(p => queryBars(p.code, { limit: 2 })))
    const map: Record<string, BarInfo> = {}
    results.forEach((res, i) => {
      const bars = res.items
      if (!bars.length) return
      const bar = bars[0]
      const prevClose = bars[1]?.close
      const pctChg = prevClose ? (bar.close - prevClose) / prevClose * 100 : 0
      map[items[i].code] = { bar, pctChg }
    })
    barMap.value = map
  } catch {
    // silently ignore
  } finally {
    loadingBars.value = false
  }
}

async function loadQuotes(items: Portfolio[]) {
  if (!items.length) return
  const results = await Promise.allSettled(items.map(p => getQuote(p.tsCode)))
  const map: Record<string, RealtimeQuote> = { ...quoteMap.value }
  results.forEach((r, i) => {
    if (r.status === 'fulfilled') map[items[i].code] = r.value
  })
  quoteMap.value = map
}

onMounted(async () => {
  await portfolioStore.fetch()
  loadBars(portfolioStore.items)
  loadQuotes(portfolioStore.items)
})

async function searchStockOptions(query: string) {
  if (!query) { stockOptions.value = []; return }
  try {
    const list = await searchStocks(query, 20)
    stockOptions.value = list.map((s: Stock) => ({ value: s.code, label: `${s.code} ${s.name}` }))
  } catch {
    stockOptions.value = []
  }
}

async function doAdd() {
  if (!selectedCode.value) { wMessage('warning', '请选择股票'); return }
  loadingAdd.value = true
  try {
    await portfolioStore.add(selectedCode.value, note.value)
    wMessage('success', '添加成功')
    showAdd.value = false
    selectedCode.value = ''
    note.value = ''
    loadBars(portfolioStore.items)
    loadQuotes(portfolioStore.items)
  } finally {
    loadingAdd.value = false
  }
}

function goDetail(row: Portfolio) {
  router.push(`/stocks/${row.code}`)
}

function removeItem(row: Portfolio) {
  portfolioStore.remove(row.code)
}

function fmtDate(d: string): string {
  if (!d || d.length !== 8) return d
  return `${d.slice(4, 6)}/${d.slice(6, 8)}`
}

function fmtPrice(v: number): string {
  return v ? v.toFixed(2) : '-'
}

function pctClass(v: number): string {
  if (v > 0) return 'up'
  if (v < 0) return 'down'
  return ''
}

function fmtPct(v: number): string {
  if (!v) return '-'
  return (v > 0 ? '+' : '') + v.toFixed(2) + '%'
}
</script>

<template>
  <div>
    <div class="header-bar">
      <h2>{{ $t('stockList.title') }}</h2>
      <el-button type="primary" @click="showAdd = true">{{ $t('stockList.addStock') }}</el-button>
    </div>

    <el-table :data="portfolioStore.items" v-loading="loadingBars" style="margin-top: 16px">
      <el-table-column prop="code" :label="$t('stockList.code')" width="100">
        <template #default="{ row }">
          <el-button link type="primary" @click="goDetail(row)">{{ row.code }}</el-button>
        </template>
      </el-table-column>
      <el-table-column prop="name" :label="$t('stockList.name')" width="120" />
      <el-table-column label="日期" width="60" align="center">
        <template #default="{ row }">
          <span class="secondary">{{ quoteMap[row.code]?.quoteTime?.slice(0, 5) || fmtDate(barMap[row.code]?.bar.tradeDate ?? '') }}</span>
        </template>
      </el-table-column>
      <el-table-column label="开盘" align="right" width="80">
        <template #default="{ row }">{{ fmtPrice(quoteMap[row.code]?.open ?? barMap[row.code]?.bar.open ?? 0) }}</template>
      </el-table-column>
      <el-table-column label="最高" align="right" width="80">
        <template #default="{ row }">{{ fmtPrice(quoteMap[row.code]?.high ?? barMap[row.code]?.bar.high ?? 0) }}</template>
      </el-table-column>
      <el-table-column label="最低" align="right" width="80">
        <template #default="{ row }">{{ fmtPrice(quoteMap[row.code]?.low ?? barMap[row.code]?.bar.low ?? 0) }}</template>
      </el-table-column>
      <el-table-column label="最新" align="right" width="80">
        <template #default="{ row }">
          <span :class="pctClass(quoteMap[row.code]?.changePct ?? barMap[row.code]?.pctChg ?? 0)">
            {{ fmtPrice(quoteMap[row.code]?.price ?? barMap[row.code]?.bar.close ?? 0) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="涨跌幅" align="right" width="90">
        <template #default="{ row }">
          <span :class="pctClass(quoteMap[row.code]?.changePct ?? barMap[row.code]?.pctChg ?? 0)">
            {{ quoteMap[row.code] ? fmtPct(quoteMap[row.code].changePct) : fmtPct(barMap[row.code]?.pctChg ?? 0) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column prop="note" :label="$t('stockList.note')" />
      <el-table-column :label="$t('stockList.action')" width="120">
        <template #default="{ row }">
          <el-button link type="primary" @click="goDetail(row)">{{ $t('stockList.detail') }}</el-button>
          <el-button link type="danger" @click="removeItem(row)">{{ $t('common.delete') }}</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showAdd" :title="$t('stockList.addStock')" width="400px">
      <el-form @submit.prevent="doAdd">
        <el-form-item :label="$t('stockList.code')">
          <el-select-v2
            v-model="selectedCode"
            :options="stockOptions"
            :placeholder="$t('stockList.selectStock')"
            clearable filterable remote
            :remote-method="searchStockOptions"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="$t('stockList.note')">
          <el-input v-model="note" :placeholder="$t('common.empty')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAdd = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="loadingAdd" @click="doAdd">{{ $t('common.add') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.up   { color: var(--el-color-danger); }
.down { color: var(--el-color-success); }
.secondary { color: var(--el-text-color-secondary); font-size: 12px; }
</style>
