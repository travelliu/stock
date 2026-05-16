<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getConceptBlocks, getFundFlow } from '@/apis/stocks'
import type { ConceptBlocks, FundFlow, OrderFlowGroups } from '@/types/api'
import { wMessage } from '@/utils/message'

const route = useRoute()
const code = route.params.code as string

const blocks = ref<ConceptBlocks | null>(null)
const fundFlow = ref<FundFlow | null>(null)
const loading = ref(false)

const activeLevel = ref(0)
const levelTabs = computed(() => fundFlow.value?.levels ?? [])
const stockFlow = computed(() => fundFlow.value?.stock_flow ?? null)

// ── helpers ───────────────────────────────────────────────────────────────────

function pctClass(val: string): string {
  const n = parseFloat(val)
  return isNaN(n) ? '' : n >= 0 ? 'up' : 'down'
}

function fmtPct(val: string): string {
  const n = parseFloat(val)
  if (isNaN(n)) return val
  return (n >= 0 ? '+' : '') + n.toFixed(2) + '%'
}

function fmtYi(val: string): string {
  const n = parseFloat(val)
  if (isNaN(n)) return val
  return (n >= 0 ? '+' : '') + n.toFixed(2)
}

function fmtAmt(val: string): string {
  const n = parseFloat(val)
  return isNaN(n) ? val : n.toFixed(2)
}

// ── 今日主力流向 net-flow bars (左上) ─────────────────────────────────────────

interface NetBar { label: string; value: number }

function netBars(src: OrderFlowGroups): NetBar[] {
  return [
    { label: '净特大单', value: parseFloat(src.super.net_turnover) },
    { label: '净大单',   value: parseFloat(src.large.net_turnover) },
    { label: '净中单',   value: parseFloat(src.medium.net_turnover) },
    { label: '净小单',   value: parseFloat(src.little.net_turnover) },
  ]
}

function netBarWidth(bars: NetBar[], value: number): number {
  const max = Math.max(...bars.map(b => Math.abs(b.value)), 0.01)
  return Math.round((Math.abs(value) / max) * 100)
}

// ── 资金分布 butterfly chart (右上) ──────────────────────────────────────────

interface DistRow {
  label: string
  inAmount: string; inRate: string; inWidth: number
  outAmount: string; outRate: string; outWidth: number
}

function distRows(src: OrderFlowGroups): DistRow[] {
  const cats = [
    { label: '特', g: src.super },
    { label: '大', g: src.large },
    { label: '中', g: src.medium },
    { label: '小', g: src.little },
  ]
  const maxRate = Math.max(
    ...cats.flatMap(c => [
      parseFloat(c.g.turnover_in_rate) || 0,
      parseFloat(c.g.turnover_out_rate) || 0,
    ]),
    0.01,
  )
  return cats.map(c => ({
    label:     c.label,
    inAmount:  c.g.turnover_in,
    inRate:    c.g.turnover_in_rate,
    inWidth:   Math.round(((parseFloat(c.g.turnover_in_rate) || 0) / maxRate) * 100),
    outAmount: c.g.turnover_out,
    outRate:   c.g.turnover_out_rate,
    outWidth:  Math.round(((parseFloat(c.g.turnover_out_rate) || 0) / maxRate) * 100),
  }))
}

