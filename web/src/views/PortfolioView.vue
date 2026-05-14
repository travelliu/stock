<template>
  <div>
    <h2>我的持仓</h2>
    <el-button type="primary" @click="showAdd = true">添加股票</el-button>
    <el-table :data="portfolioStore.items" style="margin-top: 16px;">
      <el-table-column prop="tsCode" label="代码" />
      <el-table-column prop="note" label="备注" />
      <el-table-column label="操作">
        <template #default="{ row }">
          <el-button link type="primary" @click="$router.push(`/stock/${row.tsCode}`)">详情</el-button>
          <el-button link type="danger" @click="portfolioStore.remove(row.tsCode)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showAdd" title="添加股票" width="400px">
      <el-form @submit.prevent="doAdd">
        <el-form-item label="股票">
          <el-select-v2
            v-model="selectedStock"
            :options="stockOptions"
            placeholder="输入代码或名称搜索"
            clearable
            filterable
            remote
            :remote-method="searchStocks"
            style="width: 100%;"
          />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="note" placeholder="可选" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAdd = false">取消</el-button>
        <el-button type="primary" @click="doAdd">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { usePortfolioStore } from '@/stores/portfolio'
import client from '@/api/client'

const portfolioStore = usePortfolioStore()
const showAdd = ref(false)
const selectedStock = ref('')
const note = ref('')
const stockOptions = ref<{ value: string; label: string }[]>([])

onMounted(() => {
  portfolioStore.fetch()
})

async function searchStocks(query: string) {
  if (!query) return
  const { data } = await client.get('/stocks', { params: { q: query, limit: 20 } })
  stockOptions.value = data.map((s: any) => ({
    value: s.tsCode,
    label: `${s.tsCode} ${s.name}`,
  }))
}

async function doAdd() {
  if (!selectedStock.value) {
    ElMessage.warning('请选择股票')
    return
  }
  await portfolioStore.add(selectedStock.value, note.value)
  showAdd.value = false
  selectedStock.value = ''
  note.value = ''
  ElMessage.success('添加成功')
}
</script>
