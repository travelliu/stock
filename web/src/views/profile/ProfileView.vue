<script setup lang="ts">
import { useRoute } from 'vue-router'

const route = useRoute()

const menuItems = [
  { label: 'profile.info', path: '/profile' },
  { label: 'profile.changePassword', path: '/profile/password' },
  { label: 'profile.tushareToken', path: '/profile/token' },
  { label: 'profile.apiTokens', path: '/profile/api-tokens' },
]

const activeIndex = route.path
</script>

<template>
  <div class="profile-layout">
    <el-menu :default-active="activeIndex" :router="true" class="profile-menu">
      <el-menu-item v-for="item in menuItems" :key="item.path" :index="item.path">
        {{ $t(item.label) }}
      </el-menu-item>
    </el-menu>
    <div class="profile-content">
      <router-view />
    </div>
  </div>
</template>

<style scoped lang="scss">
.profile-layout {
  display: flex;
  gap: 16px;
}
.profile-menu {
  width: 200px;
  flex-shrink: 0;
}
.profile-content {
  flex: 1;
}
</style>
