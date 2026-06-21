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
import { ApiError } from './types'

const now = () => new Date()
const iso = (d: Date) => d.toISOString()
const hoursAgo = (h: number) => new Date(Date.now() - h * 3600_000)
const hoursLater = (h: number) => new Date(Date.now() + h * 3600_000)

interface MockState {
  users: { id: number; phone: string; password: string; name: string; role: string }[]
  areas: Area[]
  devices: Device[]
  declarations: Declaration[]
  alarms: Alarm[]
  limits: LimitCommand[]
  remarks: LimitExecutionRemark[]
  seq: number
}

const state: MockState = {
  users: [
    { id: 1, phone: '13800000000', password: '123456', name: '系统管理员', role: 'admin' },
    { id: 2, phone: '13800000001', password: '123456', name: '业主·张工', role: 'owner' },
    { id: 3, phone: '13800000002', password: '123456', name: '阳光供电所', role: 'station' },
    { id: 4, phone: '13800000003', password: '123456', name: '调度员·李班', role: 'dispatcher' },
  ],
  areas: [
    { id: 1, name: '阳光花园台区', org_name: '阳光供电所', capacity_kw: 500, threshold: 0.8, created_at: iso(hoursAgo(720)) },
    { id: 2, name: '工业园区台区', org_name: '工业园供电所', capacity_kw: 300, threshold: 0.8, created_at: iso(hoursAgo(720)) },
  ],
  devices: [
    { id: 1, area_id: 1, owner_id: 2, model: '华为SUN2000-10KTL', rated_kw: 10, phase: 'ABC', grid_status: 'grid', created_at: iso(hoursAgo(600)) },
    { id: 2, area_id: 1, owner_id: 2, model: '阳光SG8K', rated_kw: 8, phase: 'ABC', grid_status: 'grid', created_at: iso(hoursAgo(500)) },
    { id: 3, area_id: 1, owner_id: 2, model: '固德威GW15K', rated_kw: 15, phase: 'ABC', grid_status: 'pending', created_at: iso(hoursAgo(20)) },
    { id: 4, area_id: 2, owner_id: 2, model: '锦浪GCI-20K', rated_kw: 20, phase: 'ABC', grid_status: 'grid', created_at: iso(hoursAgo(400)) },
  ],
  declarations: [
    { id: 1, area_id: 1, device_id: 1, owner_id: 2, type: 'grid', capacity_kw: 10, status: 'approved', created_at: iso(hoursAgo(600)) },
    { id: 2, area_id: 1, device_id: 2, owner_id: 2, type: 'grid', capacity_kw: 8, status: 'approved', created_at: iso(hoursAgo(500)) },
    { id: 3, area_id: 1, device_id: 3, owner_id: 2, type: 'expand', capacity_kw: 15, status: 'pending', created_at: iso(hoursAgo(20)) },
    { id: 4, area_id: 2, device_id: 4, owner_id: 2, type: 'grid', capacity_kw: 20, status: 'approved', created_at: iso(hoursAgo(400)) },
  ],
  alarms: [
    { id: 1, area_id: 1, device_id: 2, level: 'danger', reverse_kw: 6.4, alarm_time: iso(hoursAgo(3)), status: 'open', handled_by: null, handled_at: null, remark: '' },
    { id: 2, area_id: 2, device_id: 4, level: 'warn', reverse_kw: 2.1, alarm_time: iso(hoursAgo(30)), status: 'closed', handled_by: 3, handled_at: iso(hoursAgo(26)), remark: '已调整逆变器无功输出' },
  ],
  limits: [
    { id: 1, area_id: 2, ratio: 0.3, start_at: iso(hoursAgo(2)), end_at: iso(hoursLater(1)), status: 'executing', est_loss_kwh: 5.2, avg_gen_kw: 8.5, sample_count: 12, duration_hours: 3, remark_status: 'pending', remarked_est_loss_kwh: 0, created_by: 4, created_at: iso(hoursAgo(3)) },
  ],
  remarks: [],
  seq: 100,
}

const nextId = () => ++state.seq

function sumApproved(areaId: number): number {
  return state.declarations
    .filter((d) => d.area_id === areaId && d.status === 'approved')
    .reduce((s, d) => s + d.capacity_kw, 0)
}

function summaryOf(a: Area): AreaSummary {
  const grid = sumApproved(a.id)
  const allowed = a.capacity_kw * a.threshold
  return {
    id: a.id,
    name: a.name,
    capacity_kw: a.capacity_kw,
    threshold: a.threshold,
    org_name: a.org_name,
    grid_capacity_kw: grid,
    allowed_capacity_kw: allowed,
    remaining_kw: allowed - grid,
  }
}

