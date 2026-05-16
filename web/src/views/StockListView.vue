<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { wMessage } from '@/utils/message'
import { usePortfolioStore } from '@/stores/portfolio'
import { searchStocks } from '@/apis/stocks'
import type { Stock, Portfolio } from '@/types/api'
import { fmtQuoteTime } from '@/utils/format'

const router = useRouter()
const portfolioStore = usePortfolioStore()
const showAdd = ref(false)
const selectedCode = ref('')
const note = ref('')
const stockOptions = ref<{ value: string; label: string }[]>([])
const loadingAdd = ref(false)

onMounted(() => portfolioStore.fetch())

async function searchStockOptions(query: string) {
  if (!query) { stockOptions.value = []; return }
  try {
    const list = await searchStocks(query, 20)
    stockOptions.value = list.map((s: Stock) => ({ value: s.code, label: `${s.code} ${s.name}` }))
  } catch {
    stockOptions.value = []
  }
}

async function doAdd() {
  if (!selectedCode.value) { wMessage('warning', '请选择股票'); return }
  loadingAdd.value = true
  try {
    await portfolioStore.add(selectedCode.value, note.value)
    wMessage('success', '添加成功')
    showAdd.value = false
    selectedCode.value = ''
    note.value = ''
  } finally {
    loadingAdd.value = false
  }
}

function goDetail(row: Portfolio) {
  router.push(`/stocks/${row.code}`)
}

function removeItem(row: Portfolio) {
  portfolioStore.remove(row.code)
}

function fmtPrice(v: number | undefined): string {
  return v ? v.toFixed(2) : '-'
}

function pctClass(v: number): string {
  if (v > 0) return 'up'
  if (v < 0) return 'down'
  return ''
}

function fmtPct(v: number | undefined): string {
  if (v == null) return '-'
  return (v > 0 ? '+' : '') + v.toFixed(2) + '%'
}
</script>

<template>
  <div>
    <div class="header-bar">
      <h2>{{ $t('stockList.title') }}</h2>
      <el-button type="primary" @click="showAdd = true">{{ $t('stockList.addStock') }}</el-button>
    </div>

    <el-table :data="portfolioStore.items" style="margin-top: 16px">
      <el-table-column prop="code" :label="$t('stockList.code')" width="100">
        <template #default="{ row }">
          <el-button link type="primary" @click="goDetail(row)">{{ row.code }}</el-button>
        </template>
      </el-table-column>
      <el-table-column prop="name" :label="$t('stockList.name')" width="120" />
      <el-table-column label="时间" width="60" align="center">
        <template #default="{ row }">
          <span class="secondary">{{ fmtQuoteTime(row.quote?.quoteTime) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="开盘" align="right" width="80">
        <template #default="{ row }">{{ fmtPrice(row.quote?.open) }}</template>
      </el-table-column>
      <el-table-column label="最高" align="right" width="80">
        <template #default="{ row }">{{ fmtPrice(row.quote?.high) }}</template>
      </el-table-column>
      <el-table-column label="最低" align="right" width="80">
        <template #default="{ row }">{{ fmtPrice(row.quote?.low) }}</template>
      </el-table-column>
      <el-table-column label="最新" align="right" width="80">
        <template #default="{ row }">
          <span :class="pctClass(row.quote?.changePct ?? 0)">
            {{ fmtPrice(row.quote?.price) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="涨跌幅" align="right" width="90">
        <template #default="{ row }">
          <span :class="pctClass(row.quote?.changePct ?? 0)">
            {{ fmtPct(row.quote?.changePct) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column prop="note" :label="$t('stockList.note')" />
      <el-table-column :label="$t('stockList.action')" width="120">
        <template #default="{ row }">
          <el-button link type="primary" @click="goDetail(row)">{{ $t('stockList.detail') }}</el-button>
          <el-button link type="danger" @click="removeItem(row)">{{ $t('common.delete') }}</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showAdd" :title="$t('stockList.addStock')" width="400px">
      <el-form @submit.prevent="doAdd">
        <el-form-item :label="$t('stockList.code')">
          <el-select-v2
            v-model="selectedCode"
            :options="stockOptions"
            :placeholder="$t('stockList.selectStock')"
            clearable filterable remote
            :remote-method="searchStockOptions"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="$t('stockList.note')">
          <el-input v-model="note" :placeholder="$t('common.empty')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAdd = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="loadingAdd" @click="doAdd">{{ $t('common.add') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.header-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.up   { color: var(--el-color-danger); }
.down { color: var(--el-color-success); }
.secondary { color: var(--el-text-color-secondary); font-size: 12px; }
</style>
