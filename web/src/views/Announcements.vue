<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import dayjs from 'dayjs'
import { announcementApi } from '@/api'
import type { Announcement, AnnouncementForm } from '@/api/types'
import PageHeader from '@/components/PageHeader.vue'
import { useTable } from '@/composables/useTable'
import { useUserStore } from '@/stores/user'
import { useServerStore } from '@/stores/servers'
import { API_TIME_FORMAT, formatTime, toApiTime } from '@/utils/format'

const store = useUserStore()
const serverStore = useServerStore()
const canEdit = store.hasRole('operator')

const filters = reactive({ status: '' })
const { list, total, page, pageSize, loading, load, search } = useTable(announcementApi.list, filters)

const dialogVisible = ref(false)
const editingId = ref<string>()
const formRef = ref<FormInstance>()
const timeRange = ref<[string, string]>()
const emptyForm = (): AnnouncementForm => ({
  server_ids: [],
  title: '',
  content: '',
  start_at: '',
  end_at: '',
  interval_minutes: 30,
})
const form = reactive<AnnouncementForm>(emptyForm())
const rules: FormRules = {
  title: [{ required: true, message: '请输入标题' }],
  content: [{ required: true, message: '请输入内容' }],
}

function openCreate() {
  editingId.value = undefined
  Object.assign(form, emptyForm())
  timeRange.value = [toApiTime(), toApiTime(dayjs().add(1, 'day').toDate())]
  dialogVisible.value = true
}

function openEdit(row: Announcement) {
  editingId.value = row.id
  Object.assign(form, {
    server_ids: [...row.server_ids],
    title: row.title,
    content: row.content,
    interval_minutes: row.interval_minutes,
  })
  timeRange.value = [toApiTime(row.start_at), toApiTime(row.end_at)]
  dialogVisible.value = true
}

async function submit() {
  await formRef.value?.validate()
  if (!timeRange.value) return ElMessage.warning('请选择生效时间')
  ;[form.start_at, form.end_at] = timeRange.value
  if (editingId.value) {
    await announcementApi.update(editingId.value, form)
  } else {
    await announcementApi.create(form)
  }
  ElMessage.success('保存成功')
  dialogVisible.value = false
  await load()
}

async function publish(row: Announcement) {
  await ElMessageBox.confirm(`确定发布公告「${row.title}」？`, '提示')
  const res = await announcementApi.publish(row.id)
  const failed = res.results?.filter((r) => !r.success).length ?? 0
  failed ? ElMessage.warning(`已发布，${failed} 个区服失败`) : ElMessage.success('发布成功')
  await load()
}

async function remove(row: Announcement) {
  await ElMessageBox.confirm(`确定删除公告「${row.title}」？`, '提示', { type: 'warning' })
  await announcementApi.remove(row.id)
  ElMessage.success('已删除')
  await load()
}

onMounted(() => {
  serverStore.load()
  load()
})
</script>

<template>
  <div>
    <PageHeader>
      <template #actions>
        <el-button v-if="canEdit" type="primary" icon="Plus" @click="openCreate">新建公告</el-button>
      </template>
    </PageHeader>

    <el-card class="filter-card">
      <div class="filter-bar">
        <div class="filter-item">
          <label>状态</label>
          <el-radio-group v-model="filters.status" @change="search">
            <el-radio-button value="">全部</el-radio-button>
            <el-radio-button value="draft">草稿</el-radio-button>
            <el-radio-button value="published">已发布</el-radio-button>
          </el-radio-group>
        </div>
      </div>
    </el-card>

    <el-card class="table-card">
      <template #header>
        <div class="section-title">
          <el-icon><Bell /></el-icon>公告列表<span class="list-count">{{ total }}</span>
          <span class="extra"><el-button size="small" icon="Refresh" circle @click="load" /></span>
        </div>
      </template>
      <el-table v-loading="loading" :data="list">
        <el-table-column label="公告" min-width="280" class-name="col-left" label-class-name="col-left">
          <template #default="{ row }">
            <div class="main" style="font-weight: 600">{{ row.title }}</div>
            <div class="ann-content">{{ row.content }}</div>
          </template>
        </el-table-column>
        <el-table-column label="区服" min-width="140">
          <template #default="{ row }">
            <span class="code-tag">{{ row.server_ids.length ? row.server_ids.join(', ') : '全部' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="生效时间" min-width="200">
          <template #default="{ row }">
            <div>{{ formatTime(row.start_at) }}</div>
            <div style="color: var(--gm-text-secondary)">至 {{ formatTime(row.end_at) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="滚动间隔" min-width="100">
          <template #default="{ row }">{{ row.interval_minutes ? `${row.interval_minutes} 分钟` : '仅一次' }}</template>
        </el-table-column>
        <el-table-column label="状态" min-width="100">
          <template #default="{ row }">
            <span class="status-dot" :class="row.status === 'published' ? 'success' : 'info'">
              {{ row.status === 'published' ? '已发布' : '草稿' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="operator" label="创建人" min-width="100" />
        <el-table-column v-if="canEdit" label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" icon="Promotion" @click="publish(row)">{{ row.status === 'published' ? '重新推送' : '发布' }}</el-button>
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" class="pager" :total="total"
        layout="total, sizes, prev, pager, next" @change="load" />
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑公告' : '新建公告'" width="600px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="区服">
          <el-select v-model="form.server_ids" multiple placeholder="不选则为全部区服" style="width: 100%">
            <el-option v-for="o in serverStore.options" :key="o.value" :value="o.value" :label="o.label" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题" prop="title"><el-input v-model="form.title" maxlength="64" /></el-form-item>
        <el-form-item label="内容" prop="content">
          <el-input v-model="form.content" type="textarea" :rows="4" maxlength="1000" show-word-limit />
        </el-form-item>
        <el-form-item label="生效时间">
          <el-date-picker v-model="timeRange" type="datetimerange" :value-format="API_TIME_FORMAT"
            start-placeholder="开始" end-placeholder="结束" />
        </el-form-item>
        <el-form-item label="滚动间隔">
          <el-input-number v-model="form.interval_minutes" :min="0" />
          <span style="margin-left: 8px">分钟（0 为只播一次）</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.ann-content {
  margin-top: 4px;
  font-size: 12px;
  color: var(--gm-text-secondary);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>

