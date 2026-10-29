<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { announcementApi, mailApi, systemApi } from '@/api'
import type { Announcement, DashboardStats, Mail } from '@/api/types'
import { useUserStore } from '@/stores/user'
import { formatNumber, formatTime, serverStatusOf } from '@/utils/format'

const store = useUserStore()
const stats = ref<DashboardStats>()
const latestMail = ref<Mail>()
const latestAnnouncement = ref<Announcement>()
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const [s, mails, anns] = await Promise.all([
      systemApi.dashboard(),
      mailApi.list({ page: 1, page_size: 1 }),
      announcementApi.list({ page: 1, page_size: 1 }),
    ])
    stats.value = s
    latestMail.value = mails.list[0]
    latestAnnouncement.value = anns.list[0]
  } finally {
    loading.value = false
  }
}

const systemStatus = computed(() => {
  const s = stats.value
  if (!s || s.server_total === 0) return '待配置'
  if (s.server_reachable === s.server_total) return '在线'
  return s.server_reachable === 0 ? '离线' : '部分异常'
})

const mailFailed = computed(() => latestMail.value?.results.filter((r) => !r.success).length ?? 0)

onMounted(load)
</script>

<template>
  <div v-loading="loading">
    <div class="page-header">
      <h1>欢迎使用游戏管理后台</h1>
      <p>{{ store.user?.nickname || store.user?.username }}，您已成功登录系统。</p>
    </div>

    <div class="stat-row">
      <div class="stat-card">
        <div class="value">{{ stats?.server_total ?? 0 }}</div>
        <div class="label">游戏区服</div>
      </div>
      <div class="stat-card">
        <div class="value">{{ formatNumber(stats?.online_total ?? 0) }}</div>
        <div class="label">当前在线</div>
      </div>
      <div class="stat-card">
        <div class="value">{{ stats?.mail_today ?? 0 }}</div>
        <div class="label">今日邮件</div>
      </div>
      <div class="stat-card">
        <div class="value">{{ systemStatus }}</div>
        <div class="label">系统状态（{{ stats?.server_reachable ?? 0 }}/{{ stats?.server_total ?? 0 }} 区服可达）</div>
      </div>
    </div>

    <el-card class="page-card">
      <template #header>
        <div class="section-title"><el-icon><Message /></el-icon>最新邮件</div>
      </template>
      <div v-if="latestMail" class="kv-grid">
        <div class="kv-item"><span class="k">标题：</span><span class="v">{{ latestMail.title }}</span></div>
        <div class="kv-item">
          <span class="k">操作人员：</span>
          <span class="v"><el-icon style="vertical-align: -2px; color: var(--el-color-primary)"><User /></el-icon> {{ latestMail.operator }}</span>
        </div>
        <div class="kv-item">
          <span class="k">发送对象：</span>
          <el-tag size="small" :type="latestMail.target === 'player' ? 'primary' : 'warning'">
            {{ latestMail.target === 'player' ? '指定玩家' : '全服' }}
          </el-tag>
        </div>
        <div class="kv-item"><span class="k">发送时间：</span><span class="v">{{ formatTime(latestMail.created_at) }}</span></div>
        <div class="kv-item">
          <span class="k">区服：</span><span class="code-tag">{{ latestMail.server_ids.join(', ') }}</span>
        </div>
        <div class="kv-item">
          <span class="k">附件：</span>
          <span v-if="!latestMail.attachments.length" class="v">无</span>
          <span v-for="a in latestMail.attachments.slice(0, 3)" :key="a.item_id" class="code-tag">{{ a.item_id }} × {{ a.count }}</span>
        </div>
        <div class="kv-item">
          <span class="k">下发结果：</span>
          <span class="v">
            <el-icon :style="{ verticalAlign: '-2px', color: mailFailed ? 'var(--el-color-danger)' : 'var(--el-color-success)' }">
              <component :is="mailFailed ? 'CircleCloseFilled' : 'CircleCheckFilled'" />
            </el-icon>
            {{ mailFailed ? `${mailFailed} 个区服失败` : '成功' }}
          </span>
        </div>
      </div>
      <el-empty v-else description="暂无邮件记录" :image-size="60" />
    </el-card>

    <el-card class="page-card">
      <template #header>
        <div class="section-title"><el-icon><Bell /></el-icon>最新公告</div>
      </template>
      <div v-if="latestAnnouncement" class="kv-grid">
        <div class="kv-item"><span class="k">标题：</span><span class="v">{{ latestAnnouncement.title }}</span></div>
        <div class="kv-item"><span class="k">创建人：</span><span class="v">{{ latestAnnouncement.operator }}</span></div>
        <div class="kv-item">
          <span class="k">状态：</span>
          <el-tag size="small" :type="latestAnnouncement.status === 'published' ? 'success' : 'info'">
            {{ latestAnnouncement.status === 'published' ? '已发布' : '草稿' }}
          </el-tag>
        </div>
        <div class="kv-item">
          <span class="k">区服：</span>
          <span class="code-tag">{{ latestAnnouncement.server_ids.length ? latestAnnouncement.server_ids.join(', ') : '全部' }}</span>
        </div>
        <div class="kv-item"><span class="k">开始时间：</span><span class="v">{{ formatTime(latestAnnouncement.start_at) }}</span></div>
        <div class="kv-item"><span class="k">结束时间：</span><span class="v">{{ formatTime(latestAnnouncement.end_at) }}</span></div>
        <div class="kv-item" style="grid-column: 1 / -1">
          <span class="k">内容：</span><span class="v">{{ latestAnnouncement.content }}</span>
        </div>
      </div>
      <el-empty v-else description="暂无公告" :image-size="60" />
    </el-card>

    <el-card>
      <template #header>
        <div class="section-title">
          <el-icon><Monitor /></el-icon>区服实时状态
          <span class="extra"><el-button size="small" icon="Refresh" @click="load">刷新</el-button></span>
        </div>
      </template>
      <el-table :data="stats?.servers ?? []">
        <el-table-column label="服务器ID" min-width="120">
          <template #default="{ row }">
            <span v-if="row.server_id" class="code-tag">{{ row.server_id }}</span>
            <span v-else style="color: var(--gm-text-secondary)">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="服务器名" min-width="140" />
        <el-table-column label="开服状态" min-width="120">
          <template #default="{ row }">
            <el-tag :type="serverStatusOf(row.status, row.open_at).tag" size="small">
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
          <template #default="{ row }">{{ formatNumber(row.online) }}</template>
        </el-table-column>
        <el-table-column label="开服时间" min-width="180">
          <template #default="{ row }">{{ formatTime(row.open_at) }}</template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

