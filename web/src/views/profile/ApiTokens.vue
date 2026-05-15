<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { wMessage } from '@/utils/message'
import { listTokens, issueToken, revokeToken } from '@/apis/me'
import type { APIToken } from '@/types/api'

const tokens = ref<APIToken[]>([])
const showIssue = ref(false)
const showNewToken = ref(false)
const newTokenName = ref('')
const issuedToken = ref('')
const loading = ref(false)

async function load() {
  tokens.value = await listTokens()
}

async function create() {
  if (!newTokenName.value) return
  loading.value = true
  try {
    const res = await issueToken({ name: newTokenName.value })
    issuedToken.value = res.token
    showIssue.value = false
    showNewToken.value = true
    newTokenName.value = ''
    await load()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '创建失败'
    wMessage('error', msg)
  } finally {
    loading.value = false
  }
}

async function revoke(id: number) {
  try {
    await revokeToken(id)
    wMessage('success', '已撤销')
    await load()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '撤销失败'
    wMessage('error', msg)
  }
}

onMounted(load)
</script>

<template>
  <el-card>
    <template #header>{{ $t('profile.apiTokens') }}</template>
    <el-table :data="tokens" style="margin-bottom: 16px">
      <el-table-column prop="name" :label="$t('profile.tokenName')" />
      <el-table-column prop="createdAt" :label="$t('profile.createdAt')" />
      <el-table-column :label="$t('stockList.action')" width="100">
        <template #default="{ row }">
          <el-button link type="danger" @click="revoke(row.id)">{{ $t('profile.revoke') }}</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-button type="primary" @click="showIssue = true">{{ $t('profile.createToken') }}</el-button>

    <el-dialog v-model="showIssue" :title="$t('profile.createToken')" width="400px">
      <el-form @submit.prevent="create">
        <el-form-item :label="$t('profile.tokenName')">
          <el-input v-model="newTokenName" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showIssue = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="loading" @click="create">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showNewToken" :title="$t('profile.apiTokens')" width="400px" :close-on-click-modal="false">
      <p>{{ $t('profile.copyPrompt') }}</p>
      <el-input v-model="issuedToken" readonly />
      <template #footer>
        <el-button @click="showNewToken = false">{{ $t('common.close') }}</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>
