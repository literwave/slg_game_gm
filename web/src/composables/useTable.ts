import { ref, type Ref } from 'vue'
import type { PageData, PageQuery } from '@/api/types'

/** 分页表格通用逻辑：filters 为响应式筛选条件，fetcher 负责请求数据。 */
export function useTable<T, F extends object>(
  fetcher: (q: PageQuery & F) => Promise<PageData<T>>,
  filters: F,
) {
  const list = ref<T[]>([]) as Ref<T[]>
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(20)
  const loading = ref(false)

  async function load() {
    loading.value = true
    try {
      const data = await fetcher({ ...filters, page: page.value, page_size: pageSize.value })
      list.value = data.list
      total.value = data.total
    } finally {
      loading.value = false
    }
  }

  function search() {
    page.value = 1
    return load()
  }

  return { list, total, page, pageSize, loading, load, search }
}
