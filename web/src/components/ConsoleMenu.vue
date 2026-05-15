<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useLangStore } from '@/stores/lang'
import { i18n } from '@/intl'
import GIcon from './GIcon.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const lang = useLangStore()

const activeIndex = computed(() => route.path)

function switchLang() {
  const next = lang.lang === 'zh' ? 'en' : 'zh'
  lang.setLang(next)
  i18n.global.locale.value = next
}
</script>

<template>
  <div class="console-menu">
    <div class="logo" @click="router.push('/')">
      <img src="@/assets/image/logo-mini.svg" alt="logo" />
    </div>

    <div class="menu-center">
      <div class="menu-item" :class="{ active: activeIndex.startsWith('/stocks') }" @click="router.push('/stocks')">
        <GIcon name="TrendCharts" />
        <span>{{ $t('menu.stock') }}</span>
      </div>
      <template v-if="auth.user?.role === 'admin'">
        <div class="menu-item" :class="{ active: activeIndex.startsWith('/admin/users') }" @click="router.push('/admin/users')">
          <GIcon name="UserFilled" />
          <span>{{ $t('menu.users') }}</span>
        </div>
        <div class="menu-item" :class="{ active: activeIndex.startsWith('/admin/sync') }" @click="router.push('/admin/sync')">
          <GIcon name="Refresh" />
          <span>{{ $t('menu.sync') }}</span>
        </div>
      </template>
    </div>

    <div class="menu-bottom">
      <div v-if="auth.user" class="menu-item" :class="{ active: activeIndex.startsWith('/profile') }" @click="router.push('/profile')">
        <GIcon name="User" />
        <span>{{ $t('menu.profile') }}</span>
      </div>
      <div class="menu-item" @click="switchLang">
        <GIcon name="Globe" />
        <span>{{ lang.lang === 'zh' ? '中' : 'EN' }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.console-menu {
  width: 65px;
  height: 100vh;
  background: var(--sider-bg);
  color: var(--sider-text);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 0;
  flex-shrink: 0;
}
.logo {
  padding: 12px 0;
  cursor: pointer;
}
.menu-center {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 24px;
  gap: 8px;
}
.menu-bottom {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding-bottom: 8px;
}
.menu-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px 4px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  width: 52px;
  transition: background 0.2s;
}
.menu-item:hover {
  background: var(--sider-bg-hover);
}
.menu-item.active {
  color: var(--sider-active);
}
</style>
