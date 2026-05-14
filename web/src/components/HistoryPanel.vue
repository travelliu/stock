<template>
  <div>
    <el-table :data="bars" height="500">
      <el-table-column prop="tradeDate" label="日期" />
      <el-table-column prop="open" label="开盘" />
      <el-table-column prop="high" label="最高" />
      <el-table-column prop="low" label="最低" />
      <el-table-column prop="close" label="收盘" />
      <el-table-column prop="vol" label="成交量" />
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import client from '@/api/client'

const props = defineProps<{ tsCode: string }>()
const bars = ref([])

onMounted(async () => {
  const { data } = await client.get(`/bars/${props.tsCode}`)
  bars.value = data
})
</script>
