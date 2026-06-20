<script setup lang="ts">
defineProps<{ title?: string; subtitle?: string; pad?: boolean }>()
</script>

<template>
  <section class="panel" :class="pad !== false && 'p-5'">
    <header v-if="title || $slots.header" class="flex items-center justify-between mb-4">
      <div>
        <h3 v-if="title" class="font-display text-sm tracking-wider text-fog uppercase">{{ title }}</h3>
        <p v-if="subtitle" class="text-xs text-fog-dim mt-0.5">{{ subtitle }}</p>
      </div>
      <slot name="header" />
    </header>
    <slot />
  </section>
</template>
