<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Stock, DailyBar, RealtimeQuote } from '@/types/api'
import { fmtPrice, fmtPct, priceClass, fmtQuoteTime } from '@/utils/format'

const props = defineProps<{ stock: Stock; lastBar?: DailyBar; prevClose?: number; quote?: RealtimeQuote }>()
const emit = defineEmits<{ back: [] }>()

const expanded = ref(false)

const displayPrice = computed(() => {
  if (props.quote) return props.quote.price
  return props.lastBar?.close
})

const changeClass = computed(() => {
  if (props.quote) return priceClass(props.quote.price, props.quote.prevClose)
  if (!props.lastBar) return 'g-flat'
  return priceClass(props.lastBar.close, props.prevClose || props.lastBar.open)
})

const changePct = computed(() => {
  if (props.quote) return `${props.quote.changePct >= 0 ? '+' : ''}${props.quote.changePct.toFixed(2)}%`
  if (!props.lastBar) return '--'
  return fmtPct(props.lastBar.close, props.prevClose || props.lastBar.open)
})

function fmtDate(d: string): string {
  if (d?.length === 8) return `${d.slice(0, 4)}-${d.slice(4, 6)}-${d.slice(6)}`
  return d
}

function fmtVol(v: number): string {
  if (v >= 10000) return `${(v / 10000).toFixed(1)}万手`
  return `${v.toFixed(0)}手`
}

function fmtAmount(a: number): string {
  // Tencent API returns amount in 万元
  if (a >= 10000) return `${(a / 10000).toFixed(2)}亿`
  return `${a.toFixed(2)}万`
}

function fmtMktCap(v: number): string {
  // Tencent API returns market cap in 亿元
  return `${v.toFixed(2)}亿`
}
</script>

<template>
  <el-card body-style="padding: 10px 16px">
    <div class="stock-row">
      <el-button link size="small" class="back-btn" @click="emit('back')">←</el-button>
      <span class="name">{{ stock.name }}</span>
      <span class="code">{{ stock.tsCode }}</span>
      <el-tag size="small" type="info">{{ stock.industry }}</el-tag>
      <span class="list-date">上市 {{ fmtDate(stock.listDate) }}</span>
      <template v-if="displayPrice !== undefined">
        <span class="price" :class="changeClass">{{ fmtPrice(displayPrice) }}</span>
        <span class="pct" :class="changeClass">{{ changePct }}</span>
      </template>
    </div>

    <div v-if="quote" class="quote-row">
      <span class="q-item">开 <b>{{ fmtPrice(quote.open) }}</b></span>
      <span class="q-item">高 <b :class="priceClass(quote.high, quote.prevClose)">{{ fmtPrice(quote.high) }}</b></span>
      <span class="q-item">低 <b :class="priceClass(quote.low, quote.prevClose)">{{ fmtPrice(quote.low) }}</b></span>
      <span class="q-item">昨收 <b>{{ fmtPrice(quote.prevClose) }}</b></span>
      <span class="q-item">量 <b>{{ fmtVol(quote.vol) }}</b></span>
      <span class="q-item">额 <b>{{ fmtAmount(quote.amount) }}</b></span>
      <span class="q-item">换手 <b>{{ quote.turnoverRate.toFixed(2) }}%</b></span>
      <span class="q-item">量比 <b>{{ quote.volRatio.toFixed(2) }}</b></span>
      <span class="q-item g-up">涨停 {{ fmtPrice(quote.limitUp) }}</span>
      <span class="q-item g-down">跌停 {{ fmtPrice(quote.limitDown) }}</span>
      <span class="q-time">{{ fmtQuoteTime(quote.quoteTime) }}</span>
      <el-button link size="small" class="expand-btn" @click="expanded = !expanded">
        {{ expanded ? '▲' : '▼' }}
      </el-button>
    </div>

    <div v-if="quote && expanded" class="expand-row">
      <span class="q-item">市盈率 <b>{{ quote.pe.toFixed(2) }}</b></span>
      <span class="q-item">市净率 <b>{{ quote.pb.toFixed(2) }}</b></span>
      <span class="q-item">振幅 <b>{{ quote.amplitude.toFixed(2) }}%</b></span>
      <span class="q-item">流通市值 <b>{{ fmtMktCap(quote.circMarketCap) }}</b></span>
      <span class="q-item">总市值 <b>{{ fmtMktCap(quote.totalMarketCap) }}</b></span>
      <span class="q-item">外盘 <b>{{ fmtVol(quote.outerVol) }}</b></span>
      <span class="q-item">内盘 <b>{{ fmtVol(quote.innerVol) }}</b></span>
      <span class="q-item">52周高 <b>{{ fmtPrice(quote.high52w) }}</b></span>
      <span class="q-item">52周低 <b>{{ fmtPrice(quote.low52w) }}</b></span>
    </div>
  </el-card>
</template>

<style scoped lang="scss">
.stock-row {
  display: flex;
  align-items: center;
  gap: 10px;
  white-space: nowrap;
}
.back-btn {
  padding: 0;
  color: #909399;
  font-size: 15px;
}
.name {
  font-size: 16px;
  font-weight: 600;
}
.code {
  font-size: 13px;
  color: #909399;
}
.list-date {
  font-size: 12px;
  color: #909399;
}
.price {
  font-size: 20px;
  font-weight: bold;
  margin-left: auto;
}
.pct {
  font-size: 13px;
}
.quote-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  margin-top: 6px;
  font-size: 12px;
  color: #606266;
}
.expand-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 4px;
  padding-top: 6px;
  border-top: 1px solid var(--el-border-color-lighter);
  font-size: 12px;
  color: #606266;
}
.q-item {
  white-space: nowrap;
  b {
    font-weight: 600;
  }
}
.q-time {
  color: #909399;
  font-size: 11px;
}
.expand-btn {
  margin-left: auto;
  padding: 0;
  font-size: 11px;
  color: #909399;
}
</style>
