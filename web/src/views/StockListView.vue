<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { wMessage } from '@/utils/message'
import { usePortfolioStore } from '@/stores/portfolio'
import { searchStocks } from '@/apis/stocks'
import type { Stock } from '@/types/api'

const router = useRouter()
const portfolioStore = usePortfolioStore()
const showAdd = ref(false)
const selectedStock = ref('')
const note = ref('')
const stockOptions = ref<{ value: string; label: string }[]>([])
const loadingAdd = ref(false)

onMounted(() => {
  portfolioStore.fetch()
})

async function searchStockOptions(query: string) {
  if (!query) {
    stockOptions.value = []
    return
  }
  try {
    const list = await searchStocks(query, 20)
    stockOptions.value = list.map((s: Stock) => ({
      value: s.tsCode,
      label: `${s.tsCode} ${s.name}`,
    }))
  } catch {
    stockOptions.value = []
  }
}

async function doAdd() {
  if (!selectedStock.value) {
    wMessage('warning', '请选择股票')
    return
  }
  loadingAdd.value = true
  try {
    await portfolioStore.add(selectedStock.value, note.value)
    wMessage('success', '添加成功')
    showAdd.value = false
    selectedStock.value = ''
    note.value = ''
  } finally {
    loadingAdd.value = false
  }
}

function goDetail(tsCode: string) {
  router.push(`/stocks/${tsCode}`)
}
</script>

<template>
  <div>
    <div class="header-bar">
      <h2>{{ $t('stockList.title') }}</h2>
      <el-button type="primary" @click="showAdd = true">{{ $t('stockList.addStock') }}</el-button>
    </div>

    <el-table :data="portfolioStore.items" style="margin-top: 16px">
      <el-table-column prop="tsCode" :label="$t('stockList.code')" />
      <el-table-column prop="name" :label="$t('stockList.name')" />
      <el-table-column prop="note" :label="$t('stockList.note')" />
      <el-table-column :label="$t('stockList.action')" width="140">
        <template #default="{ row }">
          <el-button link type="primary" @click="goDetail(row.tsCode)">{{ $t('stockList.detail') }}</el-button>
          <el-button link type="danger" @click="portfolioStore.remove(row.tsCode)">{{ $t('common.delete') }}</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showAdd" :title="$t('stockList.addStock')" width="400px">
      <el-form @submit.prevent="doAdd">
        <el-form-item :label="$t('stockList.code')">
          <el-select-v2
            v-model="selectedStock"
            :options="stockOptions"
            :placeholder="$t('stockList.selectStock')"
            clearable
            filterable
            remote
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
</style>
