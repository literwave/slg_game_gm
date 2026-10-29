import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { serverApi } from '@/api'
import type { GameServer } from '@/api/types'

/** 区服列表缓存，供各页面的区服下拉框复用。 */
export const useServerStore = defineStore('servers', () => {
  const list = ref<GameServer[]>([])
  let loaded = false

  async function load(force = false) {
    if (loaded && !force) return
    list.value = await serverApi.list()
    loaded = true
  }

  const options = computed(() =>
    list.value
      .filter((s) => s.status !== 3)
      .map((s) => ({ value: s.server_id, label: `${s.server_id} - ${s.name}` })),
  )

  return { list, options, load }
})
