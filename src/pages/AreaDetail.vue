<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '@/api/client'
import type { Alarm, AreaSummary, Declaration, Device, LimitCommand, Point } from '@/api/types'
import Panel from '@/components/Panel.vue'
import CapacityMeter from '@/components/CapacityMeter.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import EmptyState from '@/components/EmptyState.vue'
import { useECharts, type EChartsOption } from '@/composables/useECharts'
import { useAuthStore } from '@/stores/auth'
import { alarmStatusMeta, declStatusMeta, deviceStatusMeta, fmtDateTime, fmtKW, limitStatusMeta } from '@/lib/format'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const areaId = computed(() => Number(route.params.id))

const summary = ref<AreaSummary | null>(null)
const devices = ref<Device[]>([])
const decls = ref<Declaration[]>([])
const alarms = ref<Alarm[]>([])
const limits = ref<LimitCommand[]>([])
const genSeries = ref<Point[]>([])

const canEdit = computed(() => ['station', 'admin'].includes(auth.role))

const chartOption = computed<EChartsOption>(() => ({
  backgroundColor: 'transparent',
  grid: { left: 48, right: 16, top: 24, bottom: 28 },
  tooltip: { trigger: 'axis', valueFormatter: (v) => fmtKW(Number(v)) },
  xAxis: { type: 'time', axisLine: { lineStyle: { color: '#1E3050' } }, splitLine: { show: false } },
  yAxis: { type: 'value', name: 'kW', nameTextStyle: { color: '#5E7299' }, axisLine: { show: false }, splitLine: { lineStyle: { color: '#16263F' } }, axisLabel: { color: '#5E7299' } },
  series: [
    {
      type: 'line',
      smooth: true,
      symbol: 'none',
      lineStyle: { width: 2, color: '#2EE6A6' },
      areaStyle: { color: '#2EE6A6', opacity: 0.15 },
      data: genSeries.value.map((p) => [p.ts * 1000, +p.v.toFixed(2)]),
    },
  ],
}))
const { el } = useECharts(chartOption)

async function load() {
  const id = areaId.value
  if (!id) return
  const [s, ds, dl, al, lm] = await Promise.all([
    api.areaSummary(id),
    api.listDevices(id),
    api.listDeclarations(undefined, id),
    api.listAlarms(undefined, id),
    api.listLimits(id),
  ])
  summary.value = s
  devices.value = ds
  decls.value = dl
  alarms.value = al
  limits.value = lm
  const from = new Date(Date.now() - 72 * 3600_000).toISOString()
  const to = new Date().toISOString()
  genSeries.value = await api.timeseries(id, 'gen', from, to)
}

onMounted(load)
watch(areaId, load)
</script>

