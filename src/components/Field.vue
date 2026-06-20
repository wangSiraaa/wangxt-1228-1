<script setup lang="ts">
defineProps<{ label: string; required?: boolean; hint?: string }>()
</script>

<template>
  <label class="block">
    <span class="label-tag">{{ label }}<span v-if="required" class="text-danger-glow">*</span></span>
    <slot />
    <span v-if="hint" class="block text-[11px] text-fog-dim mt-1">{{ hint }}</span>
  </label>
</template>
