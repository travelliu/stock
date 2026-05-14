<template>
  <div>
    <h3>价差模型</h3>
    <el-form inline @submit.prevent="runAnalysis">
      <el-form-item label="今开">
        <el-input-number v-model="params.open" :precision="2" />
      </el-form-item>
      <el-form-item label="最高">
        <el-input-number v-model="params.high" :precision="2" />
      </el-form-item>
      <el-form-item label="最低">
        <el-input-number v-model="params.low" :precision="2" />
      </el-form-item>
      <el-form-item label="收盘">
        <el-input-number v-model="params.close" :precision="2" />
      </el-form-item>
      <el-form-item>
        <el-checkbox v-model="useDraft">使用草稿</el-checkbox>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="runAnalysis">分析</el-button>
      </el-form-item>
    </el-form>

    <el-table v-if="result" :data="tableRows" border>
      <el-table-column prop="label" label="指标" />
      <el-table-column v-for="w in result.windows" :key="w" :label="w">
        <template #default="{ row }">
          {{ row.values[result.windows.indexOf(w)] }}
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import client from '@/api/client'

const props = defineProps<{ tsCode: string }>()

const params = ref({ open: undefined as number | undefined, high: undefined as number | undefined, low: undefined as number | undefined, close: undefined as number | undefined })
const useDraft = ref(true)
const result = ref<any>(null)

const tableRows = computed(() => {
  if (!result.value) return []
  const mt = result.value.model_table
  return mt.rows.map((row: string[]) => ({
    label: row[0],
    values: row.slice(1),
  }))
})

async function runAnalysis() {
  const qs = new URLSearchParams()
  if (params.value.open !== undefined) qs.set('actual_open', String(params.value.open))
  if (params.value.high !== undefined) qs.set('actual_high', String(params.value.high))
  if (params.value.low !== undefined) qs.set('actual_low', String(params.value.low))
  if (params.value.close !== undefined) qs.set('actual_close', String(params.value.close))
  qs.set('with_draft', String(useDraft.value))
  const { data } = await client.get(`/analysis/${props.tsCode}?${qs.toString()}`)
  result.value = data
}

runAnalysis()
</script>
