<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { playerApi, systemApi } from '@/api'
import type { Player } from '@/api/types'
import PageHeader from '@/components/PageHeader.vue'
import { useUserStore } from '@/stores/user'
import { useServerStore } from '@/stores/servers'
import { formatNumber, formatTime, resourceLabels } from '@/utils/format'

const store = useUserStore()
const serverStore = useServerStore()
const canOperate = store.hasRole('operator')

const query = reactive({ server_id: undefined as number | undefined, keyword: '' })
const list = ref<Player[]>([])
const loading = ref(false)
const resourceTypes = ref<string[]>([])

async function search() {
  if (!query.server_id) {
    ElMessage.warning('请先选择区服')
    return
  }
  loading.value = true
  try {
    list.value = await playerApi.search(query.server_id, query.keyword)
  } finally {
    loading.value = false
  }
}

// 详情
const detailVisible = ref(false)
const detail = ref<Player>()
async function openDetail(row: Player) {
  detail.value = await playerApi.get(row.server_id, row.player_id)
  detailVisible.value = true
}

async function refreshRow(row: Player) {
  const fresh = await playerApi.get(row.server_id, row.player_id)
  const idx = list.value.findIndex((p) => p.player_id === row.player_id)
  if (idx >= 0) list.value[idx] = fresh
  if (detail.value?.player_id === row.player_id) detail.value = fresh
}

// 封禁
const banVisible = ref(false)
const banTarget = ref<Player>()
const banForm = reactive({ duration_hours: 24, reason: '' })
function openBan(row: Player) {
  banTarget.value = row
  Object.assign(banForm, { duration_hours: 24, reason: '' })
  banVisible.value = true
}
async function submitBan() {
  const p = banTarget.value!
  if (!banForm.reason) return ElMessage.warning('请填写封禁原因')
  await playerApi.ban(p.server_id, p.player_id, banForm.duration_hours, banForm.reason)
  ElMessage.success('已封禁')
  banVisible.value = false
  await refreshRow(p)
}

async function unban(row: Player) {
  await ElMessageBox.confirm(`确定解封玩家 ${row.name}？`, '提示')
  await playerApi.unban(row.server_id, row.player_id)
  ElMessage.success('已解封')
  await refreshRow(row)
}

async function kick(row: Player) {
  await ElMessageBox.confirm(`确定将玩家 ${row.name} 踢下线？`, '提示', { type: 'warning' })
  await playerApi.kick(row.server_id, row.player_id)
  ElMessage.success('已踢下线')
  await refreshRow(row)
}

// 修改资源
const resVisible = ref(false)
const resTarget = ref<Player>()
const resChanges = reactive<Record<string, number>>({})
function openResources(row: Player) {
  resTarget.value = row
  resourceTypes.value.forEach((k) => (resChanges[k] = 0))
  resVisible.value = true
}
async function submitResources() {
  const p = resTarget.value!
  const changes = Object.fromEntries(Object.entries(resChanges).filter(([, v]) => v !== 0))
  if (Object.keys(changes).length === 0) return ElMessage.warning('未修改任何资源')
  await playerApi.modifyResources(p.server_id, p.player_id, changes)
  ElMessage.success('资源已修改')
  resVisible.value = false
  await refreshRow(p)
}

onMounted(async () => {
  await serverStore.load()
  resourceTypes.value = await systemApi.resourceTypes()
  if (!query.server_id && serverStore.options.length) {
    query.server_id = serverStore.options[0].value
  }
})
</script>

