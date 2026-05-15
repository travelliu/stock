<script setup lang="ts">
defineProps<{ line?: number }>()
</script>

<template>
  <div :style="{ overflow: 'hidden', textOverflow: 'ellipsis', display: '-webkit-box', WebkitLineClamp: line || 1, WebkitBoxOrient: 'vertical' }">
    <slot />
  </div>
</template>
