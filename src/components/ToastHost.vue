<script setup lang="ts">
import { useUiStore } from '@/stores/ui'
const ui = useUiStore()
</script>

<template>
  <Teleport to="body">
    <div class="fixed top-4 right-4 z-[100] flex flex-col gap-2 w-80">
      <TransitionGroup name="toast">
        <div
          v-for="t in ui.toasts"
          :key="t.id"
          :class="[
            'panel px-4 py-3 text-sm border-l-2 flex items-start gap-2 cursor-pointer',
            t.type === 'success' && 'border-l-pv',
            t.type === 'error' && 'border-l-danger-glow',
            t.type === 'info' && 'border-l-cyan-glow',
          ]"
          @click="ui.dismiss(t.id)"
        >
          <span
            :class="[
              'mt-0.5',
              t.type === 'success' && 'text-pv',
              t.type === 'error' && 'text-danger-glow',
              t.type === 'info' && 'text-cyan-glow',
            ]"
          >●</span>
          <span class="text-fog flex-1">{{ t.message }}</span>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(20px);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>
