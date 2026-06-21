<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '@/api/client'
import type { AreaSummary, LimitCommand, LimitExecutionRemark, LimitImpactResp, Point } from '@/api/types'
import KpiCard from '@/components/KpiCard.vue'
import Panel from '@/components/Panel.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { useECharts, type EChartsOption } from '@/composables/useECharts'
import { fmtDateTime, fmtKW, fmtKWh, limitStatusMeta, remarkStatusMeta } from '@/lib/format'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const id = Number(route.params.id)

const cmd = ref<LimitCommand | null>(null)
const area = ref<AreaSummary | null>(null)
const impact = ref<LimitImpactResp | null>(null)
const genSeries = ref<Point[]>([])
const remarks = ref<LimitExecutionRemark[]>([])

const blockReason = ref('')
const estLossKWh = ref(0)
const remarkText = ref('')
const submitting = ref(false)

const canRemark = computed(() => ['owner', 'station', 'admin'].includes(auth.role))

const chartOption = computed<EChartsOption>(() => {
  const r = cmd.value?.ratio ?? 0
  const allowed = genSeries.value.map((p) => [p.ts * 1000, +(p.v * (1 - r)).toFixed(2)])
  const loss = genSeries.value.map((p) => [p.ts * 1000, +(p.v * r).toFixed(2)])
  return {
    backgroundColor: 'transparent',
    grid: { left: 48, right: 16, top: 32, bottom: 28 },
    tooltip: {
      trigger: 'axis',
      valueFormatter: (v) => fmtKW(Number(v)),
    },
    legend: { data: ['允许出力', '限发损失'], textStyle: { color: '#8AA0BD' }, top: 4, right: 8, itemWidth: 12, itemHeight: 8 },
    xAxis: { type: 'time', axisLine: { lineStyle: { color: '#1E3050' } }, splitLine: { show: false } },
    yAxis: { type: 'value', name: 'kW', nameTextStyle: { color: '#5E7299' }, axisLine: { show: false }, splitLine: { lineStyle: { color: '#16263F' } }, axisLabel: { color: '#5E7299' } },
    series: [
      { name: '允许出力', type: 'line', stack: 'power', symbol: 'none', lineStyle: { width: 1, color: '#2EE6A6' }, areaStyle: { color: '#2EE6A6', opacity: 0.25 }, data: allowed },
      { name: '限发损失', type: 'line', stack: 'power', symbol: 'none', lineStyle: { width: 1, color: '#FF4D4F' }, areaStyle: { color: '#FF4D4F', opacity: 0.45 }, data: loss },
    ],
  }
})
const { el } = useECharts(chartOption)

async function loadRemarks() {
  remarks.value = await api.listLimitRemarks(id)
}

async function submitRemark() {
  if (!blockReason.value.trim() || estLossKWh.value < 0) return
  submitting.value = true
  try {
    const r = await api.createLimitRemark(id, {
      block_reason: blockReason.value.trim(),
      est_loss_kwh: estLossKWh.value,
      remark: remarkText.value.trim(),
    })
    remarks.value.unshift(r)
    if (cmd.value) {
      cmd.value.remark_status = 'remarked'
      cmd.value.remarked_est_loss_kwh = r.est_loss_kwh
    }
    blockReason.value = ''
    estLossKWh.value = 0
    remarkText.value = ''
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  const all = await api.listLimits()
  cmd.value = all.find((l) => l.id === id) ?? null
  if (!cmd.value) return
  const [s, imp] = await Promise.all([api.areaSummary(cmd.value.area_id), api.limitImpact(id)])
  area.value = s
  impact.value = imp
  genSeries.value = await api.timeseries(cmd.value.area_id, 'gen', cmd.value.start_at, cmd.value.end_at)
  await loadRemarks()
})
</script>

