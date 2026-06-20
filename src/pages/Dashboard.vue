<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import type { Alarm, AreaSummary, Declaration, LimitCommand, Point } from '@/api/types'
import KpiCard from '@/components/KpiCard.vue'
import Panel from '@/components/Panel.vue'
import CapacityMeter from '@/components/CapacityMeter.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import EmptyState from '@/components/EmptyState.vue'
import { useECharts, type EChartsOption } from '@/composables/useECharts'
import { alarmStatusMeta, declStatusMeta, fmtDateTime, fmtKW, fmtKWh } from '@/lib/format'

const router = useRouter()
const summaries = ref<AreaSummary[]>([])
const pendingDecls = ref<Declaration[]>([])
const openAlarms = ref<Alarm[]>([])
const limits = ref<LimitCommand[]>([])
const genSeries = ref<Point[]>([])
const reverseSeries = ref<Point[]>([])
const loading = ref(true)

const totalCapacity = computed(() => summaries.value.reduce((s, a) => s + a.capacity_kw, 0))
const totalGrid = computed(() => summaries.value.reduce((s, a) => s + a.grid_capacity_kw, 0))
const totalRemaining = computed(() => summaries.value.reduce((s, a) => s + Math.max(0, a.remaining_kw), 0))
const execLimits = computed(() => limits.value.filter((l) => l.status === 'executing'))
const totalLoss = computed(() => execLimits.value.reduce((s, l) => s + (l.est_loss_kwh ?? 0), 0))

const chartOption = computed<EChartsOption>(() => {
  const mk = (pts: Point[], color: string, area: string) => ({
    name: area,
    type: 'line',
    smooth: true,
    symbol: 'none',
    lineStyle: { width: 2, color },
    areaStyle: { color, opacity: 0.12 },
    data: pts.map((p) => [p.ts * 1000, +p.v.toFixed(2)]),
  })
  return {
    backgroundColor: 'transparent',
    grid: { left: 48, right: 16, top: 36, bottom: 28 },
    tooltip: { trigger: 'axis', valueFormatter: (v) => fmtKW(Number(v)) },
    legend: { data: ['总发电功率', '反送电功率'], textStyle: { color: '#8AA0BD' }, top: 4, right: 8, itemWidth: 12, itemHeight: 8 },
    xAxis: { type: 'time', axisLine: { lineStyle: { color: '#1E3050' } }, splitLine: { show: false } },
    yAxis: { type: 'value', name: 'kW', nameTextStyle: { color: '#5E7299' }, axisLine: { show: false }, splitLine: { lineStyle: { color: '#16263F' } }, axisLabel: { color: '#5E7299' } },
    series: [mk(genSeries.value, '#2EE6A6', '总发电功率'), mk(reverseSeries.value, '#FF4D4F', '反送电功率')],
  }
})
const { el } = useECharts(chartOption)

