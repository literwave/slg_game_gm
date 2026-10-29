<script setup lang="ts">
import { nextTick, onMounted, reactive, ref } from 'vue'
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
const saving = ref(false)

async function load() {
  loading.value = true
  try {
    applyLocal(await serverApi.list(true))
  } finally {
    loading.value = false
  }
}

type GmProtocol = 'http' | 'https'
type ServerForm = Omit<GameServerForm, 'gm_addr'> & {
  gm_protocol: GmProtocol
  gm_host: string
  gm_port: number
}

const protocolOptions: { value: GmProtocol; label: string }[] = [
  { value: 'http', label: 'http' },
  { value: 'https', label: 'https' },
]

const dialogVisible = ref(false)
const editingId = ref<string>()
const formRef = ref<FormInstance>()
const emptyForm = (): ServerForm => ({
  name: '',
  gm_protocol: 'http',
  gm_host: '',
  gm_port: 80,
  status: 1,
  open_at: toApiTime(),
  remark: '',
})
const form = reactive<ServerForm>(emptyForm())
const rules: FormRules = {
  name: [{ required: true, message: '请输入区服名称' }],
  gm_host: [{
    required: true,
    validator: (_r, value: string, cb) => {
      const host = (value ?? '').trim()
      if (!host) {
        cb(new Error('请输入 IP 或域名，如 10.0.0.1'))
        return
      }
      if (/^https?:\/\//i.test(host) || host.includes('/')) {
        cb(new Error('只需填写 IP 或域名，协议请用左侧下拉选择'))
        return
      }
      if (host.includes(':')) {
        cb(new Error('端口请填在右侧端口框'))
        return
      }
      try {
        const url = new URL(`${form.gm_protocol}://${host}`)
        if (!url.hostname) {
          cb(new Error('请输入合法的 IP 或域名'))
          return
        }
        cb()
      } catch {
        cb(new Error('请输入合法的 IP 或域名'))
      }
    },
  }],
  gm_port: [{ required: true, type: 'number', min: 1, max: 65535, message: '请输入 1–65535 的端口' }],
}

function parseGmAddr(raw: string) {
  try {
    const url = new URL(raw)
    const protocol: GmProtocol = url.protocol === 'https:' ? 'https' : 'http'
    const port = url.port ? Number(url.port) : protocol === 'https' ? 443 : 80
    return { protocol, host: url.hostname, port }
  } catch {
    return { protocol: 'http' as GmProtocol, host: raw, port: 80 }
  }
}

function buildGmAddr(protocol: GmProtocol, host: string, port: number) {
  const url = new URL(`${protocol}://${host.trim()}`)
  url.port = String(port)
  return url.origin
}

/** 列表里展示协议、主机和端口。 */
function gmHost(raw: string) {
  try {
    const url = new URL(raw)
    const port = url.port || (url.protocol === 'https:' ? '443' : '80')
    return `${url.protocol}//${url.hostname}:${port}`
  } catch {
    return raw
  }
}

