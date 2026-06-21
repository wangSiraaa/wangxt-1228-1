<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import type { AreaSummary, LimitCommand } from '@/api/types'
import Panel from '@/components/Panel.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import EmptyState from '@/components/EmptyState.vue'
import { useAuthStore } from '@/stores/auth'
import { fmtDateTime, fmtKWh, limitStatusMeta } from '@/lib/format'

const router = useRouter()
const auth = useAuthStore()
const limits = ref<LimitCommand[]>([])
const summaries = ref<AreaSummary[]>([])
const statusFilter = ref<string>('')
const loading = ref(true)

const areaName = (id: number) => summaries.value.find((a) => a.id === id)?.name ?? `台区#${id}`
const filtered = computed(() => (statusFilter.value ? limits.value.filter((l) => l.status === statusFilter.value) : limits.value))
const totalLoss = computed(() => limits.value.reduce((s, l) => s + (l.est_loss_kwh ?? 0), 0))
const canCreate = computed(() => ['dispatcher', 'admin'].includes(auth.role))

onMounted(async () => {
  try {
    const [lm, s] = await Promise.all([api.listLimits(), api.areaSummaries()])
    limits.value = lm.sort((a, b) => +new Date(b.created_at) - +new Date(a.created_at))
    summaries.value = s
  } finally {
    loading.value = false
  }
})

const tabs = [
  { v: '', label: '全部' },
  { v: 'executing', label: '执行中' },
  { v: 'done', label: '已完成' },
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
        <span class="chip border-amber-glow/40 text-amber-glow">累计影响 {{ fmtKWh(totalLoss, 0) }}</span>
      </div>
      <button v-if="canCreate" class="btn-primary" @click="router.push({ name: 'limit-new' })">+ 发布限发</button>
    </div>

    <Panel :pad="false">
      <table v-if="filtered.length" class="w-full text-sm">
        <thead>
          <tr class="text-left text-xs text-fog-dim font-mono uppercase tracking-wider">
            <th class="px-5 py-3 font-normal">台区</th>
            <th class="px-5 py-3 font-normal">限发比例</th>
            <th class="px-5 py-3 font-normal">执行时段</th>
            <th class="px-5 py-3 font-normal">平均功率</th>
            <th class="px-5 py-3 font-normal">影响电量</th>
            <th class="px-5 py-3 font-normal">样本</th>
            <th class="px-5 py-3 font-normal">状态</th>
            <th class="px-5 py-3 font-normal text-right">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-ink-600/50">
          <tr v-for="l in filtered" :key="l.id" class="row-hover cursor-pointer" @click="router.push({ name: 'limit-detail', params: { id: l.id } })">
            <td class="px-5 py-3 text-fog">{{ areaName(l.area_id) }}</td>
            <td class="px-5 py-3"><span class="font-display text-amber-glow">{{ (l.ratio * 100).toFixed(0) }}%</span></td>
            <td class="px-5 py-3 text-fog-dim font-mono text-xs">{{ fmtDateTime(l.start_at) }} → {{ fmtDateTime(l.end_at) }}</td>
            <td class="px-5 py-3 font-mono text-fog">{{ l.avg_gen_kw.toFixed(1) }} kW</td>
            <td class="px-5 py-3 text-amber-glow font-mono">{{ fmtKWh(l.est_loss_kwh) }}</td>
            <td class="px-5 py-3">
              <span 
                class="font-mono text-xs" 
                :class="l.sample_count > 0 ? 'text-emerald-glow' : 'text-red-glow'"
              >
                {{ l.sample_count }}
              </span>
            </td>
            <td class="px-5 py-3"><StatusBadge :label="limitStatusMeta(l.status).label" :cls="limitStatusMeta(l.status).cls" /></td>
            <td class="px-5 py-3 text-right"><span class="text-xs text-cyan-glow">查看影响 →</span></td>
          </tr>
        </tbody>
      </table>
      <EmptyState v-else-if="!loading" text="无限发指令" />
    </Panel>
  </div>
</template>