// 时序生成：按小时生成 72h 的发电/反送功率
const genSeries: Record<number, Point[]> = {}
function buildSeries(areaId: number): Point[] {
  if (genSeries[areaId]) return genSeries[areaId]
  const cap = state.areas.find((a) => a.id === areaId)?.capacity_kw ?? 100
  const pts: Point[] = []
  for (let h = 72; h >= 0; h--) {
    const t = new Date(Date.now() - h * 3600_000)
    const hour = t.getHours()
    const dayFactor = Math.max(0, Math.sin(((hour - 6) / 12) * Math.PI))
    const gen = +(cap * 0.6 * dayFactor * (0.8 + Math.random() * 0.4)).toFixed(2)
    const reverse = dayFactor > 0.1 && Math.random() > 0.82 ? +(gen * (0.1 + Math.random() * 0.25)).toFixed(2) : 0
    pts.push({ ts: Math.floor(t.getTime() / 1000), v: 0 })
    pts[pts.length - 1] = { ts: Math.floor(t.getTime() / 1000), v: gen }
    ;(pts[pts.length - 1] as any).reverse = reverse
  }
  genSeries[areaId] = pts
  return pts
}

export const mockApi = {
  login(phone: string, password: string): LoginResp {
    const u = state.users.find((x) => x.phone === phone && x.password === password)
    if (!u) throw new ApiError('unauthorized', '手机号或密码错误', 401)
    return { token: 'mock.' + u.id, user: { id: u.id, phone: u.phone, name: u.name, role: u.role as any } }
  },
  listAreas(): Area[] {
    return state.areas
  },
  areaSummaries(): AreaSummary[] {
    return state.areas.map(summaryOf)
  },
  areaSummary(id: number): AreaSummary {
    const a = state.areas.find((x) => x.id === id)
    if (!a) throw new ApiError('not_found', '台区不存在', 404)
    return summaryOf(a)
  },
  createArea(body: Partial<Area>): Area {
    const a: Area = {
      id: nextId(),
      name: body.name || '新建台区',
      org_name: body.org_name || '供电所',
      capacity_kw: body.capacity_kw ?? 0,
      threshold: body.threshold ?? 0.8,
      created_at: iso(now()),
    }
    state.areas.push(a)
    return a
  },
  updateArea(id: number, body: Partial<Area>): Area {
    const a = state.areas.find((x) => x.id === id)
    if (!a) throw new ApiError('not_found', '台区不存在', 404)
    Object.assign(a, body)
    return a
  },
  listDevices(areaId?: number): Device[] {
    return state.devices.filter((d) => !areaId || d.area_id === areaId)
  },
  listDeclarations(status?: string, areaId?: number): Declaration[] {
    return state.declarations.filter(
      (d) => (!status || d.status === status) && (!areaId || d.area_id === areaId),
    )
  },
  createDeclaration(body: { area_id: number; device_id: number; type: string; capacity_kw: number }, ownerId: number): Declaration {
    const dev = state.devices.find((d) => d.id === body.device_id)
    if (!dev) throw new ApiError('not_found', '设备不存在', 404)
    if (dev.area_id !== body.area_id) throw new ApiError('bad_request', '设备不属于该台区', 400)
    if (body.type === 'expand') {
      const unhandled = state.alarms.some((al) => al.area_id === body.area_id && al.status !== 'closed')
      if (unhandled) throw new ApiError('alarm_unhandled', '该台区存在未关闭的反送电告警，无法提交扩容申报', 422)
    }
    const decl: Declaration = {
      id: nextId(),
      area_id: body.area_id,
      device_id: body.device_id,
      owner_id: ownerId,
      type: body.type as any,
      capacity_kw: body.capacity_kw,
      status: 'pending',
      created_at: iso(now()),
    }
    state.declarations.push(decl)
    return decl
  },
  approveDeclaration(id: number): Declaration {
    const decl = state.declarations.find((d) => d.id === id)
    if (!decl) throw new ApiError('not_found', '申报不存在', 404)
    if (decl.status !== 'pending') throw new ApiError('bad_request', '该申报不在待审批状态', 400)
    const a = state.areas.find((x) => x.id === decl.area_id)!
    const allowed = a.capacity_kw * a.threshold
    if (sumApproved(decl.area_id) + decl.capacity_kw > allowed)
      throw new ApiError('capacity_insufficient', '已并网容量+申报容量超过消纳安全阈值', 422)
    decl.status = 'approved'
    const dev = state.devices.find((d) => d.id === decl.device_id)
    if (dev) dev.grid_status = 'grid'
    return decl
  },
  rejectDeclaration(id: number, reason: string): Declaration {
    const decl = state.declarations.find((d) => d.id === id)
    if (!decl) throw new ApiError('not_found', '申报不存在', 404)
    decl.status = 'rejected'
    decl.reject_reason = reason
    return decl
  },
  listAlarms(status?: string, areaId?: number): Alarm[] {
    return state.alarms.filter((a) => (!status || a.status === status) && (!areaId || a.area_id === areaId))
  },
  handleAlarm(id: number, remark: string, handler: number): Alarm {
    const al = state.alarms.find((a) => a.id === id)
    if (!al) throw new ApiError('not_found', '告警不存在', 404)
    al.status = 'closed'
    al.handled_by = handler
    al.handled_at = iso(now())
    al.remark = remark
    return al
  },
  listLimits(areaId?: number, status?: string): LimitCommand[] {
    return state.limits.filter((l) => (!areaId || l.area_id === areaId) && (!status || l.status === status))
  },
  createLimit(body: { area_id: number; ratio: number; start_at: string; end_at: string }, userId: number): LimitCommand {
    const start = new Date(body.start_at)
    const end = new Date(body.end_at)
    const hours = (end.getTime() - start.getTime()) / 3600_000
    const pts = buildSeries(body.area_id)
    const windowPts = pts.filter((p) => {
      const t = new Date(p.ts * 1000)
      const hour = t.getHours()
      return hour >= 6 && hour <= 18
    })
    const avg = windowPts.reduce((s, p) => s + p.v, 0) / Math.max(1, windowPts.length)
    const loss = +(avg * body.ratio * hours).toFixed(2)
    const cmd: LimitCommand = {
      id: nextId(),
      area_id: body.area_id,
      ratio: body.ratio,
      start_at: body.start_at,
      end_at: body.end_at,
      status: 'executing',
      est_loss_kwh: loss,
      avg_gen_kw: +avg.toFixed(3),
      sample_count: windowPts.length,
      duration_hours: +hours.toFixed(2),
      remark_status: 'pending',
      remarked_est_loss_kwh: 0,
      created_by: userId,
      created_at: iso(now()),
    }
    state.limits.push(cmd)
    return cmd
  },
  limitImpact(id: number): LimitImpactResp {
    const cmd = state.limits.find((l) => l.id === id)
    if (!cmd) throw new ApiError('not_found', '限发指令不存在', 404)
    const start = new Date(cmd.start_at)
    const end = new Date(cmd.end_at)
    const hours = (end.getTime() - start.getTime()) / 3600_000
    const pts = buildSeries(cmd.area_id).filter((p) => {
      const t = new Date(p.ts * 1000)
      return t.getHours() >= 6 && t.getHours() <= 18
    })
    const avg = pts.reduce((s, p) => s + p.v, 0) / Math.max(1, pts.length)
    return {
      id: cmd.id,
      area_id: cmd.area_id,
      ratio: cmd.ratio,
      start_at: cmd.start_at,
      end_at: cmd.end_at,
      duration_hours: +hours.toFixed(2),
      avg_gen_kw: +avg.toFixed(3),
      est_loss_kwh: +(avg * cmd.ratio * hours).toFixed(2),
      sample_count: pts.length,
    }
  },
  listLimitRemarks(id: number): LimitExecutionRemark[] {
    const cmd = state.limits.find((l) => l.id === id)
    if (!cmd) throw new ApiError('not_found', '限发指令不存在', 404)
    return state.remarks.filter((r) => r.limit_command_id === id).sort((a, b) => b.id - a.id)
  },
  createLimitRemark(id: number, body: { block_reason: string; est_loss_kwh: number; remark: string }, userId: number): LimitExecutionRemark {
    const cmd = state.limits.find((l) => l.id === id)
    if (!cmd) throw new ApiError('not_found', '限发指令不存在', 404)
    const user = state.users.find((u) => u.id === userId)
    const remark: LimitExecutionRemark = {
      id: nextId(),
      limit_command_id: id,
      block_reason: body.block_reason,
      est_loss_kwh: body.est_loss_kwh,
      remark: body.remark,
      remarked_by: userId,
      remarked_by_name: user?.name ?? '未知用户',
      remarked_at: iso(now()),
    }
    state.remarks.push(remark)
    cmd.remark_status = 'remarked'
    cmd.remarked_est_loss_kwh = body.est_loss_kwh
    return remark
  },
  timeseries(areaId: number, metric: 'gen' | 'reverse', from: string, to: string): Point[] {
    const fromTs = new Date(from).getTime() / 1000
    const toTs = new Date(to).getTime() / 1000
    return buildSeries(areaId)
      .filter((p) => p.ts >= fromTs && p.ts <= toTs)
      .map((p) => ({ ts: p.ts, v: metric === 'reverse' ? +(p as any).reverse : p.v }))
  },
}
