<script setup lang="ts">
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import ConsoleMenu from '@/components/ConsoleMenu.vue'

const route = useRoute()
const auth = useAuthStore()
</script>

<template>
  <div class="app-root">
    <template v-if="auth.user && route.path !== '/login'">
      <ConsoleMenu />
      <div class="main">
        <router-view />
      </div>
    </template>
    <template v-else>
      <router-view />
    </template>
  </div>
</template>

<style scoped lang="scss">
.app-root {
  display: flex;
  min-height: 100vh;
}
.main {
  flex: 1;
  padding: 12px;
  overflow: auto;
}
</style>