<template>
  <div>
    <PageHeader />

    <el-card class="filter-card">
      <div class="filter-bar">
        <div class="filter-item">
          <label>区服</label>
          <el-select v-model="query.server_id" placeholder="选择区服" style="width: 200px" filterable>
            <el-option v-for="o in serverStore.options" :key="o.value" :value="o.value" :label="o.label" />
          </el-select>
        </div>
        <div class="filter-item">
          <label>玩家</label>
          <el-input v-model="query.keyword" placeholder="玩家ID / 昵称" style="width: 240px" clearable
            prefix-icon="Search" @keyup.enter="search" />
        </div>
        <el-button type="primary" @click="search">查询</el-button>
      </div>
    </el-card>

    <el-card class="table-card">
      <template #header>
        <div class="section-title">
          <el-icon><User /></el-icon>玩家列表<span class="list-count">{{ list.length }}</span>
        </div>
      </template>
      <el-table v-loading="loading" :data="list" empty-text="选择区服后点击查询">
        <el-table-column label="玩家" min-width="190" class-name="col-left" label-class-name="col-left">
          <template #default="{ row }">
            <div class="cell-user">
              <el-avatar :size="36">{{ row.name.slice(-1) }}</el-avatar>
              <div>
                <div class="main">{{ row.name }}</div>
                <div class="sub">ID {{ row.player_id }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="主城 / VIP" min-width="120">
          <template #default="{ row }">
            Lv.{{ row.level }}
            <el-tag v-if="row.vip" size="small" type="warning" effect="dark" round style="margin-left: 4px">V{{ row.vip }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="战力" min-width="130">
          <template #default="{ row }"><b>{{ formatNumber(row.power) }}</b></template>
        </el-table-column>
        <el-table-column label="联盟" min-width="110">
          <template #default="{ row }">{{ row.alliance || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" min-width="100">
          <template #default="{ row }">
            <span v-if="row.banned" class="status-dot danger">封禁</span>
            <span v-else-if="row.online" class="status-dot success">在线</span>
            <span v-else class="status-dot info">离线</span>
          </template>
        </el-table-column>
        <el-table-column label="最后登录" min-width="180">
          <template #default="{ row }">{{ formatTime(row.last_login_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="270" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" icon="View" @click="openDetail(row)">详情</el-button>
            <template v-if="canOperate">
              <el-button link type="primary" @click="openResources(row)">改资源</el-button>
              <el-button link type="warning" :disabled="!row.online" @click="kick(row)">踢下线</el-button>
              <el-button v-if="row.banned" link type="success" @click="unban(row)">解封</el-button>
              <el-button v-else link type="danger" @click="openBan(row)">封禁</el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-drawer v-model="detailVisible" title="玩家详情" size="460px">
      <template v-if="detail">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="玩家ID">{{ detail.player_id }}</el-descriptions-item>
          <el-descriptions-item label="区服">{{ detail.server_id }}</el-descriptions-item>
          <el-descriptions-item label="昵称">{{ detail.name }}</el-descriptions-item>
          <el-descriptions-item label="主城等级">{{ detail.level }}</el-descriptions-item>
          <el-descriptions-item label="VIP">{{ detail.vip }}</el-descriptions-item>
          <el-descriptions-item label="战力">{{ formatNumber(detail.power) }}</el-descriptions-item>
          <el-descriptions-item label="联盟">{{ detail.alliance || '-' }}</el-descriptions-item>
          <el-descriptions-item label="封禁">
            <template v-if="detail.banned">
              至 {{ detail.ban_until ? formatTime(detail.ban_until) : '永久' }}（{{ detail.ban_reason }}）
            </template>
            <template v-else>否</template>
          </el-descriptions-item>
          <el-descriptions-item label="注册时间">{{ formatTime(detail.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="最后登录">{{ formatTime(detail.last_login_at) }}</el-descriptions-item>
        </el-descriptions>
        <h4>资源</h4>
        <el-descriptions :column="2" border>
          <el-descriptions-item v-for="(v, k) in detail.resources" :key="k" :label="resourceLabels[k] ?? k">
            {{ formatNumber(v) }}
          </el-descriptions-item>
        </el-descriptions>
      </template>
    </el-drawer>

    <el-dialog v-model="banVisible" :title="`封禁玩家 ${banTarget?.name ?? ''}`" width="440px">
      <el-form label-width="90px">
        <el-form-item label="封禁时长">
          <el-input-number v-model="banForm.duration_hours" :min="0" />
          <span style="margin-left: 8px">小时（0 为永久）</span>
        </el-form-item>
        <el-form-item label="原因"><el-input v-model="banForm.reason" type="textarea" maxlength="128" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="banVisible = false">取消</el-button>
        <el-button type="danger" @click="submitBan">封禁</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="resVisible" :title="`修改资源 - ${resTarget?.name ?? ''}`" width="480px">
      <p style="margin-top: 0; color: #909399">填写增量，正数为增加、负数为扣除，扣除后不会低于 0。</p>
      <el-form label-width="90px">
        <el-form-item v-for="k in resourceTypes" :key="k" :label="resourceLabels[k] ?? k">
          <el-input-number v-model="resChanges[k]" :step="1000" style="width: 200px" />
          <span style="margin-left: 8px; color: #909399">当前 {{ formatNumber(resTarget?.resources[k] ?? 0) }}</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resVisible = false">取消</el-button>
        <el-button type="primary" @click="submitResources">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

