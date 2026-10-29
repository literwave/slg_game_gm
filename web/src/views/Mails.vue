<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { mailApi } from '@/api'
import type { Mail, SendMailForm } from '@/api/types'
import PageHeader from '@/components/PageHeader.vue'
import { useTable } from '@/composables/useTable'
import { useUserStore } from '@/stores/user'
import { useServerStore } from '@/stores/servers'
import { formatTime, parseIds } from '@/utils/format'

const store = useUserStore()
const serverStore = useServerStore()
const canSend = store.hasRole('operator')

const filters = reactive({ server_id: undefined as number | undefined, keyword: '' })
const { list, total, page, pageSize, loading, load, search } = useTable(mailApi.list, filters)

const failedCount = (m: Mail) => m.results.filter((r) => !r.success).length

// 发送邮件
const dialogVisible = ref(false)
const formRef = ref<FormInstance>()
const playerIdsText = ref('')
const emptyForm = (): SendMailForm => ({
  target: 'player',
  server_ids: [],
  player_ids: [],
  title: '',
  content: '',
  attachments: [],
  expire_days: 7,
})
const form = reactive<SendMailForm>(emptyForm())
const singleServer = ref<number>()
const rules: FormRules = {
  title: [{ required: true, message: '请输入标题' }],
  content: [{ required: true, message: '请输入正文' }],
}
const sending = ref(false)

function openSend() {
  Object.assign(form, emptyForm())
  singleServer.value = undefined
  playerIdsText.value = ''
  dialogVisible.value = true
}

function addAttachment() {
  form.attachments.push({ item_id: 0, count: 1 })
}