onMounted(async () => {
  loading.value = true
  try {
    const [b, f] = await Promise.all([getConceptBlocks(code), getFundFlow(code)])
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

    <!-- 上半区：2 列 -->
    <div class="top-grid">

      <!-- 左上：今日主力流向 -->
      <div class="panel">
        <h3 class="panel-title">
          今日主力流向
          <small v-if="stockFlow">单位: {{ stockFlow.unit }}</small>
        </h3>
        <template v-if="stockFlow">
          <div class="main-summary">
            <div class="summary-item">
              <span class="summary-label">主力流入</span>
              <span class="up">{{ fmtAmt(stockFlow.today_main.main_in) }}</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">主力流出</span>
              <span class="down">{{ fmtAmt(stockFlow.today_main.main_out) }}</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">主力净流入</span>
              <span :class="pctClass(stockFlow.today_main.main_net_in)">
                {{ fmtYi(stockFlow.today_main.main_net_in) }}
              </span>
            </div>
          </div>
          <div class="net-bars">
            <div class="net-bar-header">
              <span>类别</span>
              <span class="spacer" />
              <span>净流入</span>
            </div>
            <div
              v-for="bar in netBars(stockFlow)"
              :key="bar.label"
              class="net-bar-row"
            >
              <span class="net-label">{{ bar.label }}</span>
              <div class="net-track">
                <div
                  class="net-fill"
                  :class="bar.value >= 0 ? 'bar-up' : 'bar-down'"
                  :style="{ width: netBarWidth(netBars(stockFlow!), bar.value) + '%' }"
                />
              </div>
              <span :class="['net-value', pctClass(String(bar.value))]">
                {{ fmtYi(String(bar.value)) }}
              </span>
            </div>
          </div>
        </template>
        <el-empty v-else description="暂无数据" :image-size="40" />
      </div>

      <!-- 右上：资金分布 -->
      <div class="panel">
        <h3 class="panel-title">资金分布</h3>
        <template v-if="stockFlow">
          <div class="dist-header">
            <span>流入 <b class="up">{{ fmtAmt(stockFlow.turnover_in_total) }}</b></span>
            <span class="spacer" />
            <span><b class="down">{{ fmtAmt(stockFlow.turnover_out_total) }}</b> 流出</span>
          </div>
          <div
            v-for="row in distRows(stockFlow)"
            :key="row.label"
            class="dist-row"
          >
            <div class="dist-left">
              <span class="d-label">{{ row.label }}</span>
              <span class="d-amount">{{ fmtAmt(row.inAmount) }}</span>
              <span class="d-rate">{{ row.inRate }}</span>
            </div>
            <div class="dist-center">
              <div class="in-half">
                <div class="in-bar" :style="{ width: row.inWidth + '%' }" />
              </div>
              <div class="out-half">
                <div class="out-bar" :style="{ width: row.outWidth + '%' }" />
              </div>
            </div>
            <div class="dist-right">
              <span class="d-rate">{{ row.outRate }}</span>
              <span class="d-amount">{{ fmtAmt(row.outAmount) }}</span>
              <span class="d-label">{{ row.label }}</span>
            </div>
          </div>
        </template>
        <el-empty v-else description="暂无数据" :image-size="40" />
      </div>

    </div>

    <!-- 下半区：2 列 -->
    <div class="bottom-grid">

      <!-- 左下：所属板块 -->
      <div class="panel">
        <h3 class="panel-title">所属板块</h3>
        <template v-if="blocks">
          <template
            v-for="(group, label) in { 行业: blocks.industry, 概念: blocks.concept, 地域: blocks.region }"
            :key="label"
          >
            <template v-if="group?.length">
              <p class="group-label">{{ label }}</p>
              <div class="block-list">
                <div v-for="item in group" :key="item.name" class="block-row">
                  <span class="block-name">{{ item.name }}</span>
                  <span v-if="item.describe" class="block-desc">{{ item.describe }}</span>
                  <span :class="['block-pct', pctClass(item.change_pct)]">{{ fmtPct(item.change_pct) }}</span>
                </div>
              </div>
            </template>
          </template>
        </template>
        <el-empty v-else description="暂无数据" :image-size="40" />
      </div>

      <!-- 右下：所属行业资金流向 (2 tabs) -->
      <div class="panel">
        <h3 class="panel-title">所属行业资金流向</h3>
        <template v-if="levelTabs.length">
          <el-tabs v-model="activeLevel" type="card" size="small">
            <el-tab-pane
              v-for="(level, idx) in levelTabs"
              :key="level.belongs"
              :label="level.industry.name + '（' + level.industry.desc + '）'"
              :name="idx"
            >
              <div class="level-meta">单位：{{ level.unit }} &nbsp;·&nbsp; {{ level.update_time }}</div>
              <el-table
                :data="[
                  { label: '超大单', ...level.super },
                  { label: '大单',   ...level.large },
                  { label: '中单',   ...level.medium },
                  { label: '小单',   ...level.little },
                ]"
                size="small"
                stripe
              >
                <el-table-column prop="label" label="类型" width="60" />
                <el-table-column label="净流入" align="right" min-width="76">
                  <template #default="{ row }">
                    <span :class="pctClass(row.net_turnover)">{{ fmtYi(row.net_turnover) }}</span>
                  </template>
                </el-table-column>
                <el-table-column label="流入" align="right" min-width="76">
                  <template #default="{ row }">{{ fmtAmt(row.turnover_in) }}</template>
                </el-table-column>
                <el-table-column label="流出" align="right" min-width="76">
                  <template #default="{ row }">{{ fmtAmt(row.turnover_out) }}</template>
                </el-table-column>
                <el-table-column prop="turnover_in_rate" label="流入占比" align="right" width="68" />
                <el-table-column prop="turnover_out_rate" label="流出占比" align="right" width="68" />
              </el-table>
              <div v-if="level.recently?.length" class="recent-row" style="margin-top: 8px">
                <div v-for="agg in level.recently" :key="agg.key" class="recent-cell">
                  <span class="recent-label">{{ agg.key }}</span>
                  <span :class="['recent-value', pctClass(agg.value)]">{{ fmtYi(agg.value) }}</span>
                </div>
              </div>
            </el-tab-pane>
          </el-tabs>
        </template>
        <el-empty v-else description="暂无数据" :image-size="40" />
      </div>

    </div>

  </div>
</template>

<style scoped lang="scss">
.blocks-fund-tab {
  padding: 12px 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

// ── 网格布局 ──────────────────────────────────────────────────────────────────

.top-grid,
.bottom-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  align-items: start;
}

// ── 通用面板 ──────────────────────────────────────────────────────────────────

.panel {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  padding: 12px;
  background: #fff;
}

.panel-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin: 0 0 10px;

  small {
    font-size: 11px;
    font-weight: normal;
    color: var(--el-text-color-secondary);
    margin-left: 6px;
  }
}

.spacer { flex: 1; }

// ── 今日主力流向 (左上) ───────────────────────────────────────────────────────

.main-summary {
  display: flex;
  gap: 6px;
  margin-bottom: 10px;
}

.summary-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 6px 4px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  background: var(--el-fill-color-lighter);
  font-size: 13px;
  font-weight: 600;
}

