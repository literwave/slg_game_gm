<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { adminApi } from '@/api'
import type { AdminUser, Role } from '@/api/types'
import PageHeader from '@/components/PageHeader.vue'
import { useTable } from '@/composables/useTable'
import { useUserStore } from '@/stores/user'
import { formatTime, roleLabel, roleOptions, roleTag } from '@/utils/format'

const store = useUserStore()
const filters = reactive({ keyword: '' })
const { list, total, page, pageSize, loading, load, search } = useTable(adminApi.list, filters)

const isSelf = (row: AdminUser) => row.id === store.user?.id

// 新建
const createVisible = ref(false)
const createRef = ref<FormInstance>()
const createForm = reactive({ username: '', password: '', nickname: '', role: 'operator' as Role })
const createRules: FormRules = {
  username: [{ required: true, pattern: /^[a-zA-Z0-9]{3,32}$/, message: '3-32 位字母或数字' }],
  password: [{ required: true, min: 6, message: '密码至少 6 位' }],
}
function openCreate() {
  Object.assign(createForm, { username: '', password: '', nickname: '', role: 'operator' })
  createVisible.value = true
}
async function submitCreate() {
  await createRef.value?.validate()
  await adminApi.create(createForm)
  ElMessage.success('创建成功')
  createVisible.value = false
  await load()
}

// 编辑
const editVisible = ref(false)
const editing = ref<AdminUser>()
const editForm = reactive({ nickname: '', role: 'operator' as Role })
function openEdit(row: AdminUser) {
  editing.value = row
  Object.assign(editForm, { nickname: row.nickname, role: row.role })
  editVisible.value = true
}
async function submitEdit() {
  await adminApi.update(editing.value!.id, editForm)
  ElMessage.success('保存成功')
  editVisible.value = false
  await load()
}

async function toggleDisabled(row: AdminUser) {
  await adminApi.update(row.id, { disabled: !row.disabled })
  ElMessage.success(row.disabled ? '已启用' : '已禁用')
  await load()
}

async function resetPassword(row: AdminUser) {
  const { value } = await ElMessageBox.prompt(`为 ${row.username} 设置新密码`, '重置密码', {
    inputType: 'password',
    inputPattern: /^.{6,64}$/,
    inputErrorMessage: '密码长度 6-64 位',
  })
  await adminApi.resetPassword(row.id, value)
  ElMessage.success('密码已重置')
}

async function remove(row: AdminUser) {
  await ElMessageBox.confirm(`确定删除账号 ${row.username}？`, '提示', { type: 'warning' })
  await adminApi.remove(row.id)
  ElMessage.success('已删除')
  await load()
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader>
      <template #actions>
        <el-button type="primary" icon="Plus" @click="openCreate">新建账号</el-button>
      </template>
    </PageHeader>

    <el-card class="filter-card">
      <div class="filter-bar">
        <div class="filter-item">
          <label>关键字</label>
          <el-input v-model="filters.keyword" placeholder="用户名 / 昵称" clearable style="width: 240px"
            prefix-icon="Search" @keyup.enter="search" />
        </div>
        <el-button type="primary" @click="search">查询</el-button>
        <el-button @click="filters.keyword = ''; search()">重置</el-button>
      </div>
    </el-card>

    <el-card class="table-card">
      <template #header>
        <div class="section-title">
          <el-icon><Avatar /></el-icon>账号列表<span class="list-count">{{ total }}</span>
          <span class="extra"><el-button size="small" icon="Refresh" circle @click="load" /></span>
        </div>
      </template>
      <el-table v-loading="loading" :data="list">
        <el-table-column label="账号" min-width="200" class-name="col-left" label-class-name="col-left">
          <template #default="{ row }">
            <div class="cell-user">
              <el-avatar :size="36">{{ (row.nickname || row.username).slice(0, 1) }}</el-avatar>
              <div>
                <div class="main">
                  {{ row.username }}
                  <el-tag v-if="isSelf(row)" size="small" effect="plain" round style="margin-left: 4px">我</el-tag>
                </div>
                <div class="sub">{{ row.nickname || '未设置昵称' }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="角色" min-width="130">
          <template #default="{ row }">
            <el-tag :type="roleTag[row.role as Role]" effect="light" round>{{ roleLabel(row.role) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" min-width="100">
          <template #default="{ row }">
            <span class="status-dot" :class="row.disabled ? 'danger' : 'success'">{{ row.disabled ? '已禁用' : '正常' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="最后登录" min-width="180">
          <template #default="{ row }">{{ formatTime(row.last_login_at) }}</template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="180">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" icon="Edit" @click="openEdit(row)">编辑</el-button>
            <el-button link type="primary" icon="Key" @click="resetPassword(row)">重置密码</el-button>
            <template v-if="!isSelf(row)">
              <el-button link :type="row.disabled ? 'success' : 'warning'" @click="toggleDisabled(row)">
                {{ row.disabled ? '启用' : '禁用' }}
              </el-button>
              <el-button link type="danger" @click="remove(row)">删除</el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" class="pager" :total="total"
        layout="total, sizes, prev, pager, next" @change="load" />

    <el-dialog v-model="createVisible" title="新建账号" width="440px">
      <el-form ref="createRef" :model="createForm" :rules="createRules" label-width="80px">
        <el-form-item label="用户名" prop="username"><el-input v-model="createForm.username" /></el-form-item>
        <el-form-item label="密码" prop="password"><el-input v-model="createForm.password" type="password" show-password /></el-form-item>
        <el-form-item label="昵称"><el-input v-model="createForm.nickname" /></el-form-item>
        <el-form-item label="角色">
          <el-select v-model="createForm.role">
            <el-option v-for="o in roleOptions" :key="o.value" :value="o.value" :label="o.label" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" @click="submitCreate">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="editVisible" :title="`编辑 ${editing?.username ?? ''}`" width="440px">
      <el-form label-width="80px">
        <el-form-item label="昵称"><el-input v-model="editForm.nickname" /></el-form-item>
        <el-form-item label="角色">
          <el-select v-model="editForm.role" :disabled="editing && isSelf(editing)">
            <el-option v-for="o in roleOptions" :key="o.value" :value="o.value" :label="o.label" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" @click="submitEdit">保存</el-button>
      </template>
    </el-dialog>
    </el-card>
  </div>
</template>

