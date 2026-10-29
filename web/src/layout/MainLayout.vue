<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter, type RouteRecordRaw } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { menuRoutes } from '@/router'
import { useUserStore } from '@/stores/user'
import { authApi } from '@/api'
import { roleLabel } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const store = useUserStore()

const groupIcons: Record<string, string> = {
  运营管理: 'Operation',
  服务管理: 'Platform',
  系统管理: 'Setting',
}

interface MenuEntry {
  group?: string
  items: RouteRecordRaw[]
}

// 按 meta.group 聚合菜单，保持路由声明顺序
const menus = computed(() => {
  const entries: MenuEntry[] = []
  for (const r of menuRoutes) {
    if (r.meta?.role && !store.hasRole(r.meta.role)) continue
    const group = r.meta?.group
    const existing = group && entries.find((e) => e.group === group)
    if (existing) existing.items.push(r)
    else entries.push({ group, items: [r] })
  }
  return entries
})
const openedGroups = computed(() => menus.value.filter((m) => m.group).map((m) => m.group!))

const collapsed = ref(false)
const isMobile = ref(false)
const mobileOpen = ref(false)
const asideWidth = computed(() => (isMobile.value || !collapsed.value ? '220px' : '64px'))

function syncLayout() {
  const next = window.innerWidth < 768
  if (next !== isMobile.value) {
    isMobile.value = next
    mobileOpen.value = false
  }
}

function toggleNav() {
  if (isMobile.value) mobileOpen.value = !mobileOpen.value
  else collapsed.value = !collapsed.value
}

onMounted(() => {
  syncLayout()
  window.addEventListener('resize', syncLayout)
})
onUnmounted(() => window.removeEventListener('resize', syncLayout))
watch(() => route.path, () => {
  mobileOpen.value = false
})

function onCommand(cmd: string) {
  if (cmd === 'logout') {
    store.logout()
    router.replace({ name: 'login' })
  } else if (cmd === 'password') {
    pwdVisible.value = true
  }
}

const pwdVisible = ref(false)
const pwdFormRef = ref<FormInstance>()
const pwdForm = reactive({ old: '', next: '', confirm: '' })
const pwdRules: FormRules = {
  old: [{ required: true, message: '请输入原密码' }],
  next: [{ required: true, min: 6, message: '新密码至少 6 位' }],
  confirm: [
    {
      validator: (_r, v, cb) => (v === pwdForm.next ? cb() : cb(new Error('两次输入不一致'))),
    },
  ],
}

async function submitPassword() {
  await pwdFormRef.value?.validate()
  await authApi.changePassword(pwdForm.old, pwdForm.next)
  ElMessage.success('密码已修改，请重新登录')
  pwdVisible.value = false
  store.logout()
  router.replace({ name: 'login' })
}
</script>

