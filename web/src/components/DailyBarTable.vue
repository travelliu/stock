<script setup lang="ts">
import type { DailyBar } from '@/types/api'
import { fmtPrice, priceClass } from '@/utils/format'

defineProps<{ bars: DailyBar[] }>()

const spreadKeys = ['oh', 'ol', 'hl', 'oc', 'hc', 'lc'] as const
</script>

<template>
  <el-table :data="bars" style="width: 100%" size="small">
    <el-table-column prop="tradeDate" :label="$t('stockDetail.date')" width="100" />
    <el-table-column prop="open" :label="$t('stockDetail.open')" width="80">
      <template #default="{ row }">
        <span :class="priceClass(row.close, row.open)">{{ fmtPrice(row.open) }}</span>
      </template>
    </el-table-column>
    <el-table-column prop="high" :label="$t('stockDetail.high')" width="80">
      <template #default="{ row }">
        <span :class="priceClass(row.close, row.open)">{{ fmtPrice(row.high) }}</span>
      </template>
    </el-table-column>
    <el-table-column prop="low" :label="$t('stockDetail.low')" width="80">
      <template #default="{ row }">
        <span :class="priceClass(row.close, row.open)">{{ fmtPrice(row.low) }}</span>
      </template>
    </el-table-column>
    <el-table-column prop="close" :label="$t('stockDetail.close')" width="80">
      <template #default="{ row }">
        <span :class="priceClass(row.close, row.open)">{{ fmtPrice(row.close) }}</span>
      </template>
    </el-table-column>
    <el-table-column prop="vol" :label="$t('stockDetail.vol')" width="100" />
    <el-table-column v-for="k in spreadKeys" :key="k" :label="$t(`stockDetail.${k}`)" width="80">
      <template #default="{ row }">
        {{ fmtPrice(row.spreads[k]) }}
      </template>
    </el-table-column>
  </el-table>
</template>
