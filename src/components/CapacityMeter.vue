<script setup lang="ts">
import { computed } from 'vue'
import { fmtKW } from '@/lib/format'

const props = defineProps<{
  capacity: number
  used: number
  threshold: number
}>()

const allowed = computed(() => props.capacity * props.threshold)
const ratio = computed(() => (props.capacity > 0 ? props.used / props.capacity : 0))
const usedRatio = computed(() => (props.capacity > 0 ? Math.min(1, props.used / props.capacity) : 0))
const status = computed(() => {
  const r = ratio.value
  if (r >= props.threshold) return { label: '超阈', color: 'var(--danger)', glow: 'var(--danger)' }
  if (r >= props.threshold * 0.85) return { label: '逼近', color: 'var(--amber)', glow: 'var(--amber)' }
  return { label: '正常', color: 'var(--pv)', glow: 'var(--pv)' }
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between text-xs mb-1.5 font-mono">
      <span class="text-fog-dim">已并网 <span class="text-fog">{{ fmtKW(used) }}</span> / 容量 <span class="text-fog">{{ fmtKW(capacity) }}</span></span>
      <span :style="{ color: status.color }" class="font-semibold">{{ status.label }}</span>
    </div>
    <div class="h-3 rounded-full bg-ink-700/70 overflow-hidden relative border border-ink-600">
      <div
        class="h-full rounded-full transition-all duration-500"
        :style="{ width: usedRatio * 100 + '%', background: `linear-gradient(90deg, ${status.color}33, ${status.color})`, boxShadow: `0 0 10px ${status.glow}66` }"
      />
      <div
        class="absolute top-0 bottom-0 w-px bg-amber-glow"
        :style="{ left: threshold * 100 + '%' }"
      />
    </div>
    <div class="flex items-center justify-between text-[11px] text-fog-dim mt-1 font-mono">
      <span>安全阈值 {{ (threshold * 100).toFixed(0) }}% ({{ fmtKW(allowed) }})</span>
      <span>剩余 <span :style="{ color: status.color }">{{ fmtKW(allowed - used) }}</span></span>
    </div>
  </div>
</template>
