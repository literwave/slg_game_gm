import dayjs from 'dayjs'
import type { Role, ServerStatus } from '@/api/types'

export function formatTime(value?: string | null) {
  if (!value) return '-'
  const d = dayjs(value)
  // Go 的零值时间 0001-01-01 视为未设置
  return d.isValid() && d.year() > 1970 ? d.format('YYYY-MM-DD HH:mm:ss') : '-'
}

/** 日期选择器与后端交互使用的格式（RFC3339，Go 可直接解析） */
export const API_TIME_FORMAT = 'YYYY-MM-DDTHH:mm:ssZ'

export function toApiTime(value?: string | Date) {
  return dayjs(value).format(API_TIME_FORMAT)
}

export function formatNumber(n: number) {
  return n.toLocaleString('zh-CN')
}

export const roleLevel: Record<Role, number> = {
  viewer: 1,
  operator: 2,
  admin: 3,
  super_admin: 4,
}

export const roleOptions: { value: Role; label: string }[] = [
  { value: 'super_admin', label: '超级管理员' },
  { value: 'admin', label: '管理员' },
  { value: 'operator', label: '运营' },
  { value: 'viewer', label: '只读' },
]

export const roleLabel = (r: Role) => roleOptions.find((o) => o.value === r)?.label ?? r

type TagType = 'primary' | 'success' | 'warning' | 'danger' | 'info'

export const roleTag: Record<Role, TagType> = {
  super_admin: 'danger',
  admin: 'warning',
  operator: 'primary',
  viewer: 'info',
}

export const serverStatusOptions: { value: ServerStatus; label: string; tag: TagType }[] = [
  { value: 0, label: '维护中', tag: 'warning' },
  { value: 1, label: '已开服', tag: 'success' },
  { value: 2, label: '火爆', tag: 'danger' },
  { value: 3, label: '已关服', tag: 'info' },
]

export const serverStatusOf = (s: ServerStatus, openAt?: string) => {
  const base = serverStatusOptions.find((o) => o.value === s) ?? serverStatusOptions[1]
  // 已开服但开服时间还没到，展示为未开服
  if (s === 1 && openAt && dayjs(openAt).isValid() && dayjs(openAt).isAfter(dayjs())) {
    return { value: s, label: '未开服', tag: 'primary' as TagType }
  }
  return base
}

export const resourceLabels: Record<string, string> = {
  gold: '金币',
  food: '粮食',
  wood: '木材',
  stone: '石料',
  iron: '铁矿',
}

/** 解析以逗号、空格或换行分隔的数字 ID 列表 */
export function parseIds(text: string): number[] {
  return text
    .split(/[\s,，]+/)
    .map((s) => s.trim())
    .filter(Boolean)
    .map(Number)
    .filter((n) => Number.isInteger(n) && n > 0)
}
