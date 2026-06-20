<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import type { Alarm, AreaSummary } from '@/api/types'
import Panel from '@/components/Panel.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import EmptyState from '@/components/EmptyState.vue'
import { alarmStatusMeta, fmtDateTime, fmtKW } from '@/lib/format'

const router = useRouter()
const alarms = ref<Alarm[]>([])
const summaries = ref<AreaSummary[]>([])
const statusFilter = ref<string>('')
const loading = ref(true)

const areaName = (id: number) => summaries.value.find((a) => a.id === id)?.name ?? `台区#${id}`
const filtered = computed(() => (statusFilter.value ? alarms.value.filter((a) => a.status === statusFilter.value) : alarms.value))
const openCount = computed(() => alarms.value.filter((a) => a.status === 'open').length)

onMounted(async () => {
  try {
    const [al, s] = await Promise.all([api.listAlarms(), api.areaSummaries()])
    alarms.value = al.sort((a, b) => +new Date(b.alarm_time) - +new Date(a.alarm_time))
    summaries.value = s
  } finally {
    loading.value = false
  }
})

const tabs = [
  { v: '', label: '全部' },
  { v: 'open', label: '未处理' },
  { v: 'closed', label: '已关闭' },
]
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between flex-wrap gap-3">
      <div class="flex items-center gap-3">
        <div class="flex items-center gap-1 panel p-1">
          <button
            v-for="t in tabs"
            :key="t.v"
            class="px-3 py-1.5 rounded-md text-xs transition-colors"
            :class="statusFilter === t.v ? 'bg-pv/15 text-pv border border-pv/40' : 'text-fog-dim hover:text-fog'"
            @click="statusFilter = t.v"
          >{{ t.label }}</button>
        </div>
        <span v-if="openCount" class="chip border-danger-glow/40 text-danger-glow">未处理 {{ openCount }}</span>
      </div>
    </div>

    <Panel :pad="false">
      <table v-if="filtered.length" class="w-full text-sm">
        <thead>
          <tr class="text-left text-xs text-fog-dim font-mono uppercase tracking-wider">
            <th class="px-5 py-3 font-normal">台区</th>
            <th class="px-5 py-3 font-normal">等级</th>
            <th class="px-5 py-3 font-normal">反送功率</th>
            <th class="px-5 py-3 font-normal">告警时间</th>
            <th class="px-5 py-3 font-normal">状态</th>
            <th class="px-5 py-3 font-normal text-right">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-ink-600/50">
          <tr
            v-for="al in filtered"
            :key="al.id"
            class="row-hover cursor-pointer"
            :class="al.status === 'open' && 'bg-danger-glow/[0.04]'"
            @click="router.push({ name: 'alarm-detail', params: { id: al.id } })"
          >
            <td class="px-5 py-3 text-fog">
              <span class="flex items-center gap-2">
                <span v-if="al.status === 'open'" class="w-1.5 h-1.5 rounded-full bg-danger-glow animate-pulseAlarm" />
                {{ areaName(al.area_id) }}
              </span>
            </td>
            <td class="px-5 py-3">
              <span class="chip" :class="al.level === 'danger' ? 'border-danger-glow/40 text-danger-glow' : 'border-amber-glow/40 text-amber-glow'">
                {{ al.level === 'danger' ? '紧急' : '预警' }}
              </span>
            </td>
            <td class="px-5 py-3 text-fog-dim font-mono">{{ fmtKW(al.reverse_kw) }}</td>
            <td class="px-5 py-3 text-fog-dim font-mono text-xs">{{ fmtDateTime(al.alarm_time) }}</td>
            <td class="px-5 py-3"><StatusBadge :label="alarmStatusMeta(al.status).label" :cls="alarmStatusMeta(al.status).cls" /></td>
            <td class="px-5 py-3 text-right"><span class="text-xs text-cyan-glow">处理 →</span></td>
          </tr>
        </tbody>
      </table>
      <EmptyState v-else-if="!loading" text="无告警记录" />
    </Panel>
  </div>
</template>
