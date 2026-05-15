<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { wMessage } from '@/utils/message'
import { queryBars } from '@/apis/stocks'
import type { DailyBar } from '@/types/api'
import DailyBarTable from '@/components/DailyBarTable.vue'

const route = useRoute()
const code = route.params.code as string

const bars = ref<DailyBar[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const result = await queryBars(code, { page: page.value, limit: pageSize })
    bars.value = result.items
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

onMounted(load)
</script>

<template>
  <div v-loading="loading">
    <DailyBarTable :bars="bars" />
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
