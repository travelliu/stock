<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const user = computed(() => auth.user)
</script>

<template>
  <el-card v-if="user">
    <el-descriptions :title="$t('profile.info')" :column="1">
      <el-descriptions-item :label="$t('profile.username')">{{ user.username }}</el-descriptions-item>
      <el-descriptions-item :label="$t('profile.role')">{{ user.role }}</el-descriptions-item>
      <el-descriptions-item :label="$t('profile.registeredAt')">{{ user.createdAt }}</el-descriptions-item>
    </el-descriptions>
  </el-card>
</template>