async function submit() {
  await formRef.value?.validate()
  if (form.target === 'player') {
    form.player_ids = parseIds(playerIdsText.value)
    if (!singleServer.value) return ElMessage.warning('请选择区服')
    if (form.player_ids.length === 0) return ElMessage.warning('请填写玩家ID')
    form.server_ids = [singleServer.value]
  } else {
    form.player_ids = []
  }
  if (form.attachments.some((a) => a.item_id <= 0 || a.count <= 0)) {
    return ElMessage.warning('附件的道具ID和数量必须大于 0')
  }
  sending.value = true
  try {
    const mail = await mailApi.send(form)
    const failed = failedCount(mail)
    if (failed > 0) {
      ElMessage.warning(`已发送，其中 ${failed} 个区服失败，请查看详情`)
    } else {
      ElMessage.success('发送成功')
    }
    dialogVisible.value = false
    await search()
  } finally {
    sending.value = false
  }
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
        <el-button v-if="canSend" type="primary" icon="Promotion" @click="openSend">发送邮件</el-button>
      </template>
    </PageHeader>

    <el-card class="filter-card">
      <div class="filter-bar">
        <div class="filter-item">
          <label>区服</label>
          <el-select v-model="filters.server_id" placeholder="全部区服" clearable style="width: 200px">
            <el-option v-for="o in serverStore.options" :key="o.value" :value="o.value" :label="o.label" />
          </el-select>
        </div>
        <div class="filter-item">
          <label>标题</label>
          <el-input v-model="filters.keyword" placeholder="邮件标题" clearable style="width: 220px"
            prefix-icon="Search" @keyup.enter="search" />
        </div>
        <el-button type="primary" @click="search">查询</el-button>
      </div>
    </el-card>

    <el-card class="table-card">
      <template #header>
        <div class="section-title">
          <el-icon><Message /></el-icon>邮件记录<span class="list-count">{{ total }}</span>
          <span class="extra"><el-button size="small" icon="Refresh" circle @click="load" /></span>
        </div>
      </template>
      <el-table v-loading="loading" :data="list">
        <el-table-column type="expand">
          <template #default="{ row }">
            <div class="mail-detail">
              <div class="kv-grid">
                <div class="kv-item" style="grid-column: 1 / -1"><span class="k">正文：</span><span class="v" style="white-space: normal">{{ row.content }}</span></div>
                <div v-if="row.player_ids?.length" class="kv-item" style="grid-column: 1 / -1">
                  <span class="k">玩家：</span><span class="code-tag">{{ row.player_ids.join(', ') }}</span>
                </div>
                <div class="kv-item"><span class="k">有效期：</span><span class="v">{{ row.expire_days || '默认' }} 天</span></div>
              </div>
              <el-table :data="row.results" size="small" style="width: 520px; margin-top: 12px">
                <el-table-column prop="server_id" label="区服" min-width="100" />
                <el-table-column label="结果" min-width="90">
                  <template #default="{ row: r }">
                    <span class="status-dot" :class="r.success ? 'success' : 'danger'">{{ r.success ? '成功' : '失败' }}</span>
                  </template>
                </el-table-column>
                <el-table-column prop="error" label="错误信息" />
              </el-table>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="邮件" min-width="200" class-name="col-left" label-class-name="col-left">
          <template #default="{ row }">
            <div class="cell-user">
              <el-avatar :size="36" shape="square" icon="Message" />
              <div>
                <div class="main">{{ row.title }}</div>
                <div class="sub">{{ row.target === 'player' ? `指定玩家 · ${row.player_ids?.length ?? 0} 人` : '全服邮件' }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="区服" min-width="140">
          <template #default="{ row }"><span class="code-tag">{{ row.server_ids.join(', ') }}</span></template>
        </el-table-column>
        <el-table-column label="附件" min-width="180">
          <template #default="{ row }">
            <span v-if="!row.attachments.length" style="color: var(--gm-text-secondary)">无</span>
            <span v-for="a in row.attachments.slice(0, 3)" :key="a.item_id" class="code-tag" style="margin-right: 6px">{{ a.item_id }} × {{ a.count }}</span>
            <span v-if="row.attachments.length > 3" style="color: var(--gm-text-secondary)">等 {{ row.attachments.length }} 项</span>
          </template>
        </el-table-column>
        <el-table-column label="下发结果" min-width="120">
          <template #default="{ row }">
            <span v-if="failedCount(row) === 0" class="status-dot success">全部成功</span>
            <span v-else class="status-dot danger">{{ failedCount(row) }} 服失败</span>
          </template>
        </el-table-column>
        <el-table-column prop="operator" label="操作人" min-width="110" />
        <el-table-column label="发送时间" min-width="180">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" class="pager" :total="total"
        layout="total, sizes, prev, pager, next" @change="load" />
    </el-card>

    <el-dialog v-model="dialogVisible" title="发送邮件" width="640px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="发送对象">
          <el-radio-group v-model="form.target">
            <el-radio-button value="player">指定玩家</el-radio-button>
            <el-radio-button value="server">全服</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <template v-if="form.target === 'player'">
          <el-form-item label="区服">
            <el-select v-model="singleServer" placeholder="选择区服" style="width: 240px">
              <el-option v-for="o in serverStore.options" :key="o.value" :value="o.value" :label="o.label" />
            </el-select>
          </el-form-item>
          <el-form-item label="玩家ID">
            <el-input v-model="playerIdsText" type="textarea" :rows="3" placeholder="多个ID用逗号、空格或换行分隔" />
          </el-form-item>
        </template>
        <el-form-item v-else label="区服">
          <el-select v-model="form.server_ids" multiple placeholder="不选则发送到所有区服" style="width: 100%">
            <el-option v-for="o in serverStore.options" :key="o.value" :value="o.value" :label="o.label" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题" prop="title"><el-input v-model="form.title" maxlength="64" show-word-limit /></el-form-item>
        <el-form-item label="正文" prop="content">
          <el-input v-model="form.content" type="textarea" :rows="4" maxlength="2000" show-word-limit />
        </el-form-item>
        <el-form-item label="附件">
          <div style="width: 100%">
            <div v-for="(a, i) in form.attachments" :key="i" style="display: flex; gap: 8px; margin-bottom: 8px">
              <el-input-number v-model="a.item_id" :min="0" placeholder="道具ID" controls-position="right" />
              <el-input-number v-model="a.count" :min="1" placeholder="数量" controls-position="right" />
              <el-button icon="Delete" circle @click="form.attachments.splice(i, 1)" />
            </div>
            <el-button v-if="form.attachments.length < 10" icon="Plus" @click="addAttachment">添加附件</el-button>
          </div>
        </el-form-item>
        <el-form-item label="有效期">
          <el-input-number v-model="form.expire_days" :min="0" :max="30" />
          <span style="margin-left: 8px">天（0 为游戏服默认）</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="sending" @click="submit">发送</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.mail-detail {
  padding: 4px 48px 12px;
}
</style>