async function load() {
  loading.value = true
  try {
    summaries.value = await api.areaSummaries()
    const [decls, alarms, lims] = await Promise.all([
      api.listDeclarations('pending'),
      api.listAlarms('open'),
      api.listLimits(),
    ])
    pendingDecls.value = decls
    openAlarms.value = alarms
    limits.value = lims

    const from = new Date(Date.now() - 72 * 3600_000).toISOString()
    const to = new Date().toISOString()
    const perArea = await Promise.all(
      summaries.value.map((a) =>
        Promise.all([api.timeseries(a.id, 'gen', from, to), api.timeseries(a.id, 'reverse', from, to)]),
      ),
    )
    const merge = (idx: 0 | 1) => {
      const map = new Map<number, number>()
      perArea.forEach((arr) => {
        arr[idx].forEach((p) => map.set(p.ts, (map.get(p.ts) ?? 0) + p.v))
      })
      return [...map.entries()].map(([ts, v]) => ({ ts, v })).sort((a, b) => a.ts - b.ts)
    }
    genSeries.value = merge(0)
    reverseSeries.value = merge(1)
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-5">
    <div class="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-3">
      <KpiCard label="台区总数" :value="String(summaries.length)" unit="个" />
      <KpiCard label="总台区容量" :value="fmtKW(totalCapacity, 0)" accent="cyan" />
      <KpiCard label="已并网容量" :value="fmtKW(totalGrid, 0)" trend="消纳余量 " />
      <KpiCard label="待审批申报" :value="String(pendingDecls.length)" unit="单" accent="amber" />
      <KpiCard label="未处理告警" :value="String(openAlarms.length)" unit="条" danger />
      <KpiCard label="执行中限发" :value="String(execLimits.length)" unit="条" trend="预估影响 " />
    </div>

    <div class="grid grid-cols-1 xl:grid-cols-3 gap-4">
      <Panel title="区域功率监测" subtitle="近 72 小时发电与反送电功率" class="xl:col-span-2" :pad="false">
        <div ref="el" class="h-72 w-full" />
      </Panel>
      <Panel title="台区消纳容量" subtitle="已并网 / 安全阈值">
        <div v-if="summaries.length" class="space-y-4">
          <div v-for="a in summaries" :key="a.id" class="cursor-pointer" @click="router.push({ name: 'area-detail', params: { id: a.id } })">
            <CapacityMeter :capacity="a.capacity_kw" :used="a.grid_capacity_kw" :threshold="a.threshold" />
            <p class="text-xs text-fog mt-1.5 flex items-center justify-between">
              <span>{{ a.name }}</span>
              <span class="text-fog-dim font-mono">{{ a.org_name }}</span>
            </p>
          </div>
        </div>
        <EmptyState v-else />
      </Panel>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <Panel title="待审批并网申报" :pad="false">
        <div v-if="pendingDecls.length" class="divide-y divide-ink-600/50">
          <RouterLink
            v-for="d in pendingDecls"
            :key="d.id"
            :to="{ name: 'declaration-approve', params: { id: d.id } }"
            class="flex items-center justify-between px-5 py-3 row-hover"
          >
            <div>
              <p class="text-sm text-fog">台区 #{{ d.area_id }} · 设备 #{{ d.device_id }}</p>
              <p class="text-xs text-fog-dim font-mono mt-0.5">{{ fmtDateTime(d.created_at) }} · {{ d.type === 'expand' ? '扩容' : '并网' }} {{ d.capacity_kw }}kW</p>
            </div>
            <StatusBadge :label="declStatusMeta(d.status).label" :cls="declStatusMeta(d.status).cls" />
          </RouterLink>
        </div>
        <EmptyState v-else text="无待审批申报" />
      </Panel>

      <Panel title="反送电告警" :pad="false">
        <div v-if="openAlarms.length" class="divide-y divide-ink-600/50">
          <RouterLink
            v-for="al in openAlarms"
            :key="al.id"
            :to="{ name: 'alarm-detail', params: { id: al.id } }"
            class="flex items-center justify-between px-5 py-3 row-hover"
          >
            <div>
              <p class="text-sm text-fog flex items-center gap-2">
                <span class="w-1.5 h-1.5 rounded-full bg-danger-glow animate-pulseAlarm" />
                台区 #{{ al.area_id }} · 反送 {{ fmtKW(al.reverse_kw) }}
              </p>
              <p class="text-xs text-fog-dim font-mono mt-0.5">{{ fmtDateTime(al.alarm_time) }}</p>
            </div>
            <StatusBadge :label="alarmStatusMeta(al.status).label" :cls="alarmStatusMeta(al.status).cls" />
          </RouterLink>
        </div>
        <EmptyState v-else text="无未处理告警" />
      </Panel>
    </div>

    <Panel v-if="execLimits.length" title="执行中限发指令" :pad="false">
      <div class="divide-y divide-ink-600/50">
        <RouterLink
          v-for="l in execLimits"
          :key="l.id"
          :to="{ name: 'limit-detail', params: { id: l.id } }"
          class="flex items-center justify-between px-5 py-3 row-hover"
        >
          <div class="flex items-center gap-4">
            <span class="font-display text-lg text-amber-glow">{{ (l.ratio * 100).toFixed(0) }}%</span>
            <div>
              <p class="text-sm text-fog">台区 #{{ l.area_id }} 限发比例</p>
              <p class="text-xs text-fog-dim font-mono mt-0.5">{{ fmtDateTime(l.start_at) }} → {{ fmtDateTime(l.end_at) }}</p>
            </div>
          </div>
          <span class="text-sm text-amber-glow font-mono">影响 {{ fmtKWh(l.est_loss_kwh) }}</span>
        </RouterLink>
      </div>
    </Panel>
  </div>
</template>
