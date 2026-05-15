<script setup lang="ts">
import { ref } from 'vue'
import { wMessage } from '@/utils/message'
import { changePassword } from '@/apis/me'

const form = ref({ old: '', new: '' })
const loading = ref(false)

async function submit() {
  if (!form.value.old || !form.value.new) {
    wMessage('warning', '请输入密码')
    return
  }
  loading.value = true
  try {
    await changePassword({ old: form.value.old, new: form.value.new })
    wMessage('success', '密码已修改')
    form.value = { old: '', new: '' }
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '修改失败'
    wMessage('error', msg)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <el-card>
    <template #header>{{ $t('profile.changePassword') }}</template>
    <el-form @submit.prevent="submit" style="max-width: 400px">
      <el-form-item :label="$t('profile.oldPassword')">
        <el-input v-model="form.old" type="password" />
      </el-form-item>
      <el-form-item :label="$t('profile.newPassword')">
        <el-input v-model="form.new" type="password" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="loading" @click="submit">{{ $t('common.save') }}</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>
