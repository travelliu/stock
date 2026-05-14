<template>
  <div>
    <el-form @submit.prevent="saveDraft">
      <el-form-item label="今开">
        <el-input-number v-model="form.open" :precision="2" />
      </el-form-item>
      <el-form-item label="最高">
        <el-input-number v-model="form.high" :precision="2" />
      </el-form-item>
      <el-form-item label="最低">
        <el-input-number v-model="form.low" :precision="2" />
      </el-form-item>
      <el-form-item label="收盘">
        <el-input-number v-model="form.close" :precision="2" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="saveDraft">保存</el-button>
        <el-button @click="clearDraft">清除</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import client from '@/api/client'

const props = defineProps<{ tsCode: string }>()

const form = ref({ open: undefined as number | undefined, high: undefined as number | undefined, low: undefined as number | undefined, close: undefined as number | undefined })

async function loadDraft() {
  try {
    const { data } = await client.get('/drafts/today', { params: { ts_code: props.tsCode } })
    form.value.open = data.open ?? undefined
    form.value.high = data.high ?? undefined
    form.value.low = data.low ?? undefined
    form.value.close = data.close ?? undefined
  } catch {
    // no draft yet
  }
}

async function saveDraft() {
  const today = new Date().toISOString().slice(0, 10).replace(/-/g, '')
  const body: any = { ts_code: props.tsCode, trade_date: today }
  if (form.value.open !== undefined) body.open = form.value.open
  if (form.value.high !== undefined) body.high = form.value.high
  if (form.value.low !== undefined) body.low = form.value.low
  if (form.value.close !== undefined) body.close = form.value.close
  await client.put('/drafts', body)
  ElMessage.success('草稿已保存')
}

async function clearDraft() {
  try {
    const { data } = await client.get('/drafts/today', { params: { ts_code: props.tsCode } })
    if (data.id) {
      await client.delete(`/drafts/${data.id}`)
      form.value = { open: undefined, high: undefined, low: undefined, close: undefined }
      ElMessage.success('草稿已清除')
    }
  } catch {
    // nothing to clear
  }
}

onMounted(loadDraft)
</script>
