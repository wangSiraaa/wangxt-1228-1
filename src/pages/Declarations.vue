<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import type { AreaSummary, Declaration } from '@/api/types'
import Panel from '@/components/Panel.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import EmptyState from '@/components/EmptyState.vue'
import { useAuthStore } from '@/stores/auth'
import { declStatusMeta, fmtDateTime, fmtKW } from '@/lib/format'

const router = useRouter()
const auth = useAuthStore()
const decls = ref<Declaration[]>([])
const summaries = ref<AreaSummary[]>([])
const statusFilter = ref<string>('')
const loading = ref(true)

const areaName = (id: number) => summaries.value.find((a) => a.id === id)?.name ?? `台区#${id}`
const canApprove = computed(() => ['station', 'admin'].includes(auth.role))
const canCreate = computed(() => ['owner', 'admin'].includes(auth.role))

const filtered = computed(() => (statusFilter.value ? decls.value.filter((d) => d.status === statusFilter.value) : decls.value))

async function load() {
  loading.value = true
  try {
    const [d, s] = await Promise.all([api.listDeclarations(), api.areaSummaries()])
    decls.value = d.sort((a, b) => +new Date(b.created_at) - +new Date(a.created_at))
    summaries.value = s
  } finally {
    loading.value = false
  }
}
onMounted(load)

const tabs = [
  { v: '', label: '全部' },
  { v: 'pending', label: '待审批' },
  { v: 'approved', label: '已通过' },
  { v: 'rejected', label: '已驳回' },
]
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between flex-wrap gap-3">
      <div class="flex items-center gap-1 panel p-1">
        <button
          v-for="t in tabs"
          :key="t.v"
          class="px-3 py-1.5 rounded-md text-xs transition-colors"
          :class="statusFilter === t.v ? 'bg-pv/15 text-pv border border-pv/40' : 'text-fog-dim hover:text-fog'"
          @click="statusFilter = t.v"
        >{{ t.label }}</button>
      </div>
      <button v-if="canCreate" class="btn-primary" @click="router.push({ name: 'declaration-new' })">+ 提交申报</button>
    </div>

    <Panel :pad="false">
      <table v-if="filtered.length" class="w-full text-sm">
        <thead>
          <tr class="text-left text-xs text-fog-dim font-mono uppercase tracking-wider">
            <th class="px-5 py-3 font-normal">台区</th>
            <th class="px-5 py-3 font-normal">类型</th>
            <th class="px-5 py-3 font-normal">容量</th>
            <th class="px-5 py-3 font-normal">提交时间</th>
            <th class="px-5 py-3 font-normal">状态</th>
            <th class="px-5 py-3 font-normal text-right">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-ink-600/50">
          <tr v-for="d in filtered" :key="d.id" class="row-hover">
            <td class="px-5 py-3 text-fog">{{ areaName(d.area_id) }}</td>
            <td class="px-5 py-3">
              <span :class="d.type === 'expand' ? 'text-amber-glow' : 'text-cyan-glow'" class="font-mono text-xs">
                {{ d.type === 'expand' ? '扩容' : '并网' }}
              </span>
            </td>
            <td class="px-5 py-3 text-fog-dim font-mono">{{ fmtKW(d.capacity_kw, 0) }}</td>
            <td class="px-5 py-3 text-fog-dim font-mono text-xs">{{ fmtDateTime(d.created_at) }}</td>
            <td class="px-5 py-3"><StatusBadge :label="declStatusMeta(d.status).label" :cls="declStatusMeta(d.status).cls" /></td>
            <td class="px-5 py-3 text-right">
              <button v-if="d.status === 'pending' && canApprove" class="btn-ghost text-xs py-1" @click="router.push({ name: 'declaration-approve', params: { id: d.id } })">审批</button>
              <span v-else-if="d.status === 'rejected'" class="text-xs text-fog-dim">{{ d.reject_reason || '—' }}</span>
              <span v-else class="text-xs text-fog-dim">—</span>
            </td>
          </tr>
        </tbody>
      </table>
      <EmptyState v-else-if="!loading" text="无符合条件的申报" />
    </Panel>
  </div>
</template>
