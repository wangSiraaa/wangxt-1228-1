<script setup lang="ts">
defineProps<{ label: string; cls?: string }>()
</script>

<template>
  <span
    class="chip border"
    :class="cls || 'border-ink-500 text-fog'"
  >{{ label }}</span>
</template>
