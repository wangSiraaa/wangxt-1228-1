<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import ToastHost from '@/components/ToastHost.vue'

const route = useRoute()
const isGuest = computed(() => !!route.meta.guest)
</script>

<template>
  <AppLayout v-if="!isGuest" />
  <RouterView v-else />
  <ToastHost />
</template>
