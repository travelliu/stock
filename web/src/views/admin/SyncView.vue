<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { wMessage } from '@/utils/message'
import { syncStocklist, syncBars, jobStatus } from '@/apis/admin'
import type { JobRun } from '@/types/api'

const jobs = ref<JobRun[]>([])
const jobNames = ['daily-fetch', 'stocklist-sync']

async function loadJobs() {
  jobs.value = []
  for (const name of jobNames) {
    try {
      const j = await jobStatus(name)
      jobs.value.push(j)
    } catch {
      // ignore
    }
  }
}

async function doSyncStocklist() {
  try {
    await syncStocklist()
    wMessage('success', '股票列表同步已触发')
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '同步失败'
    wMessage('error', msg)
  }
}

async function doSyncBars() {
  try {
    await syncBars()
    wMessage('success', '行情数据同步已触发')
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '同步失败'
    wMessage('error', msg)
  }
}

onMounted(loadJobs)
</script>

<template>
  <div>
    <h2>{{ $t('admin.sync') }}</h2>
    <el-button type="primary" @click="doSyncStocklist">{{ $t('admin.syncStocklist') }}</el-button>
    <el-button type="primary" @click="doSyncBars">{{ $t('admin.syncBars') }}</el-button>

    <h3 style="margin-top: 24px">{{ $t('admin.job') }}</h3>
    <el-table :data="jobs">
      <el-table-column prop="jobName" :label="$t('admin.job')" />
      <el-table-column prop="status" :label="$t('admin.status')" />
      <el-table-column prop="startedAt" :label="$t('admin.startedAt')" />
      <el-table-column prop="finishedAt" :label="$t('admin.finishedAt')" />
    </el-table>
  </div>
</template>
