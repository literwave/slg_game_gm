import axios, { AxiosError } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'
import { useUserStore } from '@/stores/user'

interface ApiBody<T> {
  code: number
  msg: string
  data: T
}

const http = axios.create({ baseURL: '/api/v1', timeout: 15000 })

http.interceptors.request.use((config) => {
  const { token } = useUserStore()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

http.interceptors.response.use(
  (resp) => {
    const body = resp.data as ApiBody<unknown>
    if (body.code !== 0) {
      ElMessage.error(body.msg)
      return Promise.reject(new Error(body.msg))
    }
    return body.data as never
  },
  (err: AxiosError<ApiBody<unknown>>) => {
    const msg = err.response?.data?.msg ?? err.message
    if (err.response?.status === 401) {
      useUserStore().logout()
      const current = router.currentRoute.value
      if (current.name !== 'login') {
        router.replace({ name: 'login', query: { redirect: current.fullPath } })
      }
    }
    ElMessage.error(msg)
    return Promise.reject(err)
  },
)

// 去掉空查询参数，避免向后端传递无意义的空字符串
function clean(params?: object) {
  if (!params) return undefined
  return Object.fromEntries(
    Object.entries(params).filter(([, v]) => v !== '' && v !== undefined && v !== null),
  )
}

export const get = <T>(url: string, params?: object) =>
  http.get<T, T>(url, { params: clean(params) })
export const post = <T = void>(url: string, data?: unknown) => http.post<T, T>(url, data)
export const put = <T = void>(url: string, data?: unknown) => http.put<T, T>(url, data)
export const del = <T = void>(url: string) => http.delete<T, T>(url)
