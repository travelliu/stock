<script setup lang="ts">
import { computed } from 'vue'
import type { DailyBar } from '@/types/api'

const props = defineProps<{ bars: DailyBar[] }>()

const labels = [
  { key: 'oh' as const, name: '高-开' },
  { key: 'ol' as const, name: '开-低' },
  { key: 'hl' as const, name: '高-低' },
  { key: 'oc' as const, name: '开-收' },
  { key: 'hc' as const, name: '高-收' },
  { key: 'lc' as const, name: '低-收' },
]

const stats = computed(() => {
  if (!props.bars.length) return []
  const max = Math.max(...labels.map(l => Math.max(...props.bars.map(b => b.spreads[l.key]))))
  return labels.map(l => {
    const avg = props.bars.reduce((s, b) => s + b.spreads[l.key], 0) / props.bars.length
    const pct = max > 0 ? (avg / max) * 100 : 0
    return { name: l.name, avg, pct }
  })
})
</script>

<template>
  <el-card>
    <template #header>{{ $t('stockDetail.spreads') }}</template>
    <div class="histogram">
      <div v-for="s in stats" :key="s.name" class="bar-row">
        <span class="label">{{ s.name }}</span>
        <el-progress :percentage="Math.round(s.pct)" :stroke-width="16" :show-text="false" />
        <span class="value">{{ s.avg.toFixed(2) }}</span>
      </div>
    </div>
  </el-card>
</template>

<style scoped lang="scss">
.histogram {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.bar-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.label {
  width: 60px;
  font-size: 13px;
  flex-shrink: 0;
}
.value {
  width: 50px;
  text-align: right;
  font-size: 13px;
  flex-shrink: 0;
}
</style>