<template>
  <div v-if="summary" class="space-y-4">
    <div class="flex items-start justify-between flex-wrap gap-3">
      <div>
        <div class="flex items-center gap-3">
          <h2 class="font-display text-xl text-fog">{{ summary.name }}</h2>
          <span class="chip border border-ink-500 text-fog-dim font-mono">#{{ summary.id }}</span>
        </div>
        <p class="text-sm text-fog-dim mt-1">{{ summary.org_name }} · 消纳阈值 {{ (summary.threshold * 100).toFixed(0) }}%</p>
      </div>
      <button v-if="canEdit" class="btn-ghost" @click="router.push({ name: 'area-edit', params: { id: summary.id } })">编辑台区</button>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <Panel title="消纳容量" class="lg:col-span-2">
        <CapacityMeter :capacity="summary.capacity_kw" :used="summary.grid_capacity_kw" :threshold="summary.threshold" />
        <div class="grid grid-cols-3 gap-3 mt-4 text-center">
          <div class="panel p-3">
            <p class="text-[11px] text-fog-dim font-mono">台区容量</p>
            <p class="font-display text-pv mt-1">{{ fmtKW(summary.capacity_kw, 0) }}</p>
          </div>
          <div class="panel p-3">
            <p class="text-[11px] text-fog-dim font-mono">已并网</p>
            <p class="font-display text-cyan-glow mt-1">{{ fmtKW(summary.grid_capacity_kw, 0) }}</p>
          </div>
          <div class="panel p-3">
            <p class="text-[11px] text-fog-dim font-mono">剩余可并网</p>
            <p class="font-display" :class="summary.remaining_kw > 0 ? 'text-pv' : 'text-danger-glow'">{{ fmtKW(summary.remaining_kw, 0) }}</p>
          </div>
        </div>
      </Panel>
      <Panel title="发电功率" subtitle="近 72 小时" :pad="false">
        <div ref="el" class="h-52 w-full" />
      </Panel>
    </div>

    <Panel title="逆变器设备" :pad="false">
      <table v-if="devices.length" class="w-full text-sm">
        <thead>
          <tr class="text-left text-xs text-fog-dim font-mono uppercase tracking-wider">
            <th class="px-5 py-3 font-normal">型号</th>
            <th class="px-5 py-3 font-normal">额定功率</th>
            <th class="px-5 py-3 font-normal">相位</th>
            <th class="px-5 py-3 font-normal text-right">并网状态</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-ink-600/50">
          <tr v-for="d in devices" :key="d.id" class="row-hover">
            <td class="px-5 py-3 text-fog">{{ d.model }} <span class="text-fog-dim font-mono">#{{ d.id }}</span></td>
            <td class="px-5 py-3 text-fog-dim font-mono">{{ fmtKW(d.rated_kw, 0) }}</td>
            <td class="px-5 py-3 text-fog-dim font-mono">{{ d.phase }}</td>
            <td class="px-5 py-3 text-right"><StatusBadge :label="deviceStatusMeta(d.grid_status).label" :cls="deviceStatusMeta(d.grid_status).cls" /></td>
          </tr>
        </tbody>
      </table>
      <EmptyState v-else text="该台区暂无设备" />
    </Panel>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <Panel title="并网申报" :pad="false">
        <div v-if="decls.length" class="divide-y divide-ink-600/50">
          <div v-for="d in decls" :key="d.id" class="px-5 py-3 flex items-center justify-between">
            <div>
              <p class="text-sm text-fog">{{ d.type === 'expand' ? '扩容' : '并网' }} · {{ fmtKW(d.capacity_kw, 0) }}</p>
              <p class="text-xs text-fog-dim font-mono mt-0.5">{{ fmtDateTime(d.created_at) }}</p>
            </div>
            <StatusBadge :label="declStatusMeta(d.status).label" :cls="declStatusMeta(d.status).cls" />
          </div>
        </div>
        <EmptyState v-else text="无申报记录" />
      </Panel>
      <Panel title="反送电告警" :pad="false">
        <div v-if="alarms.length" class="divide-y divide-ink-600/50">
          <RouterLink
            v-for="al in alarms"
            :key="al.id"
            :to="{ name: 'alarm-detail', params: { id: al.id } }"
            class="px-5 py-3 flex items-center justify-between row-hover"
          >
            <div>
              <p class="text-sm text-fog flex items-center gap-2">
                <span v-if="al.status === 'open'" class="w-1.5 h-1.5 rounded-full bg-danger-glow animate-pulseAlarm" />
                反送 {{ fmtKW(al.reverse_kw) }}
              </p>
              <p class="text-xs text-fog-dim font-mono mt-0.5">{{ fmtDateTime(al.alarm_time) }}</p>
            </div>
            <StatusBadge :label="alarmStatusMeta(al.status).label" :cls="alarmStatusMeta(al.status).cls" />
          </RouterLink>
        </div>
        <EmptyState v-else text="无告警记录" />
      </Panel>
    </div>

    <Panel v-if="limits.length" title="限发指令" :pad="false">
      <div class="divide-y divide-ink-600/50">
        <RouterLink
          v-for="l in limits"
          :key="l.id"
          :to="{ name: 'limit-detail', params: { id: l.id } }"
          class="px-5 py-3 flex items-center justify-between row-hover"
        >
          <div class="flex items-center gap-4">
            <span class="font-display text-amber-glow">{{ (l.ratio * 100).toFixed(0) }}%</span>
            <div>
              <p class="text-sm text-fog">{{ fmtDateTime(l.start_at) }} → {{ fmtDateTime(l.end_at) }}</p>
              <p class="text-xs text-fog-dim font-mono">影响 {{ fmtKW(l.est_loss_kwh) }} · kWh</p>
            </div>
          </div>
          <StatusBadge :label="limitStatusMeta(l.status).label" :cls="limitStatusMeta(l.status).cls" />
        </RouterLink>
      </div>
    </Panel>
  </div>
</template>
