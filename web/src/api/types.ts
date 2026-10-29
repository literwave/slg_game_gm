export type Role = 'super_admin' | 'admin' | 'operator' | 'viewer'

export interface PageData<T> {
  list: T[]
  total: number
}

export interface PageQuery {
  page: number
  page_size: number
}

export interface AdminUser {
  id: string
  username: string
  nickname: string
  role: Role
  disabled: boolean
  last_login_at?: string
  created_at: string
  updated_at: string
}

export interface LoginResp {
  token: string
  expire_at: string
  user: AdminUser
}

export type ServerStatus = 0 | 1 | 2 | 3

export interface GameServer {
  id: string
  server_id: number
  name: string
  gm_addr: string
  status: ServerStatus
  open_at: string
  remark: string
  created_at: string
  updated_at: string
  online: number
  reachable: boolean
  probed_at?: string | null
  mock: boolean
}

export interface GameServerForm {
  name: string
  gm_addr: string
  status: ServerStatus
  open_at: string
  remark: string
}

export interface Player {
  player_id: number
  server_id: number
  name: string
  level: number
  vip: number
  power: number
  alliance: string
  online: boolean
  banned: boolean
  ban_until?: string
  ban_reason?: string
  resources: Record<string, number>
  last_login_at: string
  created_at: string
}

export interface Attachment {
  item_id: number
  count: number
}

export interface DispatchResult {
  server_id: number
  success: boolean
  error?: string
}

export type MailTarget = 'player' | 'server'

export interface Mail {
  id: string
  target: MailTarget
  server_ids: number[]
  player_ids?: number[]
  title: string
  content: string
  attachments: Attachment[]
  expire_days: number
  results: DispatchResult[]
  operator: string
  created_at: string
}

export interface SendMailForm {
  target: MailTarget
  server_ids: number[]
  player_ids: number[]
  title: string
  content: string
  attachments: Attachment[]
  expire_days: number
}

export type AnnouncementStatus = 'draft' | 'published'

export interface Announcement {
  id: string
  server_ids: number[]
  title: string
  content: string
  start_at: string
  end_at: string
  interval_minutes: number
  status: AnnouncementStatus
  published_at?: string
  results?: DispatchResult[]
  operator: string
  created_at: string
  updated_at: string
}

export interface AnnouncementForm {
  server_ids: number[]
  title: string
  content: string
  start_at: string
  end_at: string
  interval_minutes: number
}

export interface OperationLog {
  id: string
  admin_id: string
  username: string
  method: string
  path: string
  query?: string
  body?: string
  ip: string
  status: number
  latency_ms: number
  created_at: string
}

export interface DashboardStats {
  server_total: number
  server_reachable: number
  online_total: number
  mail_today: number
  servers: GameServer[]
}
