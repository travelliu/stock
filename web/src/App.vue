<template>
  <el-container v-if="authStore.user" style="height: 100vh;">
    <el-aside width="200px">
      <el-menu :router="true" :default-active="$route.path">
        <el-menu-item index="/portfolio">
          <span>我的持仓</span>
        </el-menu-item>
        <el-menu-item index="/settings">
          <span>设置</span>
        </el-menu-item>
        <el-menu-item v-if="authStore.user?.role === 'admin'" index="/admin/users">
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item v-if="authStore.user?.role === 'admin'" index="/admin/sync">
          <span>数据同步</span>
        </el-menu-item>
        <el-menu-item @click="authStore.logout">
          <span>退出</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-main>
      <router-view />
    </el-main>
  </el-container>
  <router-view v-else />
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from './stores/auth'

const authStore = useAuthStore()

onMounted(() => {
  authStore.fetchMe()
})
</script>
