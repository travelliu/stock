<script setup lang="ts">
import { computed } from 'vue'
import type { Stock, DailyBar } from '@/types/api'
import { fmtPrice, fmtPct, priceClass } from '@/utils/format'

const props = defineProps<{ stock: Stock; lastBar?: DailyBar; prevClose?: number }>()

const changeClass = computed(() => {
  if (!props.lastBar) return 'g-flat'
  const base = props.prevClose || props.lastBar.open
  return priceClass(props.lastBar.close, base)
})

const changePct = computed(() => {
  if (!props.lastBar) return '--'
  const base = props.prevClose || props.lastBar.open
  return fmtPct(props.lastBar.close, base)
})
</script>

<template>
  <el-card>
    <div class="stock-card">
      <div class="name-block">
        <h3>{{ stock.name }} <span class="code">{{ stock.tsCode }}</span></h3>
        <div class="meta">
          <el-tag size="small">{{ stock.industry }}</el-tag>
          <span>{{ $t('stockDetail.listDate') }}: {{ stock.listDate }}</span>
        </div>
      </div>
      <div v-if="lastBar" class="price-block">
        <div class="price" :class="changeClass">{{ fmtPrice(lastBar.close) }}</div>
        <div class="pct" :class="changeClass">{{ changePct }}</div>
      </div>
    </div>
  </el-card>
</template>

<style scoped lang="scss">
.stock-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.name-block h3 {
  margin: 0 0 8px 0;
}
.code {
  font-size: 14px;
  color: #909399;
  font-weight: normal;
}
.meta {
  display: flex;
  gap: 12px;
  align-items: center;
  font-size: 13px;
  color: #606266;
}
.price-block {
  text-align: right;
}
.price {
  font-size: 28px;
  font-weight: bold;
}
.pct {
  font-size: 14px;
}
</style>