function splitPastedAddr() {
  const raw = form.gm_host.trim()
  if (!raw) return
  try {
    const url = new URL(/^https?:\/\//i.test(raw) ? raw : `${form.gm_protocol}://${raw}`)
    form.gm_protocol = url.protocol === 'https:' ? 'https' : 'http'
    form.gm_host = url.hostname
    if (url.port) form.gm_port = Number(url.port)
  } catch {
    /* keep as typed */
  }
}

function onProtocolChange(protocol: GmProtocol) {
  if (protocol === 'https' && form.gm_port === 80) form.gm_port = 443
  if (protocol === 'http' && form.gm_port === 443) form.gm_port = 80
}

async function openCreate() {
  editingId.value = undefined
  Object.assign(form, emptyForm())
  dialogVisible.value = true
  await nextTick()
  formRef.value?.clearValidate()
}

async function openEdit(row: GameServer) {
  editingId.value = row.id
  const parsed = parseGmAddr(row.gm_addr)
  Object.assign(form, {
    name: row.name,
    gm_protocol: parsed.protocol,
    gm_host: parsed.host,
    gm_port: parsed.port,
    status: row.status,
    open_at: toApiTime(row.open_at),
    remark: row.remark,
  })
  dialogVisible.value = true
  await nextTick()
  formRef.value?.clearValidate()
}

function applyLocal(rows: GameServer[]) {
  list.value = rows
  serverStore.sync(rows)
}

async function submit() {
  await formRef.value?.validate()
  const payload: GameServerForm = {
    name: form.name,
    gm_addr: buildGmAddr(form.gm_protocol, form.gm_host, form.gm_port),
    status: form.status,
    open_at: form.open_at,
    remark: form.remark,
  }
  saving.value = true
  try {
    if (editingId.value) {
      await serverApi.update(editingId.value, payload)
      applyLocal(list.value.map((row) => (
        row.id === editingId.value ? { ...row, ...payload } : row
      )))
      ElMessage.success('保存成功')
    } else {
      const created = await serverApi.create(payload)
      applyLocal([...list.value, { ...created, online: 0, reachable: false, mock: list.value.some((s) => s.mock) }])
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
  } finally {
    saving.value = false
  }
}

async function remove(row: GameServer) {
  const label = row.server_id ? `${row.server_id} - ${row.name}` : row.name
  await ElMessageBox.confirm(`确定删除区服 ${label}？`, '提示', { type: 'warning' })
  await serverApi.remove(row.id)
  applyLocal(list.value.filter((s) => s.id !== row.id))
  ElMessage.success('已删除')
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

    <el-alert v-if="list.some((s) => s.mock)" class="page-card" type="warning" show-icon :closable="false"
      title="当前为模拟模式：连通和在线人数不会请求真实游戏服。" />

    <el-card class="table-card">
      <template #header>
        <div class="section-title">
          <el-icon><Monitor /></el-icon>区服列表<span class="list-count">{{ list.length }}</span>
        </div>
      </template>
      <el-table v-loading="loading" :data="list">
        <el-table-column label="服务器ID" min-width="120">
          <template #default="{ row }">
            <span v-if="row.server_id" class="code-tag">{{ row.server_id }}</span>
            <span v-else style="color: var(--gm-text-secondary)">-</span>
          </template>
        </el-table-column>
        <el-table-column label="服务器名" min-width="160" class-name="col-left" label-class-name="col-left" show-overflow-tooltip>
          <template #default="{ row }">{{ row.name }}</template>
        </el-table-column>
        <el-table-column label="服务器地址" min-width="200">
          <template #default="{ row }"><span class="code-tag">{{ gmHost(row.gm_addr) }}</span></template>
        </el-table-column>
        <el-table-column label="开服状态" min-width="120">
          <template #default="{ row }">
            <el-tag :type="serverStatusOf(row.status, row.open_at).tag" effect="light" round>
              {{ serverStatusOf(row.status, row.open_at).label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="连通" min-width="110">
          <template #default="{ row }">
            <span class="status-dot" :class="row.reachable ? 'success' : 'danger'">{{ row.reachable ? '可达' : '不可达' }}</span>
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
        <el-table-column prop="remark" label="备注" min-width="140" show-overflow-tooltip />
        <el-table-column v-if="canEdit" label="操作" min-width="180">
          <template #default="{ row }">
            <el-button link type="primary" icon="Edit" @click="openEdit(row)">编辑</el-button>
            <el-button link type="danger" icon="Delete" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑区服' : '新增区服'" width="520px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px" class="server-form">
        <el-form-item label="名称" prop="name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="服务器地址" required>
          <div class="gm-addr-row">
            <el-form-item>
              <el-select v-model="form.gm_protocol" @change="onProtocolChange">
                <el-option v-for="o in protocolOptions" :key="o.value" :value="o.value" :label="o.label" />
              </el-select>
            </el-form-item>
            <el-form-item prop="gm_host">
              <el-input v-model="form.gm_host" placeholder="10.0.0.1" @blur="splitPastedAddr" />
            </el-form-item>
            <span class="gm-addr-port-label">端口</span>
            <el-form-item prop="gm_port">
              <el-input-number v-model="form.gm_port" :min="1" :max="65535" :controls="false" placeholder="8081" />
            </el-form-item>
          </div>
        </el-form-item>
        <el-form-item label="开服状态">
          <el-select v-model="form.status" style="width: 100%">
            <el-option v-for="o in serverStatusOptions" :key="o.value" :value="o.value" :label="o.label" />
          </el-select>
        </el-form-item>
        <el-form-item label="开服时间">
          <el-date-picker v-model="form.open_at" type="datetime" :value-format="API_TIME_FORMAT" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submit">{{ editingId ? '保存' : '创建' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.gm-addr-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  width: 100%;
  min-width: 0;
}
.gm-addr-row > :deep(.el-form-item:first-child) {
  width: 108px;
  flex-shrink: 0;
  margin-bottom: 0;
}
.gm-addr-row > :deep(.el-form-item:nth-child(2)) {
  flex: 1;
  min-width: 0;
  margin-bottom: 0;
}
.gm-addr-row > :deep(.el-form-item:last-child) {
  width: 108px;
  flex-shrink: 0;
  margin-bottom: 0;
}
.gm-addr-row :deep(.el-form-item__content) {
  margin-left: 0 !important;
}
.gm-addr-row :deep(.el-input-number) {
  width: 100%;
}
.gm-addr-port-label {
  flex-shrink: 0;
  padding-top: 7px;
  color: var(--gm-text-secondary);
  font-size: 13px;
  line-height: 22px;
}
</style>

