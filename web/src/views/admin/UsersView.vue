<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { wMessage } from '@/utils/message'
import { listUsers, createUser, patchUser, deleteUser } from '@/apis/admin'
import type { User, CreateUserReq } from '@/types/api'

const users = ref<User[]>([])
const showCreate = ref(false)
const newUser = ref<CreateUserReq>({ username: '', password: '', role: 'user' })
const loading = ref(false)

async function fetchUsers() {
  users.value = await listUsers()
}

async function doCreate() {
  loading.value = true
  try {
    await createUser(newUser.value)
    wMessage('success', '创建成功')
    showCreate.value = false
    newUser.value = { username: '', password: '', role: 'user' }
    await fetchUsers()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '创建失败'
    wMessage('error', msg)
  } finally {
    loading.value = false
  }
}

async function toggle(row: User) {
  try {
    await patchUser(row.id, { disabled: !row.disabled })
    await fetchUsers()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '操作失败'
    wMessage('error', msg)
  }
}

async function remove(row: User) {
  try {
    await deleteUser(row.id)
    await fetchUsers()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '删除失败'
    wMessage('error', msg)
  }
}

onMounted(fetchUsers)
</script>

<template>
  <div>
    <div class="header-bar">
      <h2>{{ $t('admin.users') }}</h2>
      <el-button type="primary" @click="showCreate = true">{{ $t('admin.createUser') }}</el-button>
    </div>

    <el-table :data="users" style="margin-top: 16px">
      <el-table-column prop="username" :label="$t('profile.username')" />
      <el-table-column prop="role" :label="$t('profile.role')" />
      <el-table-column :label="$t('admin.status')">
        <template #default="{ row }">
          {{ row.disabled ? $t('admin.disabled') : $t('admin.enabled') }}
        </template>
      </el-table-column>
      <el-table-column :label="$t('stockList.action')" width="180">
        <template #default="{ row }">
          <el-button link @click="toggle(row)">{{ row.disabled ? $t('admin.enabled') : $t('admin.disabled') }}</el-button>
          <el-button link type="danger" @click="remove(row)">{{ $t('common.delete') }}</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showCreate" :title="$t('admin.createUser')" width="400px">
      <el-form @submit.prevent="doCreate">
        <el-form-item :label="$t('profile.username')">
          <el-input v-model="newUser.username" />
        </el-form-item>
        <el-form-item :label="$t('profile.newPassword')">
          <el-input v-model="newUser.password" type="password" />
        </el-form-item>
        <el-form-item :label="$t('profile.role')">
          <el-select v-model="newUser.role">
            <el-option :label="$t('admin.enabled')" value="user" />
            <el-option :label="$t('admin.users')" value="admin" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="loading" @click="doCreate">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
