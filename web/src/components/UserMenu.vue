<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import GIcon from './GIcon.vue'

const router = useRouter()
const auth = useAuthStore()

function doLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <div class="user-menu">
    <el-dropdown trigger="click">
      <div class="user-trigger">
        <el-avatar :size="28" icon="UserFilled" />
        <span class="username">{{ auth.user?.username }}</span>
        <GIcon name="ArrowDown" />
      </div>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item @click="router.push('/profile')">{{ $t('menu.profile') }}</el-dropdown-item>
          <el-dropdown-item divided @click="doLogout">{{ $t('menu.logout') }}</el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </div>
</template>

<style scoped lang="scss">
.user-menu {
  padding: 0 16px;
}
.user-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: #303133;
}
.username {
  font-size: 14px;
}
</style>
