<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { wMessage } from '@/utils/message'
import { getAnalysis } from '@/apis/analysis'
import type { AnalysisResult } from '@/types/api'
import DraftFormBlock from '@/components/DraftFormBlock.vue'
import SpreadModelTable from '@/components/SpreadModelTable.vue'
import TradePlanTable from '@/components/TradePlanTable.vue'

const props = defineProps<{ tsCode: string }>()

const result = ref<AnalysisResult | null>(null)
const loading = ref(false)

async function runAnalysis(params?: { open?: number; high?: number; low?: number; close?: number }) {
  loading.value = true
  try {
    result.value = await getAnalysis(props.tsCode, {
      actualOpen: params?.open,
      actualHigh: params?.high,
      actualLow: params?.low,
      actualClose: params?.close,
      withDraft: true,
    })
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '分析失败'
    wMessage('error', msg)
  } finally {
    loading.value = false
  }
}

onMounted(() => runAnalysis())
</script>

<template>
  <div v-loading="loading">
    <DraftFormBlock :ts-code="tsCode" @apply="runAnalysis" />
    <div style="margin-top: 16px">
      <SpreadModelTable :result="result" />
    </div>
    <TradePlanTable :result="result" />
  </div>
</template>
