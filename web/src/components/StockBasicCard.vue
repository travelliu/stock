<script setup lang="ts">
import { computed } from 'vue'
import type { Stock, DailyBar, RealtimeQuote } from '@/types/api'
import { fmtPrice, fmtPct, priceClass } from '@/utils/format'

const props = defineProps<{ stock: Stock; lastBar?: DailyBar; prevClose?: number; quote?: RealtimeQuote }>()
const emit = defineEmits<{ back: [] }>()

const displayPrice = computed(() => {
  if (props.quote) return props.quote.price
  return props.lastBar?.close
})

const changeClass = computed(() => {
  if (props.quote) {
    return priceClass(props.quote.price, props.quote.prevClose)
  }
  if (!props.lastBar) return 'g-flat'
  const base = props.prevClose || props.lastBar.open
  return priceClass(props.lastBar.close, base)
})

const changePct = computed(() => {
  if (props.quote) {
    return `${props.quote.changePct >= 0 ? '+' : ''}${props.quote.changePct.toFixed(2)}%`
  }
  if (!props.lastBar) return '--'
  const base = props.prevClose || props.lastBar.open
  return fmtPct(props.lastBar.close, base)
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
  if (a >= 10000) return `${(a / 10000).toFixed(2)}亿`
  return `${a.toFixed(0)}万`
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
      <span class="q-item g-up">涨停 {{ fmtPrice(quote.limitUp) }}</span>
      <span class="q-item g-down">跌停 {{ fmtPrice(quote.limitDown) }}</span>
      <span class="q-time">{{ quote.quoteTime }}</span>
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
  gap: 12px;
  margin-top: 6px;
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
  margin-left: auto;
  color: #909399;
  font-size: 11px;
}
</style>
