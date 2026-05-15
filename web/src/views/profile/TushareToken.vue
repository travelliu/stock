<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { wMessage } from '@/utils/message'
import { setTushareToken } from '@/apis/me'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const token = ref('')
const loading = ref(false)

onMounted(async () => {
  await auth.fetchMe()
  token.value = auth.user?.tushareToken || ''
})

async function submit() {
  loading.value = true
  try {
    await setTushareToken({ token: token.value })
    wMessage('success', 'Token 已保存')
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '保存失败'
    wMessage('error', msg)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <el-card>
    <template #header>{{ $t('profile.tushareToken') }}</template>
    <el-form @submit.prevent="submit" style="max-width: 400px">
      <el-form-item label="Token">
        <el-input v-model="token" :placeholder="$t('common.empty')" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="loading" @click="submit">{{ $t('common.save') }}</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>
