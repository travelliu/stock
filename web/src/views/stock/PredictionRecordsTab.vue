<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { wMessage } from '@/utils/message'
import { getPredictions } from '@/apis/analysis'
import type { AnalysisPrediction } from '@/types/api'

const route = useRoute()
const code = route.params.code as string

const preds = ref<AnalysisPrediction[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const result = await getPredictions(code, { page: page.value, limit: pageSize })
    preds.value = result.items
    total.value = Number(result.total)
  } catch (e: unknown) {
    wMessage('error', e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function onPageChange(p: number) {
  page.value = p
  load()
}

function fmt(v: number): string {
  return v ? v.toFixed(2) : '-'
}

function predColor(pred: number, actual: number): string {
  if (!pred || !actual) return ''
  if (pred > actual) return 'over'
  if (pred < actual) return 'under'
  return ''
}

onMounted(load)
</script>

<template>
  <div v-loading="loading">
    <el-table :data="preds" size="small" border>
      <el-table-column label="日期" prop="tradeDate" width="110" />
      <el-table-column label="开盘" align="right" width="80">
        <template #default="{ row }">{{ fmt(row.openPrice) }}</template>
      </el-table-column>
      <el-table-column label="预测高" align="right" width="80">
        <template #default="{ row }">
          <span :class="predColor(row.predictHigh, row.actualHigh)">{{ fmt(row.predictHigh) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="实际高" align="right" width="80">
        <template #default="{ row }">{{ fmt(row.actualHigh) }}</template>
      </el-table-column>
      <el-table-column label="预测低" align="right" width="80">
        <template #default="{ row }">
          <span :class="predColor(row.predictLow, row.actualLow)">{{ fmt(row.predictLow) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="实际低" align="right" width="80">
        <template #default="{ row }">{{ fmt(row.actualLow) }}</template>
      </el-table-column>
      <el-table-column label="预测收" align="right" width="80">
        <template #default="{ row }">
          <span :class="predColor(row.predictClose, row.actualClose)">{{ fmt(row.predictClose) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="实际收" align="right" width="80">
        <template #default="{ row }">{{ fmt(row.actualClose) }}</template>
      </el-table-column>
    </el-table>
    <el-pagination
      style="margin-top: 16px; display: flex; justify-content: flex-end"
      :current-page="page"
      :page-size="pageSize"
      :total="total"
      layout="prev, pager, next, total"
      @current-change="onPageChange"
    />
  </div>
</template>

<style scoped>
.over  { color: var(--el-color-danger);  }
.under { color: var(--el-color-success); }
</style>