.summary-label {
  font-size: 11px;
  font-weight: normal;
  color: var(--el-text-color-secondary);
  white-space: nowrap;
}

.net-bars {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.net-bar-header {
  display: flex;
  align-items: center;
  font-size: 11px;
  color: var(--el-text-color-secondary);
  padding-bottom: 4px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.net-bar-row {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
}

.net-label {
  width: 52px;
  flex-shrink: 0;
  color: var(--el-text-color-primary);
}

.net-track {
  flex: 1;
  height: 13px;
  background: var(--el-fill-color-light);
  border-radius: 2px;
  overflow: hidden;
}

.net-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 0.4s ease;
}

.bar-up   { background: var(--el-color-danger); }
.bar-down { background: var(--el-color-success); }

.net-value {
  width: 48px;
  text-align: right;
  font-weight: 500;
  flex-shrink: 0;
}

// ── 资金分布 butterfly (右上) ─────────────────────────────────────────────────

.dist-header {
  display: flex;
  align-items: center;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 8px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--el-border-color-lighter);

  b { font-weight: 600; font-size: 13px; }
}

.dist-row {
  display: flex;
  align-items: center;
  height: 26px;
  border-bottom: 1px solid var(--el-fill-color);
  font-size: 11px;

  &:last-child { border-bottom: none; }
}

.dist-left,
.dist-right {
  display: flex;
  align-items: center;
  gap: 4px;
  width: 108px;
  flex-shrink: 0;
}

.dist-left  { justify-content: flex-end; }
.dist-right { justify-content: flex-start; }

.d-label  { color: var(--el-text-color-secondary); width: 14px; text-align: center; }
.d-amount { color: var(--el-text-color-primary); min-width: 32px; text-align: right; }
.d-rate   { color: var(--el-text-color-secondary); min-width: 42px; }

.dist-left .d-rate   { text-align: right; }
.dist-right .d-rate  { text-align: left; }
.dist-right .d-amount { text-align: left; }

.dist-center {
  flex: 1;
  display: flex;
  height: 14px;
  margin: 0 4px;
  gap: 1px;
}

.in-half {
  flex: 1;
  display: flex;
  justify-content: flex-end;
  background: var(--el-fill-color-lighter);
  border-radius: 2px 0 0 2px;
  overflow: hidden;
}

.in-bar {
  height: 100%;
  background: var(--el-color-danger);
  transition: width 0.4s ease;
}

.out-half {
  flex: 1;
  display: flex;
  justify-content: flex-start;
  background: var(--el-fill-color-lighter);
  border-radius: 0 2px 2px 0;
  overflow: hidden;
}

.out-bar {
  height: 100%;
  background: var(--el-color-success);
  transition: width 0.4s ease;
}

// ── 所属板块 ──────────────────────────────────────────────────────────────────

.group-label {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  margin: 8px 0 2px;
}

.block-list {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  overflow: hidden;
}

.block-row {
  display: flex;
  align-items: center;
  padding: 5px 8px;
  font-size: 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);

  &:last-child { border-bottom: none; }
}

.block-name {
  flex: 1;
  color: var(--el-text-color-primary);
}

.block-desc {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  margin-right: 8px;
}

.block-pct {
  font-size: 12px;
  font-weight: 500;
  min-width: 50px;
  text-align: right;
}

// ── 所属行业 tab ──────────────────────────────────────────────────────────────

.level-meta {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  margin-bottom: 6px;
}

// ── 近期净流入 ────────────────────────────────────────────────────────────────

.recent-row {
  display: flex;
  gap: 6px;
}

.recent-cell {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 6px 2px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  background: var(--el-fill-color-lighter);
}

.recent-label {
  font-size: 10px;
  color: var(--el-text-color-secondary);
  margin-bottom: 2px;
  white-space: nowrap;
}

.recent-value {
  font-size: 12px;
  font-weight: 500;
}

// ── 颜色 ──────────────────────────────────────────────────────────────────────

.up   { color: var(--el-color-danger); }
.down { color: var(--el-color-success); }
</style>
