import { del, get, post, put } from './request'
import type {
  AdminUser,
  Announcement,
  AnnouncementForm,
  DashboardStats,
  GameServer,
  GameServerForm,
  LoginResp,
  Mail,
  OperationLog,
  PageData,
  PageQuery,
  Player,
  Role,
  SendMailForm,
} from './types'

export const authApi = {
  login: (username: string, password: string) =>
    post<LoginResp>('/auth/login', { username, password }),
  profile: () => get<AdminUser>('/auth/profile'),
  changePassword: (old_password: string, new_password: string) =>
    put('/auth/password', { old_password, new_password }),
}

export const systemApi = {
  dashboard: () => get<DashboardStats>('/dashboard'),
  resourceTypes: () => get<string[]>('/meta/resource-types'),
  opLogs: (q: PageQuery & { username?: string; method?: string; path?: string; start_at?: number; end_at?: number }) =>
    get<PageData<OperationLog>>('/oplogs', q),
}

export const serverApi = {
  list: () => get<GameServer[]>('/servers'),
  create: (data: GameServerForm) => post<GameServer>('/servers', data),
  update: (id: string, data: GameServerForm) => put(`/servers/${id}`, data),
  remove: (id: string) => del(`/servers/${id}`),
  probe: (id: string) => post<GameServer>(`/servers/${id}/probe`),
}

const playerPath = (serverId: number, playerId: number) => `/players/${serverId}/${playerId}`

export const playerApi = {
  search: (server_id: number, keyword: string) => get<Player[]>('/players', { server_id, keyword }),
  get: (serverId: number, playerId: number) => get<Player>(playerPath(serverId, playerId)),
  ban: (serverId: number, playerId: number, duration_hours: number, reason: string) =>
    post(`${playerPath(serverId, playerId)}/ban`, { duration_hours, reason }),
  unban: (serverId: number, playerId: number) => post(`${playerPath(serverId, playerId)}/unban`),
  kick: (serverId: number, playerId: number) => post(`${playerPath(serverId, playerId)}/kick`),
  modifyResources: (serverId: number, playerId: number, changes: Record<string, number>) =>
    post(`${playerPath(serverId, playerId)}/resources`, { changes }),
}

export const mailApi = {
  list: (q: PageQuery & { server_id?: number; keyword?: string }) => get<PageData<Mail>>('/mails', q),
  send: (data: SendMailForm) => post<Mail>('/mails', data),
}

export const announcementApi = {
  list: (q: PageQuery & { status?: string }) => get<PageData<Announcement>>('/announcements', q),
  create: (data: AnnouncementForm) => post<Announcement>('/announcements', data),
  update: (id: string, data: AnnouncementForm) => put(`/announcements/${id}`, data),
  remove: (id: string) => del(`/announcements/${id}`),
  publish: (id: string) => post<Announcement>(`/announcements/${id}/publish`),
}

export const adminApi = {
  list: (q: PageQuery & { keyword?: string }) => get<PageData<AdminUser>>('/admins', q),
  create: (data: { username: string; password: string; nickname: string; role: Role }) =>
    post<AdminUser>('/admins', data),
  update: (id: string, data: { nickname?: string; role?: Role; disabled?: boolean }) =>
    put(`/admins/${id}`, data),
  resetPassword: (id: string, password: string) => put(`/admins/${id}/password`, { password }),
  remove: (id: string) => del(`/admins/${id}`),
}
