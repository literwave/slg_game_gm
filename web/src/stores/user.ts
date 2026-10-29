import { defineStore } from 'pinia'
import { ref } from 'vue'
import { authApi } from '@/api'
import type { AdminUser, Role } from '@/api/types'
import { roleLevel } from '@/utils/format'

const TOKEN_KEY = 'slg_gm_token'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem(TOKEN_KEY) ?? '')
  const user = ref<AdminUser | null>(null)

  async function login(username: string, password: string) {
    const resp = await authApi.login(username, password)
    token.value = resp.token
    user.value = resp.user
    localStorage.setItem(TOKEN_KEY, resp.token)
  }

  async function fetchProfile() {
    user.value = await authApi.profile()
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem(TOKEN_KEY)
  }

  function hasRole(min: Role) {
    return !!user.value && roleLevel[user.value.role] >= roleLevel[min]
  }

  return { token, user, login, fetchProfile, logout, hasRole }
})
