<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { serverApi } from '@/api'
import PageHeader from '@/components/PageHeader.vue'
import type { GameServer, GameServerForm } from '@/api/types'
import { useUserStore } from '@/stores/user'
import { useServerStore } from '@/stores/servers'
import { API_TIME_FORMAT, formatNumber, formatTime, serverStatusOf, serverStatusOptions, toApiTime } from '@/utils/format'

const store = useUserStore()
const serverStore = useServerStore()
const canEdit = store.hasRole('admin')

const list = ref<GameServer[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    list.value = await serverApi.list(true)
  } finally {
    loading.value = false
  }
}

const dialogVisible = ref(false)
const editingId = ref<string>()
const formRef = ref<FormInstance>()
const emptyForm = (): GameServerForm => ({
  server_id: 0,
  name: '',
  gm_addr: 'http://',
  status: 1,
  open_at: toApiTime(),
  remark: '',
})
const form = reactive<GameServerForm>(emptyForm())
const rules: FormRules = {
  server_id: [{ required: true, type: 'number', min: 1, message: '请输入区服号' }],
  name: [{ required: true, message: '请输入区服名称' }],
  gm_addr: [{ required: true, type: 'url', message: '请输入合法的 GM 地址，如 http://10.0.0.1:9001' }],
}

function openCreate() {
  editingId.value = undefined
  Object.assign(form, emptyForm())
  dialogVisible.value = true
}

function openEdit(row: GameServer) {
  editingId.value = row.id
  Object.assign(form, {
    server_id: row.server_id,
    name: row.name,
    gm_addr: row.gm_addr,
    status: row.status,
    open_at: toApiTime(row.open_at),
    remark: row.remark,
  })
  dialogVisible.value = true
}

async function submit() {
  await formRef.value?.validate()
  if (editingId.value) {
    await serverApi.update(editingId.value, form)
  } else {
    await serverApi.create(form)
  }
  ElMessage.success('保存成功')
  dialogVisible.value = false
  await Promise.all([load(), serverStore.load(true)])
}

async function remove(row: GameServer) {
  await ElMessageBox.confirm(`确定删除区服 ${row.server_id} - ${row.name}？`, '提示', { type: 'warning' })
  await serverApi.remove(row.id)
  ElMessage.success('已删除')
  await Promise.all([load(), serverStore.load(true)])
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader>
      <template #actions>
        <el-button icon="Refresh" @click="load">刷新状态</el-button>
        <el-button v-if="canEdit" type="primary" icon="Plus" @click="openCreate">新增区服</el-button>
      </template>
    </PageHeader>

    <el-card class="table-card">
      <template #header>
        <div class="section-title">
          <el-icon><Monitor /></el-icon>区服列表<span class="list-count">{{ list.length }}</span>
        </div>
      </template>
      <el-table v-loading="loading" :data="list">
        <el-table-column label="区服" min-width="180" class-name="col-left" label-class-name="col-left">
          <template #default="{ row }">
            <div class="cell-user">
              <el-avatar :size="36" shape="square">{{ String(row.server_id).slice(-2) }}</el-avatar>
              <div>
                <div class="main">{{ row.name }}</div>
                <div class="sub">区服号 {{ row.server_id }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="GM 地址" min-width="200">
          <template #default="{ row }"><span class="code-tag">{{ row.gm_addr }}</span></template>
        </el-table-column>
        <el-table-column label="状态" min-width="100">
          <template #default="{ row }">
            <el-tag :type="serverStatusOf(row.status).tag" effect="light" round>{{ serverStatusOf(row.status).label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="连接" min-width="100">
          <template #default="{ row }">
            <span class="status-dot" :class="row.reachable ? 'success' : 'danger'">{{ row.reachable ? '正常' : '不可达' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="在线人数" min-width="110">
          <template #default="{ row }">
            <b v-if="row.reachable">{{ formatNumber(row.online) }}</b>
            <span v-else style="color: var(--gm-text-secondary)">-</span>
          </template>
        </el-table-column>
        <el-table-column label="开服时间" min-width="180">
          <template #default="{ row }">{{ formatTime(row.open_at) }}</template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />
        <el-table-column v-if="canEdit" label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" icon="Edit" @click="openEdit(row)">编辑</el-button>
            <el-button link type="danger" icon="Delete" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑区服' : '新增区服'" width="520px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="区服号" prop="server_id"><el-input-number v-model="form.server_id" :min="1" /></el-form-item>
        <el-form-item label="名称" prop="name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="GM 地址" prop="gm_addr"><el-input v-model="form.gm_addr" placeholder="http://10.0.0.1:9001" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status">
            <el-option v-for="o in serverStatusOptions" :key="o.value" :value="o.value" :label="o.label" />
          </el-select>
        </el-form-item>
        <el-form-item label="开服时间">
          <el-date-picker v-model="form.open_at" type="datetime" :value-format="API_TIME_FORMAT" />
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

