<script setup lang="ts">
import { useRoute } from 'vue-router'
import BasicTab from './BasicTab.vue'
import StatisticsTab from './StatisticsTab.vue'
import PredictionsTab from './PredictionsTab.vue'

const route = useRoute()
const tsCode = route.params.tsCode as string
</script>

<template>
  <div>
    <el-tabs type="border-card">
      <el-tab-pane :label="$t('stockDetail.basic')">
        <BasicTab :ts-code="tsCode" />
      </el-tab-pane>
      <el-tab-pane :label="$t('stockDetail.statistics')">
        <StatisticsTab :ts-code="tsCode" />
      </el-tab-pane>
      <el-tab-pane :label="$t('stockDetail.predictions')">
        <PredictionsTab :ts-code="tsCode" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>
