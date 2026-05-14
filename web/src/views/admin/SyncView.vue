<template>
  <div>
    <h2>数据同步</h2>
    <el-button type="primary" @click="syncStocklist">同步股票列表</el-button>
    <el-button type="primary" @click="syncBars">同步行情数据</el-button>

    <h3 style="margin-top: 24px;">最近执行</h3>
    <el-table :data="jobs">
      <el-table-column prop="jobName" label="任务" />
      <el-table-column prop="status" label="状态" />
      <el-table-column prop="startedAt" label="开始时间" />
      <el-table-column prop="finishedAt" label="结束时间" />
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import client from '@/api/client'

interface Job {
  job_name: string
  status: string
  started_at: string
  finished_at: string
}

const jobs = ref<Job[]>([])

onMounted(async () => {
  await loadJob('daily-fetch')
  await loadJob('stocklist-sync')
})

async function loadJob(name: string) {
  try {
    const { data } = await client.get('/admin/sync/status', { params: { job: name } })
    if (data) jobs.value.push(data)
  } catch {
    // ignore
  }
}

async function syncStocklist() {
  await client.post('/admin/stocks/sync')
  ElMessage.success('股票列表同步已触发')
}

async function syncBars() {
  await client.post('/admin/bars/sync')
  ElMessage.success('行情数据同步已触发')
}
</script>
