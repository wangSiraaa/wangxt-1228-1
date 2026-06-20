<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import { useUiStore } from '@/stores/ui'
import { ApiError } from '@/api/types'
import type { AreaSummary } from '@/api/types'
import Panel from '@/components/Panel.vue'
import Field from '@/components/Field.vue'
import { toLocalInput } from '@/lib/format'

const router = useRouter()
const ui = useUiStore()
const areas = ref<AreaSummary[]>([])
const areaId = ref<number | null>(null)
const ratio = ref(30)
const startAt = ref(toLocalInput(new Date()))
const endAt = ref(toLocalInput(new Date(Date.now() + 3 * 3600_000)))
const loading = ref(false)
const error = ref('')

const selectedArea = computed(() => areas.value.find((a) => a.id === areaId.value))
const durationHours = computed(() => {
  const s = new Date(startAt.value).getTime()
  const e = new Date(endAt.value).getTime()
  return e > s ? (e - s) / 3600_000 : 0
})

onMounted(async () => {
  areas.value = await api.areaSummaries()
})

const submit = async () => {
  error.value = ''
  if (!areaId.value) return (error.value = '请选择台区')
  if (durationHours.value <= 0) return (error.value = '结束时间需晚于开始时间')
  loading.value = true
  try {
    const cmd = await api.createLimit({
      area_id: areaId.value,
      ratio: ratio.value / 100,
      start_at: new Date(startAt.value).toISOString(),
      end_at: new Date(endAt.value).toISOString(),
    })
    ui.success(`限发已发布，预估影响 ${cmd.est_loss_kwh.toFixed(1)} kWh`)
    router.push({ name: 'limit-detail', params: { id: cmd.id } })
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '发布失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="max-w-2xl">
    <Panel title="发布限发指令" subtitle="调度员对台区下达限发比例，系统自动估算影响电量">
      <form class="space-y-4" @submit.prevent="submit">
        <Field label="目标台区" required>
          <select v-model.number="areaId" class="field">
            <option :value="null" disabled>请选择台区</option>
            <option v-for="a in areas" :key="a.id" :value="a.id">{{ a.name }}（容量 {{ a.capacity_kw }}kW）</option>
          </select>
        </Field>

        <Field label="限发比例" required hint="逆变器输出限制为额定容量的该比例">
          <div class="flex items-center gap-3">
            <input v-model.number="ratio" type="range" min="10" max="100" step="5" class="flex-1 accent-[#FFB020]" />
            <span class="font-display text-amber-glow w-14 text-right">{{ ratio }}%</span>
          </div>
          <div class="flex justify-between text-[11px] text-fog-dim font-mono mt-1">
            <span>10%（深度限发）</span><span>100%（不限）</span>
          </div>
        </Field>

        <div class="grid grid-cols-2 gap-4">
          <Field label="开始时间" required>
            <input v-model="startAt" type="datetime-local" class="field font-mono text-xs" />
          </Field>
          <Field label="结束时间" required>
            <input v-model="endAt" type="datetime-local" class="field font-mono text-xs" />
          </Field>
        </div>

        <div v-if="selectedArea && durationHours > 0" class="panel p-3 border-amber-glow/30 bg-amber-glow/5">
          <p class="text-xs text-amber-glow font-mono">影响电量预估</p>
          <p class="text-sm text-fog mt-1">
            限发 {{ ratio }}% · 持续 {{ durationHours.toFixed(1) }} 小时，按该台区历史同时段平均发电功率估算，
            预估影响电量将在指令生成后由系统计算并展示。
          </p>
        </div>

        <div v-if="error" class="text-sm text-danger-glow bg-danger-glow/10 border border-danger-glow/30 rounded-md px-3 py-2">{{ error }}</div>

        <div class="flex items-center gap-3 pt-1">
          <button type="submit" class="btn-primary" :disabled="loading">{{ loading ? '发布中…' : '发布限发' }}</button>
          <button type="button" class="btn-ghost" @click="router.back()">取消</button>
        </div>
      </form>
    </Panel>
  </div>
</template>
