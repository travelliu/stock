<script setup lang="ts">
import { ref } from 'vue'
import type { AnalysisResult, DistBucket } from '@/types/api'

const props = defineProps<{ result: AnalysisResult | null }>()

const activeWindow = ref('All')

function windowName(id: string): string {
  const map: Record<string, string> = { All: '历史', last_90: '近3月', last_30: '近1月', last_15: '近2周' }
  return map[id] || id
}

function fmtRange(b: DistBucket): string {
  return `${(b.lower ?? 0).toFixed(2)}~${(b.upper ?? 0).toFixed(2)}`
}

function dist(windowId: string, key: 'spreadOH' | 'spreadOL'): DistBucket[] {
  const w = props.result?.windows?.find(w => w.info.id === windowId)
  return w?.means?.[key]?.distribution ?? []
}

function count(windowId: string, key: 'spreadOH' | 'spreadOL'): number {
  const w = props.result?.windows?.find(w => w.info.id === windowId)
  return w?.means?.[key]?.count ?? 0
}
</script>

<template>
  <el-card v-if="result?.windows?.length">
    <template #header>{{ $t('stockDetail.spreads') }}</template>
    <el-tabs v-model="activeWindow" size="small">
      <el-tab-pane
        v-for="w in result.windows"
        :key="w.info.id"
        :label="windowName(w.info.id)"
        :name="w.info.id"
      >
        <div class="dist-grid">
          <div class="dist-block">
            <div class="dist-title">最高-开盘（{{ count(w.info.id, 'spreadOH') }} 条）</div>
            <el-table :data="dist(w.info.id, 'spreadOH')" size="small" border>
              <el-table-column label="区间" min-width="120">
                <template #default="{ row }">{{ fmtRange(row) }}</template>
              </el-table-column>
              <el-table-column label="数量" align="right" width="60">
                <template #default="{ row }">{{ row.count ?? 0 }}</template>
              </el-table-column>
              <el-table-column label="占比" align="right" width="65">
                <template #default="{ row }">{{ (row.pct ?? 0).toFixed(1) }}%</template>
              </el-table-column>
            </el-table>
          </div>
          <div class="dist-block">
            <div class="dist-title">开盘-最低（{{ count(w.info.id, 'spreadOL') }} 条）</div>
            <el-table :data="dist(w.info.id, 'spreadOL')" size="small" border>
              <el-table-column label="区间" min-width="120">
                <template #default="{ row }">{{ fmtRange(row) }}</template>
              </el-table-column>
              <el-table-column label="数量" align="right" width="60">
                <template #default="{ row }">{{ row.count ?? 0 }}</template>
              </el-table-column>
              <el-table-column label="占比" align="right" width="65">
                <template #default="{ row }">{{ (row.pct ?? 0).toFixed(1) }}%</template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </el-card>
</template>

<style scoped lang="scss">
.dist-grid {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}
.dist-block {
  flex: 1;
  min-width: 220px;
}
.dist-title {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 6px;
}
</style>
