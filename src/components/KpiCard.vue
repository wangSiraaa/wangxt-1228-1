<script setup lang="ts">
defineProps<{ label: string; value: string; unit?: string; accent?: string; trend?: string; danger?: boolean }>()
</script>

<template>
  <div class="panel p-4 relative overflow-hidden">
    <div
      class="absolute inset-y-0 left-0 w-0.5"
      :class="danger ? 'bg-danger-glow' : accent === 'cyan' ? 'bg-cyan-glow' : accent === 'amber' ? 'bg-amber-glow' : 'bg-pv'"
    />
    <p class="text-xs text-fog-dim tracking-wider uppercase font-mono">{{ label }}</p>
    <div class="mt-1 flex items-baseline gap-1">
      <span
        class="kpi-value font-display"
        :class="danger ? 'text-danger-glow' : accent === 'cyan' ? 'text-cyan-glow' : accent === 'amber' ? 'text-amber-glow' : 'text-pv'"
      >{{ value }}</span>
      <span v-if="unit" class="text-xs text-fog font-mono">{{ unit }}</span>
    </div>
    <p v-if="trend" class="text-[11px] text-fog-dim mt-1 font-mono">{{ trend }}</p>
  </div>
</template>