<template>
  <div v-if="cmd" class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="font-display text-lg text-fog">限发指令 #{{ cmd.id }}</h2>
        <p class="text-sm text-fog-dim font-mono mt-0.5">{{ fmtDateTime(cmd.start_at) }} → {{ fmtDateTime(cmd.end_at) }}</p>
      </div>
      <div class="flex items-center gap-2">
        <StatusBadge :label="limitStatusMeta(cmd.status).label" :cls="limitStatusMeta(cmd.status).cls" />
        <StatusBadge :label="remarkStatusMeta(cmd.remark_status).label" :cls="remarkStatusMeta(cmd.remark_status).cls" />
      </div>
    </div>

    <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
      <KpiCard label="限发比例" :value="(cmd.ratio * 100).toFixed(0) + '%'" accent="amber" />
      <KpiCard label="持续时长" :value="impact ? impact.duration_hours.toFixed(1) : '--'" unit="h" accent="cyan" />
      <KpiCard label="历史均发功率" :value="impact ? fmtKW(impact.avg_gen_kw, 2) : '--'" />
      <KpiCard label="预估影响电量" :value="impact ? fmtKWh(impact.est_loss_kwh, 1) : '--'" accent="amber" trend="基于历史同时段样本" />
    </div>

    <Panel title="影响电量估算" subtitle="限发期间出力分解：允许出力 + 限发损失">
      <div class="flex items-center gap-2 mb-3">
        <span class="chip border-pv/40 text-pv">目标台区 {{ area?.name }}</span>
        <span class="chip border-ink-500 text-fog-dim font-mono">样本数 {{ impact?.sample_count ?? '--' }}</span>
      </div>
      <div ref="el" class="h-72 w-full" />
    </Panel>

    <Panel title="执行备注" subtitle="业主或供电所可记录限发执行受阻原因及影响电量">
      <div v-if="canRemark" class="mb-6 p-4 bg-ink-700/30 rounded-lg border border-ink-600/50">
        <h4 class="text-sm font-medium text-fog mb-3">记录执行备注</h4>
        <div class="space-y-3">
          <div>
            <label class="label-tag">受阻原因 <span class="text-danger-glow">*</span></label>
            <input
              v-model="blockReason"
              type="text"
              class="input mt-1"
              placeholder="例如：设备检修、电网故障、业主申请暂缓等"
              maxlength="200"
            />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="label-tag">影响电量估算 (kWh) <span class="text-danger-glow">*</span></label>
              <input
                v-model.number="estLossKWh"
                type="number"
                step="0.01"
                min="0"
                class="input mt-1"
                placeholder="请输入估算的影响电量"
              />
            </div>
          </div>
          <div>
            <label class="label-tag">详细说明</label>
            <textarea
              v-model="remarkText"
              class="input mt-1 resize-y min-h-[80px]"
              placeholder="可补充详细情况说明（选填，最多500字）"
              maxlength="500"
            />
          </div>
          <div class="flex justify-end">
            <button
              class="btn-primary"
              :disabled="!blockReason.trim() || estLossKWh < 0 || submitting"
              @click="submitRemark"
            >
              {{ submitting ? '提交中...' : '提交备注' }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="remarks.length" class="space-y-3">
        <div
          v-for="r in remarks"
          :key="r.id"
          class="p-4 bg-ink-700/20 rounded-lg border border-ink-600/30"
        >
          <div class="flex items-start justify-between mb-2">
            <div class="flex items-center gap-2">
              <span class="chip border-cyan-glow/40 text-cyan-glow font-medium">{{ r.remarked_by_name }}</span>
              <span class="text-xs text-fog-dim font-mono">{{ fmtDateTime(r.remarked_at) }}</span>
            </div>
            <span class="text-amber-glow font-mono text-sm">{{ fmtKWh(r.est_loss_kwh) }}</span>
          </div>
          <div class="mb-2">
            <p class="label-tag mb-1">受阻原因</p>
            <p class="text-sm text-fog">{{ r.block_reason }}</p>
          </div>
          <div v-if="r.remark">
            <p class="label-tag mb-1">详细说明</p>
            <p class="text-sm text-fog-dim">{{ r.remark }}</p>
          </div>
        </div>
      </div>
      <div v-else class="text-center py-8 text-fog-dim text-sm">
        暂无执行备注记录
      </div>
    </Panel>

    <Panel title="指令信息">
      <div class="grid grid-cols-2 md:grid-cols-5 gap-4">
        <div><p class="label-tag">台区</p><p class="text-sm text-fog mt-1">{{ area?.name }}</p></div>
        <div><p class="label-tag">创建人</p><p class="text-sm text-fog mt-1 font-mono">#{{ cmd.created_by }}</p></div>
        <div><p class="label-tag">创建时间</p><p class="text-sm text-fog mt-1 font-mono text-xs">{{ fmtDateTime(cmd.created_at) }}</p></div>
        <div><p class="label-tag">预估影响</p><p class="text-sm text-amber-glow mt-1 font-mono">{{ fmtKWh(cmd.est_loss_kwh) }}</p></div>
        <div>
          <p class="label-tag">备注状态</p>
          <StatusBadge
            :label="remarkStatusMeta(cmd.remark_status).label"
            :cls="remarkStatusMeta(cmd.remark_status).cls"
            class="mt-1"
          />
        </div>
      </div>
      <div v-if="cmd.remark_status === 'remarked'" class="mt-4 p-3 bg-amber-glow/5 border border-amber-glow/30 rounded-lg">
        <p class="label-tag mb-1">已记录影响电量</p>
        <p class="text-amber-glow font-mono text-lg">{{ fmtKWh(cmd.remarked_est_loss_kwh) }}</p>
      </div>
      <div class="mt-4">
        <button class="btn-ghost" @click="router.push({ name: 'limits' })">返回列表</button>
      </div>
    </Panel>
  </div>
</template>
