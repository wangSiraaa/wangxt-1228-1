<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '@/api/client'
import { useUiStore } from '@/stores/ui'
import { ApiError } from '@/api/types'
import type { AreaSummary, Declaration, Device } from '@/api/types'
import Panel from '@/components/Panel.vue'
import Field from '@/components/Field.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { declStatusMeta, fmtDateTime, fmtKW } from '@/lib/format'

const route = useRoute()
const router = useRouter()
const ui = useUiStore()
const id = Number(route.params.id)

const decl = ref<Declaration | null>(null)
const area = ref<AreaSummary | null>(null)
const device = ref<Device | null>(null)
const reason = ref('')
const loading = ref(false)
const capError = ref('')

const usedBefore = computed(() => area.value?.grid_capacity_kw ?? 0)
const usedAfter = computed(() => usedBefore.value + (decl.value?.capacity_kw ?? 0))
const allowed = computed(() => (area.value ? area.value.capacity_kw * area.value.threshold : 0))
const willExceed = computed(() => usedAfter.value > allowed.value)
const overflow = computed(() => Math.max(0, usedAfter.value - allowed.value))

const usedRatio = computed(() => (area.value && area.value.capacity_kw > 0 ? Math.min(1, usedAfter.value / area.value.capacity_kw) : 0))
const thresholdPct = computed(() => (area.value ? area.value.threshold * 100 : 80))

onMounted(async () => {
  const d = await api.listDeclarations()
  decl.value = d.find((x) => x.id === id) ?? null
  if (!decl.value) return
  const [s, devs] = await Promise.all([api.areaSummary(decl.value.area_id), api.listDevices(decl.value.area_id)])
  area.value = s
  device.value = devs.find((x) => x.id === decl.value!.device_id) ?? null
})

const approve = async () => {
  capError.value = ''
  loading.value = true
  try {
    await api.approveDeclaration(id)
    ui.success('并网审批通过')
    router.push({ name: 'declarations' })
  } catch (e) {
    if (e instanceof ApiError && e.code === 'capacity_insufficient') capError.value = e.message
    else ui.error(e instanceof ApiError ? e.message : '审批失败')
  } finally {
    loading.value = false
  }
}
const reject = async () => {
  if (!reason.value.trim()) return ui.error('请填写驳回原因')
  loading.value = true
  try {
    await api.rejectDeclaration(id, reason.value.trim())
    ui.success('已驳回申报')
    router.push({ name: 'declarations' })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div v-if="decl && area" class="max-w-3xl space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="font-display text-lg text-fog">申报 #{{ decl.id }} 审批</h2>
        <p class="text-sm text-fog-dim font-mono mt-0.5">{{ fmtDateTime(decl.created_at) }}</p>
      </div>
      <StatusBadge :label="declStatusMeta(decl.status).label" :cls="declStatusMeta(decl.status).cls" />
    </div>

    <Panel title="申报信息">
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div><p class="label-tag">台区</p><p class="text-sm text-fog mt-1">{{ area.name }}</p></div>
        <div><p class="label-tag">设备</p><p class="text-sm text-fog mt-1">{{ device?.model ?? '—' }}</p></div>
        <div><p class="label-tag">类型</p><p class="text-sm mt-1" :class="decl.type === 'expand' ? 'text-amber-glow' : 'text-cyan-glow'">{{ decl.type === 'expand' ? '扩容' : '并网' }}</p></div>
        <div><p class="label-tag">申报容量</p><p class="text-sm text-pv font-mono mt-1">{{ fmtKW(decl.capacity_kw, 0) }}</p></div>
      </div>
    </Panel>

    <Panel title="容量校验" :subtitle="`已并网 ${fmtKW(usedBefore, 0)} + 本次 ${fmtKW(decl.capacity_kw, 0)} ≤ 阈值 ${fmtKW(allowed, 0)}`">
      <div class="h-3 rounded-full bg-ink-700/70 overflow-hidden relative border border-ink-600">
        <div
          class="h-full rounded-full transition-all duration-500"
          :style="{ width: usedRatio * 100 + '%', background: willExceed ? 'linear-gradient(90deg,#FF4D4F33,#FF4D4F)' : 'linear-gradient(90deg,#2EE6A633,#2EE6A6)', boxShadow: willExceed ? '0 0 12px #FF4D4F66' : '0 0 12px #2EE6A666' }"
        />
        <div class="absolute top-0 bottom-0 w-px bg-amber-glow" :style="{ left: thresholdPct + '%' }" />
      </div>
      <div class="flex items-center justify-between mt-3">
        <p class="text-sm font-mono" :class="willExceed ? 'text-danger-glow' : 'text-pv'">
          审批后并网 {{ fmtKW(usedAfter, 0) }} / 阈值 {{ fmtKW(allowed, 0) }}
        </p>
        <p v-if="willExceed" class="text-xs text-danger-glow font-mono">超限 {{ fmtKW(overflow, 0) }}，将无法通过审批</p>
        <p v-else class="text-xs text-pv font-mono">余量充足，可批准并网</p>
      </div>

      <div v-if="capError" class="mt-4 panel p-3 border-danger-glow/40 bg-danger-glow/5">
        <p class="text-xs text-danger-glow font-mono">审批被拒 · capacity_insufficient</p>
        <p class="text-sm text-fog mt-1">{{ capError }}</p>
      </div>
    </Panel>

    <Panel v-if="decl.status === 'pending'" title="审批操作">
      <div class="space-y-3">
        <Field label="驳回原因（仅驳回时填写）">
          <textarea v-model="reason" rows="2" class="field" placeholder="如：容量超出消纳阈值，请调整后再报" />
        </Field>
        <div class="flex items-center gap-3">
          <button class="btn-primary" :disabled="loading || willExceed" @click="approve">
            {{ willExceed ? '容量超限不可批准' : '批准并网' }}
          </button>
          <button class="btn-danger" :disabled="loading" @click="reject">驳回</button>
          <button class="btn-ghost" @click="router.back()">返回</button>
        </div>
      </div>
    </Panel>
    <Panel v-else>
      <p class="text-sm text-fog-dim">该申报已处理，状态：<span class="text-pv">{{ declStatusMeta(decl.status).label }}</span></p>
    </Panel>
  </div>
</template>
