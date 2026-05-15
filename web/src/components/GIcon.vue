<script setup lang="ts">
import { computed } from 'vue'
import * as Icons from '@element-plus/icons-vue'

const props = defineProps<{ name: string }>()
const iconComponent = computed(() => (Icons as Record<string, unknown>)[props.name])
</script>

<template>
  <el-icon>
    <component :is="iconComponent" />
  </el-icon>
</template>
