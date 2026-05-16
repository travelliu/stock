<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getConceptBlocks, getFundFlow } from '@/apis/stocks'
import type { ConceptBlocks, FundFlowDay } from '@/types/api'
import { wMessage } from '@/utils/message'

const route = useRoute()
const code = route.params.code as string

const blocks = ref<ConceptBlocks | null>(null)
const fundFlow = ref<FundFlowDay[]>([])
const loading = ref(false)

function pctClass(pct: string): string {
  const n = parseFloat(pct)
  if (isNaN(n)) return ''
  return n >= 0 ? 'up' : 'down'
}

function fmtPct(pct: string): string {
  const n = parseFloat(pct)
  if (isNaN(n)) return pct
  return (n >= 0 ? '+' : '') + n.toFixed(2) + '%'
}

function fmtWan(val: string): string {
  const n = parseFloat(val)
  if (isNaN(n)) return val
  const abs = Math.abs(n)
  if (abs >= 10000) return (n >= 0 ? '+' : '') + (n / 10000).toFixed(2) + '亿'
  return (n >= 0 ? '+' : '') + n.toFixed(0) + '万'
}

onMounted(async () => {
  loading.value = true
  try {
    const [b, f] = await Promise.all([
      getConceptBlocks(code),
      getFundFlow(code),
    ])
    blocks.value = b
    fundFlow.value = f
  } catch (e: unknown) {
    wMessage('error', e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div v-loading="loading" class="blocks-fund-tab">

    <!-- 所属板块 -->
    <template v-if="blocks">
      <section class="section">
        <h3 class="section-title">所属板块</h3>

        <template v-if="blocks.industry?.length">
          <p class="group-label">行业</p>
          <div class="block-grid">
            <div v-for="item in blocks.industry" :key="item.name" class="block-card">
              <span class="block-name">{{ item.name }}</span>
              <span :class="['block-pct', pctClass(item.change_pct)]">{{ fmtPct(item.change_pct) }}</span>
            </div>
          </div>
        </template>

        <template v-if="blocks.concept?.length">
          <p class="group-label">概念</p>
          <div class="block-grid">
            <div v-for="item in blocks.concept" :key="item.name" class="block-card">
              <span class="block-name">{{ item.name }}</span>
              <span :class="['block-pct', pctClass(item.change_pct)]">{{ fmtPct(item.change_pct) }}</span>
            </div>
          </div>
        </template>

        <template v-if="blocks.region?.length">
          <p class="group-label">地域</p>
          <div class="block-grid">
            <div v-for="item in blocks.region" :key="item.name" class="block-card">
              <span class="block-name">{{ item.name }}</span>
              <span :class="['block-pct', pctClass(item.change_pct)]">{{ fmtPct(item.change_pct) }}</span>
            </div>
          </div>
        </template>
      </section>
    </template>

    <!-- 近20日资金流向 -->
    <template v-if="fundFlow.length">
      <section class="section">
        <h3 class="section-title">近20日资金流向 <small>单位：万元</small></h3>
        <el-table :data="fundFlow" size="small" stripe>
          <el-table-column prop="date" label="日期" width="100" />
          <el-table-column prop="close" label="收盘价" width="80" align="right" />
          <el-table-column label="涨跌幅" width="80" align="right">
            <template #default="{ row }">
              <span :class="pctClass(row.change_pct)">{{ fmtPct(row.change_pct) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="主力净流入" align="right">
            <template #default="{ row }">
              <span :class="pctClass(row.main_in)">{{ fmtWan(row.main_in) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="超大单" align="right">
            <template #default="{ row }">
              <span :class="pctClass(row.super_net_in)">{{ fmtWan(row.super_net_in) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="大单" align="right">
            <template #default="{ row }">
              <span :class="pctClass(row.large_net_in)">{{ fmtWan(row.large_net_in) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="中单" align="right">
            <template #default="{ row }">
              <span :class="pctClass(row.medium_net_in)">{{ fmtWan(row.medium_net_in) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="小单" align="right">
            <template #default="{ row }">
              <span :class="pctClass(row.little_net_in)">{{ fmtWan(row.little_net_in) }}</span>
            </template>
          </el-table-column>
        </el-table>
      </section>
    </template>

    <el-empty v-if="!loading && !blocks && !fundFlow.length" description="暂无数据" />
  </div>
</template>

<style scoped lang="scss">
.blocks-fund-tab {
  padding: 12px 0;
}

.section {
  margin-bottom: 24px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin: 0 0 10px;

  small {
    font-size: 12px;
    font-weight: normal;
    color: var(--el-text-color-secondary);
    margin-left: 6px;
  }
}

.group-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin: 10px 0 6px;
}

.block-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
}

.block-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  font-size: 13px;
}

.block-name {
  color: var(--el-text-color-primary);
}

.block-pct {
  font-weight: 500;
}

.up {
  color: var(--el-color-danger);
}

.down {
  color: var(--el-color-success);
}
</style>
