<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '@/api/client'
import type { AreaSummary, LimitCommand, LimitImpactResp, Point } from '@/api/types'
import KpiCard from '@/components/KpiCard.vue'
import Panel from '@/components/Panel.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { useECharts, type EChartsOption } from '@/composables/useECharts'
import { fmtDateTime, fmtKW, fmtKWh, limitStatusMeta } from '@/lib/format'

const route = useRoute()
const router = useRouter()
const id = Number(route.params.id)

const cmd = ref<LimitCommand | null>(null)
const area = ref<AreaSummary | null>(null)
const impact = ref<LimitImpactResp | null>(null)
const genSeries = ref<Point[]>([])

const chartOption = computed<EChartsOption>(() => {
  const r = cmd.value?.ratio ?? 0
  const allowed = genSeries.value.map((p) => [p.ts * 1000, +(p.v * (1 - r)).toFixed(2)])
  const loss = genSeries.value.map((p) => [p.ts * 1000, +(p.v * r).toFixed(2)])
  return {
    backgroundColor: 'transparent',
    grid: { left: 48, right: 16, top: 32, bottom: 28 },
    tooltip: {
      trigger: 'axis',
      valueFormatter: (v) => fmtKW(Number(v)),
    },
    legend: { data: ['允许出力', '限发损失'], textStyle: { color: '#8AA0BD' }, top: 4, right: 8, itemWidth: 12, itemHeight: 8 },
    xAxis: { type: 'time', axisLine: { lineStyle: { color: '#1E3050' } }, splitLine: { show: false } },
    yAxis: { type: 'value', name: 'kW', nameTextStyle: { color: '#5E7299' }, axisLine: { show: false }, splitLine: { lineStyle: { color: '#16263F' } }, axisLabel: { color: '#5E7299' } },
    series: [
      { name: '允许出力', type: 'line', stack: 'power', symbol: 'none', lineStyle: { width: 1, color: '#2EE6A6' }, areaStyle: { color: '#2EE6A6', opacity: 0.25 }, data: allowed },
      { name: '限发损失', type: 'line', stack: 'power', symbol: 'none', lineStyle: { width: 1, color: '#FF4D4F' }, areaStyle: { color: '#FF4D4F', opacity: 0.45 }, data: loss },
    ],
  }
})
const { el } = useECharts(chartOption)

onMounted(async () => {
  const all = await api.listLimits()
  cmd.value = all.find((l) => l.id === id) ?? null
  if (!cmd.value) return
  const [s, imp] = await Promise.all([api.areaSummary(cmd.value.area_id), api.limitImpact(id)])
  area.value = s
  impact.value = imp
  genSeries.value = await api.timeseries(cmd.value.area_id, 'gen', cmd.value.start_at, cmd.value.end_at)
})
</script>

<template>
  <div v-if="cmd" class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="font-display text-lg text-fog">限发指令 #{{ cmd.id }}</h2>
        <p class="text-sm text-fog-dim font-mono mt-0.5">{{ fmtDateTime(cmd.start_at) }} → {{ fmtDateTime(cmd.end_at) }}</p>
      </div>
      <StatusBadge :label="limitStatusMeta(cmd.status).label" :cls="limitStatusMeta(cmd.status).cls" />
    </div>

    <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
      <KpiCard label="限发比例" :value="(cmd.ratio * 100).toFixed(0) + '%'" accent="amber" />
      <KpiCard label="持续时长" :value="impact ? impact.duration_hours.toFixed(1) : '--'" unit="h" accent="cyan" />
      <KpiCard label="历史均发功率" :value="impact ? fmtKW(impact.avg_gen_kw, 2) : '--'" />
      <KpiCard label="预估影响电量" :value="impact ? fmtKWh(impact.est_loss_kwh, 1) : '--'" accent="amber" trend="基于历史同时段样本" />
    </div>

    <Panel title="影响电量估算" subtitle="限发期间出力分解：允许出力 + 限发损失">
      <div class="flex items-center gap-2 mb-3">
        <span class="chip border-pv/40 text-pv">目标台区 {{ area?.name }}</span>
        <span class="chip border-ink-500 text-fog-dim font-mono">样本数 {{ impact?.sample_count ?? '--' }}</span>
      </div>
      <div ref="el" class="h-72 w-full" />
    </Panel>

    <Panel title="指令信息">
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div><p class="label-tag">台区</p><p class="text-sm text-fog mt-1">{{ area?.name }}</p></div>
        <div><p class="label-tag">创建人</p><p class="text-sm text-fog mt-1 font-mono">#{{ cmd.created_by }}</p></div>
        <div><p class="label-tag">创建时间</p><p class="text-sm text-fog mt-1 font-mono text-xs">{{ fmtDateTime(cmd.created_at) }}</p></div>
        <div><p class="label-tag">预估影响</p><p class="text-sm text-amber-glow mt-1 font-mono">{{ fmtKWh(cmd.est_loss_kwh) }}</p></div>
      </div>
      <div class="mt-4">
        <button class="btn-ghost" @click="router.push({ name: 'limits' })">返回列表</button>
      </div>
    </Panel>
  </div>
</template>
