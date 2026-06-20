<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '@/api/client'
import { useUiStore } from '@/stores/ui'
import { ApiError } from '@/api/types'
import Panel from '@/components/Panel.vue'
import Field from '@/components/Field.vue'

const route = useRoute()
const router = useRouter()
const ui = useUiStore()
const isEdit = computed(() => !!route.params.id)

const name = ref('')
const orgName = ref('')
const capacityKW = ref<number>(300)
const thresholdPct = ref(80)
const loading = ref(false)
const error = ref('')

onMounted(async () => {
  if (isEdit.value) {
    const s = await api.areaSummary(Number(route.params.id))
    name.value = s.name
    orgName.value = s.org_name
    capacityKW.value = s.capacity_kw
    thresholdPct.value = Math.round(s.threshold * 100)
  }
})

const submit = async () => {
  error.value = ''
  if (!name.value.trim()) return (error.value = '请填写台区名称')
  loading.value = true
  try {
    const body = {
      name: name.value.trim(),
      org_name: orgName.value.trim(),
      capacity_kw: capacityKW.value,
      threshold: thresholdPct.value / 100,
    }
    if (isEdit.value) await api.updateArea(Number(route.params.id), body)
    else await api.createArea(body)
    ui.success(isEdit.value ? '台区已更新' : '台区已录入')
    router.push({ name: 'areas' })
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '保存失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="max-w-2xl">
    <Panel :title="isEdit ? '编辑台区' : '录入台区'" subtitle="供电所维护台区容量与消纳安全阈值">
      <form class="space-y-4" @submit.prevent="submit">
        <Field label="台区名称" required>
          <input v-model="name" class="field" placeholder="如：阳光花园台区" />
        </Field>
        <Field label="所属供电所">
          <input v-model="orgName" class="field" placeholder="如：阳光供电所" />
        </Field>
        <div class="grid grid-cols-2 gap-4">
          <Field label="台区容量 (kW)" required hint="变压器可承载的光伏容量上限">
            <input v-model.number="capacityKW" type="number" min="0" step="1" class="field font-mono" />
          </Field>
          <Field label="消纳安全阈值" required hint="并网容量不得超过 容量 × 阈值">
            <div class="flex items-center gap-2">
              <input v-model.number="thresholdPct" type="range" min="10" max="100" step="5" class="flex-1 accent-[#2EE6A6]" />
              <span class="font-display text-pv w-12 text-right">{{ thresholdPct }}%</span>
            </div>
          </Field>
        </div>

        <div v-if="error" class="text-sm text-danger-glow bg-danger-glow/10 border border-danger-glow/30 rounded-md px-3 py-2">{{ error }}</div>

        <div class="flex items-center gap-3 pt-2">
          <button type="submit" class="btn-primary" :disabled="loading">{{ loading ? '保存中…' : '保 存' }}</button>
          <button type="button" class="btn-ghost" @click="router.back()">取消</button>
          <span v-if="isEdit" class="ml-auto text-xs text-fog-dim font-mono">编辑后容量阈值立即生效</span>
        </div>
      </form>
    </Panel>
  </div>
</template>
