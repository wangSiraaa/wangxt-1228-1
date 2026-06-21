import { mockApi } from './mock'
import { ApiError } from './types'
import type {
  Alarm,
  Area,
  AreaSummary,
  Declaration,
  Device,
  LimitCommand,
  LimitExecutionRemark,
  LimitImpactResp,
  LoginResp,
  Point,
} from './types'

const API_BASE = '/api'
const TOKEN_KEY = 'pv_token'
const USER_KEY = 'pv_user'

export const getToken = () => localStorage.getItem(TOKEN_KEY) || ''
export const setToken = (t: string) => localStorage.setItem(TOKEN_KEY, t)
export const clearToken = () => {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}
export const getStoredUser = () => {
  const raw = localStorage.getItem(USER_KEY)
  return raw ? JSON.parse(raw) : null
}
export const setStoredUser = (u: unknown) => localStorage.setItem(USER_KEY, JSON.stringify(u))

const isNetworkError = (e: unknown) =>
  e instanceof TypeError && (e.message.includes('Failed to fetch') || e.message.includes('NetworkError'))

async function http<T>(method: string, path: string, body?: unknown, params?: Record<string, string | number>): Promise<T> {
  let url = API_BASE + path
  if (params) {
    const qs = new URLSearchParams()
    Object.entries(params).forEach(([k, v]) => qs.set(k, String(v)))
    url += '?' + qs.toString()
  }
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  const token = getToken()
  if (token) headers['Authorization'] = 'Bearer ' + token
  const res = await fetch(url, { method, headers, body: body ? JSON.stringify(body) : undefined })
  const json = await res.json().catch(() => ({ code: 'internal_error', message: '非 JSON 响应' }))
  if (json.code !== 'ok') {
    if (res.status === 401) clearToken()
    throw new ApiError(json.code || 'internal_error', json.message || '请求失败', res.status)
  }
  return json.data as T
}

function withMockFallback<T>(real: () => Promise<T>, mock: () => T): Promise<T> {
  return real().catch((e) => {
    if (e instanceof ApiError) throw e
    if (isNetworkError(e)) return mock()
    throw e
  })
}

const uid = () => getStoredUser()?.id ?? 1

export const api = {
  login(phone: string, password: string): Promise<LoginResp> {
    return withMockFallback(
      () => http<LoginResp>('POST', '/auth/login', { phone, password }),
      () => mockApi.login(phone, password),
    )
  },
  listAreas(): Promise<Area[]> {
    return withMockFallback(() => http<Area[]>('GET', '/areas'), () => mockApi.listAreas())
  },
  areaSummaries(): Promise<AreaSummary[]> {
    return withMockFallback(
      async () => {
        const areas = await http<Area[]>('GET', '/areas')
        return Promise.all(areas.map((a) => http<AreaSummary>('GET', `/areas/${a.id}`)))
      },
      () => mockApi.areaSummaries(),
    )
  },
  areaSummary(id: number): Promise<AreaSummary> {
    return withMockFallback(() => http<AreaSummary>('GET', `/areas/${id}`), () => mockApi.areaSummary(id))
  },
  createArea(body: Partial<Area>): Promise<Area> {
    return withMockFallback(() => http<Area>('POST', '/areas', body), () => mockApi.createArea(body))
  },
  updateArea(id: number, body: Partial<Area>): Promise<Area> {
    return withMockFallback(() => http<Area>('PUT', `/areas/${id}`, body), () => mockApi.updateArea(id, body))
  },
  listDevices(areaId?: number): Promise<Device[]> {
    return withMockFallback(
      () => http<Device[]>('GET', '/devices', undefined, areaId ? { area_id: areaId } : undefined),
      () => mockApi.listDevices(areaId),
    )
  },
  listDeclarations(status?: string, areaId?: number): Promise<Declaration[]> {
    const params: Record<string, string | number> = {}
    if (status) params.status = status
    if (areaId) params.area_id = areaId
    return withMockFallback(
      () => http<Declaration[]>('GET', '/declarations', undefined, params),
      () => mockApi.listDeclarations(status, areaId),
    )
  },
  createDeclaration(body: { area_id: number; device_id: number; type: string; capacity_kw: number }): Promise<Declaration> {
    return withMockFallback(
      () => http<Declaration>('POST', '/declarations', body),
      () => mockApi.createDeclaration(body, uid()),
    )
  },
  approveDeclaration(id: number): Promise<Declaration> {
    return withMockFallback(
      () => http<Declaration>('POST', `/declarations/${id}/approve`),
      () => mockApi.approveDeclaration(id),
    )
  },
  rejectDeclaration(id: number, reason: string): Promise<Declaration> {
    return withMockFallback(
      () => http<Declaration>('POST', `/declarations/${id}/reject`, { reason }),
      () => mockApi.rejectDeclaration(id, reason),
    )
  },
  listAlarms(status?: string, areaId?: number): Promise<Alarm[]> {
    const params: Record<string, string | number> = {}
    if (status) params.status = status
    if (areaId) params.area_id = areaId
    return withMockFallback(
      () => http<Alarm[]>('GET', '/alarms', undefined, params),
      () => mockApi.listAlarms(status, areaId),
    )
  },
  handleAlarm(id: number, remark: string): Promise<Alarm> {
    return withMockFallback(
      () => http<Alarm>('POST', `/alarms/${id}/handle`, { remark }),
      () => mockApi.handleAlarm(id, remark, uid()),
    )
  },
  listLimits(areaId?: number, status?: string): Promise<LimitCommand[]> {
    const params: Record<string, string | number> = {}
    if (areaId) params.area_id = areaId
    if (status) params.status = status
    return withMockFallback(
      () => http<LimitCommand[]>('GET', '/limits', undefined, params),
      () => mockApi.listLimits(areaId, status),
    )
  },
  createLimit(body: { area_id: number; ratio: number; start_at: string; end_at: string }): Promise<LimitCommand> {
    return withMockFallback(() => http<LimitCommand>('POST', '/limits', body), () => mockApi.createLimit(body, uid()))
  },
  limitImpact(id: number): Promise<LimitImpactResp> {
    return withMockFallback(() => http<LimitImpactResp>('GET', `/limits/${id}/impact`), () => mockApi.limitImpact(id))
  },
  listLimitRemarks(id: number): Promise<LimitExecutionRemark[]> {
    return withMockFallback(() => http<LimitExecutionRemark[]>('GET', `/limits/${id}/remarks`), () => mockApi.listLimitRemarks(id))
  },
  createLimitRemark(id: number, body: { block_reason: string; est_loss_kwh: number; remark: string }): Promise<LimitExecutionRemark> {
    return withMockFallback(() => http<LimitExecutionRemark>('POST', `/limits/${id}/remarks`, body), () => mockApi.createLimitRemark(id, body, uid()))
  },
  timeseries(areaId: number, metric: 'gen' | 'reverse', from: string, to: string): Promise<Point[]> {
    return withMockFallback(
      () => http<Point[]>('GET', '/timeseries', undefined, { area_id: areaId, metric, from, to }),
      () => mockApi.timeseries(areaId, metric, from, to),
    )
  },
}
