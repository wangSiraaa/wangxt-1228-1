import type { Role } from '@/api/types'

export const fmtNum = (v: number, d = 1) => (Number.isFinite(v) ? v.toFixed(d) : '--')
export const fmtKW = (v: number, d = 1) => `${fmtNum(v, d)} kW`
export const fmtKWh = (v: number, d = 1) => `${fmtNum(v, d)} kWh`
export const pct = (v: number) => `${(v * 100).toFixed(0)}%`

export const fmtDateTime = (iso?: string | null) => {
  if (!iso) return '--'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '--'
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}

export const toLocalInput = (d: Date) => {
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}T${p(d.getHours())}:${p(d.getMinutes())}`
}

export const roleLabel: Record<Role, string> = {
  admin: '管理员',
  station: '供电所',
  owner: '业主',
  dispatcher: '调度员',
}

type Meta = { label: string; cls: string }

export const declStatusMeta = (s: string): Meta =>
  ({
    pending: { label: '待审批', cls: 'border-amber-glow/40 text-amber-glow' },
    approved: { label: '已通过', cls: 'border-pv/40 text-pv' },
    rejected: { label: '已驳回', cls: 'border-danger-glow/40 text-danger-glow' },
  }[s] ?? { label: s, cls: 'border-ink-500 text-fog' })

export const alarmStatusMeta = (s: string): Meta =>
  ({
    open: { label: '未处理', cls: 'border-danger-glow/50 text-danger-glow' },
    handled: { label: '处理中', cls: 'border-amber-glow/40 text-amber-glow' },
    closed: { label: '已关闭', cls: 'border-pv/40 text-pv' },
  }[s] ?? { label: s, cls: 'border-ink-500 text-fog' })

export const limitStatusMeta = (s: string): Meta =>
  ({
    executing: { label: '执行中', cls: 'border-amber-glow/40 text-amber-glow' },
    done: { label: '已完成', cls: 'border-pv/40 text-pv' },
    canceled: { label: '已取消', cls: 'border-ink-500 text-fog' },
  }[s] ?? { label: s, cls: 'border-ink-500 text-fog' })

export const deviceStatusMeta = (s: string): Meta =>
  ({
    pending: { label: '待并网', cls: 'border-amber-glow/40 text-amber-glow' },
    grid: { label: '已并网', cls: 'border-pv/40 text-pv' },
    rejected: { label: '已驳回', cls: 'border-danger-glow/40 text-danger-glow' },
  }[s] ?? { label: s, cls: 'border-ink-500 text-fog' })

export const remarkStatusMeta = (s: string): Meta =>
  ({
    pending: { label: '待记录', cls: 'border-ink-500 text-fog-dim' },
    remarked: { label: '已记录', cls: 'border-cyan-glow/40 text-cyan-glow' },
  }[s] ?? { label: s, cls: 'border-ink-500 text-fog' })