<template>
  <el-container class="layout">
    <div v-if="isMobile && mobileOpen" class="mask" @click="mobileOpen = false" />
    <el-aside :width="asideWidth" class="aside" :class="{ 'is-mobile': isMobile, 'is-open': mobileOpen }">
      <div class="logo">
        <div class="logo-mark">GM</div>
        <span v-show="!collapsed" class="logo-text">游戏管理后台</span>
      </div>
      <el-scrollbar class="menu-scroll">
        <el-menu :default-active="route.path" :default-openeds="openedGroups" :collapse="!isMobile && collapsed"
          :collapse-transition="false" router class="side-menu">
          <template v-for="entry in menus" :key="entry.group ?? entry.items[0].path">
            <el-sub-menu v-if="entry.group" :index="entry.group">
              <template #title>
                <el-icon><component :is="groupIcons[entry.group] ?? 'Menu'" /></el-icon>
                <span>{{ entry.group }}</span>
              </template>
              <el-menu-item v-for="m in entry.items" :key="m.path" :index="`/${m.path}`">
                <el-icon><component :is="m.meta?.icon" /></el-icon>
                <span>{{ m.meta?.title }}</span>
              </el-menu-item>
            </el-sub-menu>
            <el-menu-item v-else :index="`/${entry.items[0].path}`">
              <el-icon><component :is="entry.items[0].meta?.icon" /></el-icon>
              <template #title>{{ entry.items[0].meta?.title }}</template>
            </el-menu-item>
          </template>
        </el-menu>
      </el-scrollbar>
    </el-aside>

    <el-container class="body">
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="toggleNav">
            <component :is="isMobile ? 'Menu' : collapsed ? 'Expand' : 'Fold'" />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="route.meta.group">{{ route.meta.group }}</el-breadcrumb-item>
            <el-breadcrumb-item v-if="route.name !== 'dashboard'">{{ route.meta.title }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <el-dropdown @command="onCommand">
          <span class="user">
            <el-avatar :size="30" class="avatar">{{ (store.user?.nickname || store.user?.username || '?').slice(0, 1) }}</el-avatar>
            <span class="user-name">{{ store.user?.nickname || store.user?.username }}</span>
            <el-tag size="small" effect="plain" round>{{ store.user ? roleLabel(store.user.role) : '' }}</el-tag>
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="password" icon="Lock">修改密码</el-dropdown-item>
              <el-dropdown-item command="logout" icon="SwitchButton" divided>退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>

    <el-dialog v-model="pwdVisible" title="修改密码" width="420px" @closed="pwdFormRef?.resetFields()">
      <el-form ref="pwdFormRef" :model="pwdForm" :rules="pwdRules" label-width="90px">
        <el-form-item label="原密码" prop="old"><el-input v-model="pwdForm.old" type="password" show-password /></el-form-item>
        <el-form-item label="新密码" prop="next"><el-input v-model="pwdForm.next" type="password" show-password /></el-form-item>
        <el-form-item label="确认密码" prop="confirm"><el-input v-model="pwdForm.confirm" type="password" show-password /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pwdVisible = false">取消</el-button>
        <el-button type="primary" @click="submitPassword">确定</el-button>
      </template>
    </el-dialog>
  </el-container>
</template>

<style scoped>
.layout {
  height: 100%;
  overflow: hidden;
}
.mask {
  position: fixed;
  inset: 0;
  z-index: 19;
  background: rgba(0, 0, 0, 0.35);
}
.aside {
  display: flex;
  flex-direction: column;
  background: #fff;
  border-right: 1px solid var(--gm-border);
  transition: width 0.2s, transform 0.2s;
  overflow: hidden;
}
.aside.is-mobile {
  position: fixed;
  inset: 0 auto 0 0;
  z-index: 20;
  height: 100%;
  transform: translateX(-100%);
  box-shadow: 4px 0 16px rgba(0, 0, 0, 0.08);
}
.aside.is-mobile.is-open {
  transform: translateX(0);
}
.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 60px;
  padding: 0 16px;
  flex-shrink: 0;
  border-bottom: 1px solid var(--gm-border);
}
.logo-mark {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 700;
  font-size: 13px;
  background: var(--gm-gradient);
}
.logo-text {
  font-size: 16px;
  font-weight: 600;
  color: var(--gm-text);
  white-space: nowrap;
}
.menu-scroll {
  flex: 1;
}
.side-menu {
  border-right: none;
  --el-menu-item-height: 44px;
  --el-menu-sub-item-height: 42px;
}
.side-menu :deep(.el-menu-item.is-active) {
  background: var(--gm-primary-light);
  font-weight: 600;
}
.side-menu:not(.el-menu--collapse) :deep(.el-menu-item.is-active)::before {
  content: '';
  position: absolute;
  left: 0;
  top: 8px;
  bottom: 8px;
  width: 3px;
  border-radius: 0 3px 3px 0;
  background: var(--el-color-primary);
}
.body {
  min-width: 0;
  overflow: hidden;
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  padding: 0 16px;
  background: #fff;
  border-bottom: 1px solid var(--gm-border);
}
.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}
.collapse-btn {
  font-size: 18px;
  cursor: pointer;
  color: var(--gm-text-secondary);
}
.collapse-btn:hover {
  color: var(--el-color-primary);
}
.user {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  outline: none;
}
.avatar {
  background: var(--gm-gradient);
  font-size: 14px;
}
.user-name {
  font-size: 14px;
  color: var(--gm-text);
}
.main {
  background: var(--gm-bg);
  padding: 24px;
  overflow: auto;
}
@media (max-width: 768px) {
  .header-left :deep(.el-breadcrumb) {
    display: none;
  }
  .user-name {
    display: none;
  }
  .main {
    padding: 16px 12px;
  }
}
</style>
