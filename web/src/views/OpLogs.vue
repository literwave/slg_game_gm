<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { systemApi } from '@/api'
import PageHeader from '@/components/PageHeader.vue'
import { useTable } from '@/composables/useTable'
import { formatTime } from '@/utils/format'

const filters = reactive({
  username: '',
  method: '',
  path: '',
  start_at: undefined as number | undefined,
  end_at: undefined as number | undefined,
})
const timeRange = ref<[Date, Date] | null>(null)
watch(timeRange, (v) => {
  filters.start_at = v?.[0].getTime()
  filters.end_at = v?.[1].getTime()
})

const { list, total, page, pageSize, loading, load, search } = useTable(systemApi.opLogs, filters)

const methodTag = (m: string) => ({ POST: 'success', PUT: 'warning', DELETE: 'danger' } as const)[m] ?? 'info'

onMounted(load)
</script>

<template>
  <div>
    <PageHeader />

    <el-card class="filter-card">
      <div class="filter-bar">
        <div class="filter-item">
          <label>操作人</label>
          <el-input v-model="filters.username" placeholder="用户名" clearable style="width: 150px" />
        </div>
        <div class="filter-item">
          <label>方法</label>
          <el-select v-model="filters.method" placeholder="全部" clearable style="width: 110px">
            <el-option v-for="m in ['POST', 'PUT', 'DELETE']" :key="m" :value="m" :label="m" />
          </el-select>
        </div>
        <div class="filter-item">
          <label>路径</label>
          <el-input v-model="filters.path" placeholder="如 /players" clearable style="width: 200px" />
        </div>
        <div class="filter-item">
          <label>时间</label>
          <el-date-picker v-model="timeRange" type="datetimerange" start-placeholder="开始时间" end-placeholder="结束时间"
            style="width: 340px" />
        </div>
        <el-button type="primary" @click="search">查询</el-button>
      </div>
    </el-card>

    <el-card class="table-card">
      <template #header>
        <div class="section-title">
          <el-icon><Document /></el-icon>操作记录<span class="list-count">{{ total }}</span>
          <span class="extra"><el-button size="small" icon="Refresh" circle @click="load" /></span>
        </div>
      </template>
      <el-table v-loading="loading" :data="list">
        <el-table-column type="expand">
          <template #default="{ row }">
            <div class="log-detail">
              <div v-if="row.query" class="kv-item" style="margin-bottom: 8px">
                <span class="k">Query：</span><span class="code-tag">{{ row.query }}</span>
              </div>
              <pre class="log-body">{{ row.body || '(无请求体)' }}</pre>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="时间" min-width="180">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作人" min-width="140">
          <template #default="{ row }">
            <div class="cell-user">
              <el-avatar :size="26">{{ (row.username || '?').slice(0, 1) }}</el-avatar>
              <span>{{ row.username || '-' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="请求" min-width="320" class-name="col-left" label-class-name="col-left">
          <template #default="{ row }">
            <el-tag :type="methodTag(row.method)" effect="light" size="small" style="width: 58px; margin-right: 8px">{{ row.method }}</el-tag>
            <span class="code-tag" style="border: none; background: none; padding: 0">{{ row.path }}</span>
          </template>
        </el-table-column>
        <el-table-column label="结果" min-width="100">
          <template #default="{ row }">
            <span class="status-dot" :class="row.status < 400 ? 'success' : 'danger'">{{ row.status }}</span>
          </template>
        </el-table-column>
        <el-table-column label="耗时" min-width="90">
          <template #default="{ row }">{{ row.latency_ms }} ms</template>
        </el-table-column>
        <el-table-column label="IP" min-width="140">
          <template #default="{ row }"><span class="code-tag">{{ row.ip }}</span></template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" class="pager" :total="total"
        layout="total, sizes, prev, pager, next" @change="load" />
    </el-card>
  </div>
</template>

<style scoped>
.log-detail {
  padding: 4px 48px 12px;
}
@media (max-width: 768px) {
  .log-detail {
    padding: 4px 8px 12px;
  }
}
.log-body {
  margin: 0;
  padding: 10px 12px;
  border-radius: 6px;
  background: #f7f8fa;
  border: 1px solid var(--gm-border);
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>

