<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { wMessage } from '@/utils/message'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const username = ref('')
const password = ref('')
const loading = ref(false)

async function handleLogin() {
  if (!username.value || !password.value) {
    wMessage('warning', '请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    await authStore.login(username.value, password.value)
    router.push('/stocks')
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '登录失败'
    wMessage('error', msg)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-container">
    <el-card class="login-card" shadow="always">
      <h2 class="title">{{ $t('login.title') }}</h2>
      <el-form @submit.prevent="handleLogin">
        <el-form-item>
          <el-input v-model="username" :placeholder="$t('login.username')" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="password" type="password" :placeholder="$t('login.password')" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" style="width: 100%" :loading="loading" @click="handleLogin">
            {{ $t('login.login') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped lang="scss">
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  width: 100vw;
  background: var(--content-bg);
}
.login-card {
  width: 360px;
}
.title {
  text-align: center;
  margin-bottom: 24px;
  font-weight: 500;
}
</style>
