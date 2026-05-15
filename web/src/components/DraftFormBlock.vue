<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { wMessage } from '@/utils/message'
import { getDraftToday, upsertDraft, deleteDraft } from '@/apis/draft'

const props = defineProps<{ tsCode: string }>()
const emit = defineEmits<{
  apply: [params: { open?: number; high?: number; low?: number; close?: number }]
}>()

const form = ref<{ open?: number; high?: number; low?: number; close?: number }>({})
const draftId = ref<number | null>(null)
const loading = ref(false)

async function load() {
  try {
    const d = await getDraftToday(props.tsCode)
    draftId.value = d.id
    form.value = {
      open: d.open ?? undefined,
      high: d.high ?? undefined,
      low: d.low ?? undefined,
      close: d.close ?? undefined,
    }
  } catch {
    draftId.value = null
    form.value = {}
  }
}

async function save() {
  loading.value = true
  try {
    const today = new Date().toISOString().slice(0, 10).replace(/-/g, '')
    const body: Record<string, unknown> = { tsCode: props.tsCode, tradeDate: today }
    if (form.value.open !== undefined) body.open = form.value.open
    if (form.value.high !== undefined) body.high = form.value.high
    if (form.value.low !== undefined) body.low = form.value.low
    if (form.value.close !== undefined) body.close = form.value.close
    await upsertDraft(body)
    wMessage('success', '草稿已保存')
    await load()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '保存失败'
    wMessage('error', msg)
  } finally {
    loading.value = false
  }
}

async function clear() {
  if (!draftId.value) {
    form.value = {}
    return
  }
  try {
    await deleteDraft(draftId.value)
    draftId.value = null
    form.value = {}
    wMessage('success', '草稿已清除')
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '清除失败'
    wMessage('error', msg)
  }
}

function apply() {
  emit('apply', { ...form.value })
}

onMounted(load)
</script>

<template>
  <el-card>
    <template #header>{{ $t('stockDetail.draft') }}</template>
    <el-form inline @submit.prevent="save">
      <el-form-item :label="$t('stockDetail.open')">
        <el-input-number v-model="form.open" :precision="2" :controls="false" />
      </el-form-item>
      <el-form-item :label="$t('stockDetail.high')">
        <el-input-number v-model="form.high" :precision="2" :controls="false" />
      </el-form-item>
      <el-form-item :label="$t('stockDetail.low')">
        <el-input-number v-model="form.low" :precision="2" :controls="false" />
      </el-form-item>
      <el-form-item :label="$t('stockDetail.close')">
        <el-input-number v-model="form.close" :precision="2" :controls="false" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="loading" @click="save">{{ $t('stockDetail.draftSave') }}</el-button>
        <el-button @click="apply">{{ $t('stockDetail.draftApply') }}</el-button>
        <el-button @click="clear">{{ $t('common.delete') }}</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>
