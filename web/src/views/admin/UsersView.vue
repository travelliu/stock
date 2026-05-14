<template>
  <div>
    <h2>用户管理</h2>
    <el-button type="primary" @click="showCreate = true">新建用户</el-button>
    <el-table :data="users" style="margin-top: 16px;">
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="role" label="角色" />
      <el-table-column prop="disabled" label="状态">
        <template #default="{ row }">
          {{ row.disabled ? '禁用' : '正常' }}
        </template>
      </el-table-column>
      <el-table-column label="操作">
        <template #default="{ row }">
          <el-button link @click="toggleDisabled(row)">{{ row.disabled ? '启用' : '禁用' }}</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showCreate" title="新建用户" width="400px">
      <el-form @submit.prevent="createUser">
        <el-form-item label="用户名">
          <el-input v-model="newUser.username" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="newUser.password" type="password" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="newUser.role">
            <el-option label="用户" value="user" />
            <el-option label="管理员" value="admin" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" @click="createUser">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import client from '@/api/client'

const users = ref([])
const showCreate = ref(false)
const newUser = ref({ username: '', password: '', role: 'user' })

onMounted(fetchUsers)

async function fetchUsers() {
  const { data } = await client.get('/admin/users')
  users.value = data
}

async function createUser() {
  await client.post('/admin/users', newUser.value)
  showCreate.value = false
  newUser.value = { username: '', password: '', role: 'user' }
  ElMessage.success('创建成功')
  await fetchUsers()
}

async function toggleDisabled(row: any) {
  await client.patch(`/admin/users/${row.id}`, { disabled: !row.disabled })
  await fetchUsers()
}
</script>
