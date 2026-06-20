<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import type { AreaSummary } from '@/api/types'
import Panel from '@/components/Panel.vue'
import CapacityMeter from '@/components/CapacityMeter.vue'
import EmptyState from '@/components/EmptyState.vue'
import { useAuthStore } from '@/stores/auth'
import { fmtKW } from '@/lib/format'

const router = useRouter()
const auth = useAuthStore()
const summaries = ref<AreaSummary[]>([])
const loading = ref(true)
const canCreate = computed(() => ['station', 'admin'].includes(auth.role))

onMounted(async () => {
  try {
    summaries.value = await api.areaSummaries()
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <p class="text-sm text-fog-dim">共 {{ summaries.length }} 个台区，管理并网容量与消纳阈值</p>
      <button v-if="canCreate" class="btn-primary" @click="router.push({ name: 'area-new' })">+ 录入台区</button>
    </div>

    <Panel :pad="false">
      <table v-if="summaries.length" class="w-full text-sm">
        <thead>
          <tr class="text-left text-xs text-fog-dim font-mono uppercase tracking-wider">
            <th class="px-5 py-3 font-normal">台区</th>
            <th class="px-5 py-3 font-normal">供电所</th>
            <th class="px-5 py-3 font-normal">容量 / 已并网</th>
            <th class="px-5 py-3 font-normal text-right">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-ink-600/50">
          <tr v-for="a in summaries" :key="a.id" class="row-hover cursor-pointer" @click="router.push({ name: 'area-detail', params: { id: a.id } })">
            <td class="px-5 py-4">
              <p class="text-fog">{{ a.name }}</p>
              <p class="text-xs text-fog-dim font-mono">#{{ a.id }}</p>
            </td>
            <td class="px-5 py-4 text-fog-dim">{{ a.org_name }}</td>
            <td class="px-5 py-4 min-w-[280px]">
              <CapacityMeter :capacity="a.capacity_kw" :used="a.grid_capacity_kw" :threshold="a.threshold" />
            </td>
            <td class="px-5 py-4 text-right">
              <span class="text-pv font-mono text-xs">{{ fmtKW(a.remaining_kw) }} 余量</span>
            </td>
          </tr>
        </tbody>
      </table>
      <EmptyState v-else-if="!loading" text="暂无台区，请供电所角色录入" />
    </Panel>
  </div>
</template>
