<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import { useUiStore } from '@/stores/ui'
import { ApiError } from '@/api/types'
import type { AreaSummary, Device } from '@/api/types'
import Panel from '@/components/Panel.vue'
import Field from '@/components/Field.vue'
import { fmtKW } from '@/lib/format'

const router = useRouter()
const ui = useUiStore()
const areas = ref<AreaSummary[]>([])
const devices = ref<Device[]>([])
const areaId = ref<number | null>(null)
const deviceId = ref<number | null>(null)
const type = ref<'grid' | 'expand'>('grid')
const capacityKW = ref<number>(10)
const loading = ref(false)
const blocked = ref('')

const areaDevices = computed(() => devices.value.filter((d) => d.area_id === areaId.value))
const selectedArea = computed(() => areas.value.find((a) => a.id === areaId.value))
const selectedDevice = computed(() => devices.value.find((d) => d.id === deviceId.value))

onMounted(async () => {
  const [a, ds] = await Promise.all([api.areaSummaries(), api.listDevices()])
  areas.value = a
  devices.value = ds
})

watch(areaId, () => {
  deviceId.value = null
  blocked.value = ''
})
watch(deviceId, (d) => {
  if (d) {
    const dev = devices.value.find((x) => x.id === d)
    if (dev) capacityKW.value = dev.rated_kw
  }
})

const submit = async () => {
  blocked.value = ''
  if (!areaId.value || !deviceId.value) return ui.error('请选择台区与设备')
  loading.value = true
  try {
    await api.createDeclaration({
      area_id: areaId.value,
      device_id: deviceId.value,
      type: type.value,
      capacity_kw: capacityKW.value,
    })
    ui.success('申报已提交，等待审批')
    router.push({ name: 'declarations' })
  } catch (e) {
    if (e instanceof ApiError && e.code === 'alarm_unhandled') {
      blocked.value = e.message
    } else {
      ui.error(e instanceof ApiError ? e.message : '提交失败')
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="max-w-2xl">
    <Panel title="提交并网申报" subtitle="业主选择台区设备并提交并网 / 扩容申报">
      <form class="space-y-4" @submit.prevent="submit">
        <Field label="台区" required>
          <select v-model.number="areaId" class="field">
            <option :value="null" disabled>请选择台区</option>
            <option v-for="a in areas" :key="a.id" :value="a.id">{{ a.name }}（剩余 {{ fmtKW(a.remaining_kw, 0) }}）</option>
          </select>
        </Field>

        <Field label="逆变器设备" required hint="仅显示所选台区下的设备">
          <select v-model.number="deviceId" class="field" :disabled="!areaId">
            <option :value="null" disabled>请选择设备</option>
            <option v-for="d in areaDevices" :key="d.id" :value="d.id">{{ d.model }} · {{ fmtKW(d.rated_kw, 0) }}</option>
          </select>
        </Field>

        <Field label="申报类型" required>
          <div class="grid grid-cols-2 gap-3">
            <label class="panel p-3 cursor-pointer flex items-center gap-2" :class="type === 'grid' && 'border-pv/60 bg-pv/5'">
              <input v-model="type" type="radio" value="grid" class="accent-[#2EE6A6]" />
              <span class="text-sm text-fog">新增并网</span>
            </label>
            <label class="panel p-3 cursor-pointer flex items-center gap-2" :class="type === 'expand' && 'border-amber-glow/60 bg-amber-glow/5'">
              <input v-model="type" type="radio" value="expand" class="accent-[#FFB020]" />
              <span class="text-sm text-fog">扩容申报</span>
            </label>
          </div>
        </Field>

        <Field label="申报容量 (kW)" required>
          <input v-model.number="capacityKW" type="number" min="0" step="0.1" class="field font-mono" />
        </Field>

        <div v-if="type === 'expand' && selectedArea" class="panel p-3 border-amber-glow/30 bg-amber-glow/5">
          <p class="text-xs text-amber-glow font-mono">业务规则提示</p>
          <p class="text-sm text-fog mt-1">扩容申报前需确保该台区无未关闭的反送电告警，否则将被系统拦截。</p>
        </div>

        <div v-if="blocked" class="panel p-3 border-danger-glow/40 bg-danger-glow/5">
          <p class="text-xs text-danger-glow font-mono">申报被拦截 · alarm_unhandled</p>
          <p class="text-sm text-fog mt-1">{{ blocked }}</p>
          <button type="button" class="btn-ghost text-xs mt-2 py-1" @click="router.push({ name: 'alarms' })">前往处理告警 →</button>
        </div>

        <div class="flex items-center gap-3 pt-1">
          <button type="submit" class="btn-primary" :disabled="loading">{{ loading ? '提交中…' : '提交申报' }}</button>
          <button type="button" class="btn-ghost" @click="router.back()">取消</button>
        </div>
      </form>
    </Panel>
  </div>
</template>
