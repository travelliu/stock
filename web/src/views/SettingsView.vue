<template>
  <div>
    <h2>设置</h2>

    <el-card title="修改密码" style="margin-bottom: 16px;">
      <el-form @submit.prevent="changePassword">
        <el-form-item label="原密码">
          <el-input v-model="pwdForm.old" type="password" />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="pwdForm.new" type="password" />
        </el-form-item>
        <el-button type="primary" @click="changePassword">修改密码</el-button>
      </el-form>
    </el-card>

    <el-card title="Tushare Token" style="margin-bottom: 16px;">
      <el-form @submit.prevent="saveToken">
        <el-form-item label="Token">
          <el-input v-model="tushareToken" placeholder="可选，覆盖服务器默认值" />
        </el-form-item>
        <el-button type="primary" @click="saveToken">保存</el-button>
      </el-form>
    </el-card>

    <el-card title="API Tokens">
      <el-table :data="tokens" style="margin-bottom: 16px;">
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="createdAt" label="创建时间" />
        <el-table-column label="操作">
          <template #default="{ row }">
            <el-button link type="danger" @click="revokeToken(row.id)">撤销</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-button type="primary" @click="showIssue = true">新建 Token</el-button>
    </el-card>

    <el-dialog v-model="showIssue" title="新建 API Token" width="400px">
      <el-form @submit.prevent="issueToken">
        <el-form-item label="名称">
          <el-input v-model="newTokenName" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showIssue = false">取消</el-button>
        <el-button type="primary" @click="issueToken">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showNewToken" title="API Token" width="400px" :close-on-click-modal="false">
      <p>请立即复制，只会显示一次：</p>
      <el-input v-model="issuedToken" readonly />
      <template #footer>
        <el-button @click="showNewToken = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import client from '@/api/client'

const pwdForm = ref({ old: '', new: '' })
const tushareToken = ref('')
const tokens = ref([])
const showIssue = ref(false)
const showNewToken = ref(false)
const newTokenName = ref('')
const issuedToken = ref('')

onMounted(async () => {
  await loadTokens()
  try {
    const { data } = await client.get('/auth/me')
    tushareToken.value = data.tushareToken || ''
  } catch {
    // ignore
  }
})

async function loadTokens() {
  const { data } = await client.get('/me/tokens')
  tokens.value = data
}

async function changePassword() {
  await client.post('/me/password', pwdForm.value)
  ElMessage.success('密码已修改')
  pwdForm.value = { old: '', new: '' }
}

async function saveToken() {
  await client.patch('/me/tushare_token', { token: tushareToken.value })
  ElMessage.success('Token 已保存')
}

async function issueToken() {
  const { data } = await client.post('/me/tokens', { name: newTokenName.value })
  issuedToken.value = data.token
  showIssue.value = false
  showNewToken.value = true
  newTokenName.value = ''
  await loadTokens()
}

async function revokeToken(id: number) {
  await client.delete(`/me/tokens/${id}`)
  ElMessage.success('已撤销')
  await loadTokens()
}
</script>
