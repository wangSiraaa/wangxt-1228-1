<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { api } from '@/api/client'
import type { AreaSummary, Point } from '@/api/types'
import Panel from '@/components/Panel.vue'
import KpiCard from '@/components/KpiCard.vue'
import EmptyState from '@/components/EmptyState.vue'
import { useECharts, type EChartsOption } from '@/composables/useECharts'
import { fmtKW, fmtKWh } from '@/lib/format'

const summaries = ref<AreaSummary[]>([])
const areaId = ref<number | null>(null)
const metric = ref<'gen' | 'reverse'>('gen')
const rangeH = ref(72)
const points = ref<Point[]>([])
const loading = ref(false)

const ranges = [
  { v: 24, label: '24h' },
  { v: 72, label: '72h' },
  { v: 168, label: '7天' },
]
const metrics = [
  { v: 'gen' as const, label: '发电功率' },
  { v: 'reverse' as const, label: '反送电功率' },
]

const stats = computed(() => {
  if (!points.value.length) return { peak: 0, avg: 0, energy: 0 }
  const vs = points.value.map((p) => p.v)
  const peak = Math.max(...vs)
  const avg = vs.reduce((s, v) => s + v, 0) / vs.length
  const energy = vs.reduce((s, v) => s + v, 0)
  return { peak, avg, energy }
})

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
      lineStyle: { width: 2, color: metric.value === 'reverse' ? '#FF4D4F' : '#2EE6A6' },
      areaStyle: { color: metric.value === 'reverse' ? '#FF4D4F' : '#2EE6A6', opacity: 0.15 },
      data: points.value.map((p) => [p.ts * 1000, +p.v.toFixed(2)]),
    },
  ],
}))
const { el } = useECharts(chartOption)

async function load() {
  if (!areaId.value) return
  loading.value = true
  try {
    const from = new Date(Date.now() - rangeH.value * 3600_000).toISOString()
    const to = new Date().toISOString()
    points.value = await api.timeseries(areaId.value, metric.value, from, to)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  summaries.value = await api.areaSummaries()
  if (summaries.value[0]) areaId.value = summaries.value[0].id
  await load()
})
watch([areaId, metric, rangeH], load)
</script>

<template>
  <div class="space-y-4">
    <Panel>
      <div class="flex flex-wrap items-center gap-4">
        <div class="flex items-center gap-2">
          <span class="label-tag">台区</span>
          <select v-model.number="areaId" class="field w-auto">
            <option v-for="a in summaries" :key="a.id" :value="a.id">{{ a.name }}</option>
          </select>
        </div>
        <div class="flex items-center gap-2">
          <span class="label-tag">指标</span>
          <div class="flex items-center gap-1 panel p-1">
            <button
              v-for="m in metrics"
              :key="m.v"
              class="px-3 py-1 rounded-md text-xs transition-colors"
              :class="metric === m.v ? 'bg-pv/15 text-pv border border-pv/40' : 'text-fog-dim hover:text-fog'"
              @click="metric = m.v"
            >{{ m.label }}</button>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <span class="label-tag">范围</span>
          <div class="flex items-center gap-1 panel p-1">
            <button
              v-for="r in ranges"
              :key="r.v"
              class="px-3 py-1 rounded-md text-xs transition-colors"
              :class="rangeH === r.v ? 'bg-pv/15 text-pv border border-pv/40' : 'text-fog-dim hover:text-fog'"
              @click="rangeH = r.v"
            >{{ r.label }}</button>
          </div>
        </div>
      </div>
    </Panel>

    <div class="grid grid-cols-3 gap-3">
      <KpiCard label="峰值功率" :value="fmtKW(stats.peak, 1)" :danger="metric === 'reverse' && stats.peak > 0" />
      <KpiCard label="平均功率" :value="fmtKW(stats.avg, 2)" accent="cyan" />
      <KpiCard label="累计电量" :value="fmtKWh(stats.energy, 1)" :accent="metric === 'reverse' ? 'amber' : undefined" />
    </div>

    <Panel :title="metric === 'reverse' ? '反送电功率曲线' : '发电功率曲线'" :pad="false">
      <div v-if="points.length" ref="el" class="h-80 w-full" />
      <EmptyState v-else-if="!loading" text="所选范围内无时序数据" />
    </Panel>
  </div>
</template>
