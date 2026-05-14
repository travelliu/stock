<template>
  <div>
    <h2>{{ tsCode }}</h2>
    <el-tabs>
      <el-tab-pane label="分析">
        <AnalysisPanel :ts-code="tsCode" />
      </el-tab-pane>
      <el-tab-pane label="历史">
        <HistoryPanel :ts-code="tsCode" />
      </el-tab-pane>
      <el-tab-pane label="草稿">
        <DraftPanel :ts-code="tsCode" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import AnalysisPanel from '@/components/AnalysisPanel.vue'
import HistoryPanel from '@/components/HistoryPanel.vue'
import DraftPanel from '@/components/DraftPanel.vue'

const route = useRoute()
const tsCode = route.params.tsCode as string
</script>
