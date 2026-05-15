<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import ConsoleMenu from '@/components/ConsoleMenu.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

onMounted(() => {
  auth.fetchMe().then(() => {
    if (!auth.user && route.meta.requiresAuth) {
      router.push('/login')
    }
  })
})
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
  padding: 16px;
  overflow: auto;
}
</style>
