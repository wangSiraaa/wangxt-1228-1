export type Role = 'admin' | 'station' | 'owner' | 'dispatcher'

export interface User {
  id: number
  phone: string
  name: string
  role: Role
}

export interface LoginResp {
  token: string
  user: User
}

export interface Area {
  id: number
  name: string
  org_name: string
  capacity_kw: number
  threshold: number
  created_at: string
}

export interface AreaSummary {
  id: number
  name: string
  capacity_kw: number
  threshold: number
  org_name: string
  grid_capacity_kw: number
  allowed_capacity_kw: number
  remaining_kw: number
}

export type DeviceStatus = 'pending' | 'grid' | 'rejected'

export interface Device {
  id: number
  area_id: number
  owner_id: number
  model: string
  rated_kw: number
  phase: string
  grid_status: DeviceStatus
  created_at: string
}

export type DeclarationType = 'grid' | 'expand'
export type DeclarationStatus = 'pending' | 'approved' | 'rejected'

export interface Declaration {
  id: number
  area_id: number
  device_id: number
  owner_id: number
  type: DeclarationType
  capacity_kw: number
  status: DeclarationStatus
  reject_reason?: string
  created_at: string
}

export type AlarmLevel = 'warn' | 'danger'
export type AlarmStatus = 'open' | 'handled' | 'closed'

export interface Alarm {
  id: number
  area_id: number
  device_id: number
  level: AlarmLevel
  reverse_kw: number
  alarm_time: string
  status: AlarmStatus
  handled_by?: number | null
  handled_at?: string | null
  remark?: string
}

export type LimitStatus = 'executing' | 'done' | 'canceled'

export interface LimitCommand {
  id: number
  area_id: number
  ratio: number
  start_at: string
  end_at: string
  status: LimitStatus
  est_loss_kwh: number
  avg_gen_kw: number
  sample_count: number
  duration_hours: number
  created_by: number
  created_at: string
}

export interface LimitImpactResp {
  id: number
  area_id: number
  ratio: number
  start_at: string
  end_at: string
  duration_hours: number
  avg_gen_kw: number
  est_loss_kwh: number
  sample_count: number
}

export interface Point {
  ts: number
  v: number
}

export interface ApiResponse<T> {
  code: string
  message: string
  data?: T
}

export class ApiError extends Error {
  code: string
  status: number
  constructor(code: string, message: string, status: number) {
    super(message)
    this.code = code
    this.status = status
  }
}
