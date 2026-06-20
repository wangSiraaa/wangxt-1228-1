<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '@/api/client'
import { useUiStore } from '@/stores/ui'
import type { Alarm, AreaSummary, Device, Point } from '@/api/types'
import Panel from '@/components/Panel.vue'
import Field from '@/components/Field.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { useECharts, type EChartsOption } from '@/composables/useECharts'
import { alarmStatusMeta, fmtDateTime, fmtKW } from '@/lib/format'

const route = useRoute()
const router = useRouter()
const ui = useUiStore()
const id = Number(route.params.id)

const alarm = ref<Alarm | null>(null)
const area = ref<AreaSummary | null>(null)
const device = ref<Device | null>(null)
const genSeries = ref<Point[]>([])
const reverseSeries = ref<Point[]>([])
const remark = ref('')
const loading = ref(false)

const chartOption = computed<EChartsOption>(() => ({
  backgroundColor: 'transparent',
  grid: { left: 48, right: 16, top: 32, bottom: 28 },
  tooltip: { trigger: 'axis', valueFormatter: (v) => fmtKW(Number(v)) },
  legend: { data: ['发电功率', '反送电功率'], textStyle: { color: '#8AA0BD' }, top: 4, right: 8, itemWidth: 12, itemHeight: 8 },
  xAxis: { type: 'time', axisLine: { lineStyle: { color: '#1E3050' } }, splitLine: { show: false } },
  yAxis: { type: 'value', name: 'kW', nameTextStyle: { color: '#5E7299' }, axisLine: { show: false }, splitLine: { lineStyle: { color: '#16263F' } }, axisLabel: { color: '#5E7299' } },
  series: [
    { name: '发电功率', type: 'line', smooth: true, symbol: 'none', lineStyle: { width: 2, color: '#2EE6A6' }, areaStyle: { color: '#2EE6A6', opacity: 0.12 }, data: genSeries.value.map((p) => [p.ts * 1000, +p.v.toFixed(2)]) },
    { name: '反送电功率', type: 'line', smooth: true, symbol: 'none', lineStyle: { width: 2, color: '#FF4D4F' }, areaStyle: { color: '#FF4D4F', opacity: 0.15 }, data: reverseSeries.value.map((p) => [p.ts * 1000, +p.v.toFixed(2)]) },
  ],
}))
const { el } = useECharts(chartOption)

onMounted(async () => {
  const all = await api.listAlarms()
  alarm.value = all.find((a) => a.id === id) ?? null
  if (!alarm.value) return
  const [s, devs] = await Promise.all([api.areaSummary(alarm.value.area_id), api.listDevices(alarm.value.area_id)])
  area.value = s
  device.value = devs.find((d) => d.id === alarm.value!.device_id) ?? null
  const t = new Date(alarm.value.alarm_time).getTime()
  const from = new Date(t - 12 * 3600_000).toISOString()
  const to = new Date(t + 12 * 3600_000).toISOString()
  const [g, r] = await Promise.all([
    api.timeseries(alarm.value.area_id, 'gen', from, to),
    api.timeseries(alarm.value.area_id, 'reverse', from, to),
  ])
  genSeries.value = g
  reverseSeries.value = r
})

const handle = async () => {
  if (!alarm.value) return
  loading.value = true
  try {
    const updated = await api.handleAlarm(alarm.value.id, remark.value.trim() || '已处理')
    alarm.value = updated
    remark.value = ''
    ui.success('告警已关闭，该台区扩容申报限制已解除')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div v-if="alarm" class="max-w-3xl space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="font-display text-lg text-fog flex items-center gap-2">
          <span v-if="alarm.status === 'open'" class="w-2 h-2 rounded-full bg-danger-glow animate-pulseAlarm" />
          反送电告警 #{{ alarm.id }}
        </h2>
        <p class="text-sm text-fog-dim font-mono mt-0.5">{{ fmtDateTime(alarm.alarm_time) }}</p>
      </div>
      <StatusBadge :label="alarmStatusMeta(alarm.status).label" :cls="alarmStatusMeta(alarm.status).cls" />
    </div>

    <Panel title="告警信息">
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div><p class="label-tag">台区</p><p class="text-sm text-fog mt-1">{{ area?.name }}</p></div>
        <div><p class="label-tag">设备</p><p class="text-sm text-fog mt-1">{{ device?.model ?? '—' }}</p></div>
        <div><p class="label-tag">等级</p><p class="text-sm mt-1" :class="alarm.level === 'danger' ? 'text-danger-glow' : 'text-amber-glow'">{{ alarm.level === 'danger' ? '紧急' : '预警' }}</p></div>
        <div><p class="label-tag">反送功率</p><p class="text-sm text-danger-glow font-mono mt-1">{{ fmtKW(alarm.reverse_kw) }}</p></div>
      </div>
      <div v-if="alarm.status === 'closed'" class="mt-4 panel p-3">
        <p class="label-tag">处理记录</p>
        <p class="text-sm text-fog mt-1">{{ alarm.remark || '—' }}</p>
        <p class="text-xs text-fog-dim font-mono mt-1">处理时间 {{ fmtDateTime(alarm.handled_at) }}</p>
      </div>
    </Panel>

    <Panel title="功率曲线" subtitle="告警前后 24 小时" :pad="false">
      <div ref="el" class="h-64 w-full" />
    </Panel>

    <Panel v-if="alarm.status === 'open'" title="处理告警">
      <div class="space-y-3">
        <div class="panel p-3 border-amber-glow/30 bg-amber-glow/5">
          <p class="text-xs text-amber-glow font-mono">业务规则提示</p>
          <p class="text-sm text-fog mt-1">关闭此告警后，该台区的扩容申报限制将被解除，业主可重新提交扩容申报。</p>
        </div>
        <Field label="处理备注">
          <textarea v-model="remark" rows="2" class="field" placeholder="如：已调整逆变器无功输出，反送电消除" />
        </Field>
        <div class="flex items-center gap-3">
          <button class="btn-primary" :disabled="loading" @click="handle">{{ loading ? '处理中…' : '确认处理并关闭' }}</button>
          <button class="btn-ghost" @click="router.back()">返回</button>
        </div>
      </div>
    </Panel>
  </div>
</template>
