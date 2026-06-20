<script setup lang="ts">
defineProps<{ text?: string }>()
</script>

<template>
  <div class="flex flex-col items-center justify-center py-12 text-fog-dim">
    <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="mb-3 opacity-50">
      <path d="M12 2L2 7l10 5 10-5-10-5z" />
      <path d="M2 17l10 5 10-5M2 12l10 5 10-5" />
    </svg>
    <p class="text-sm">{{ text || '暂无数据' }}</p>
  </div>
</template>
