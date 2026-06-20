<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useUiStore } from '@/stores/ui'
import { ApiError } from '@/api/types'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const ui = useUiStore()
const phone = ref('13800000000')
const password = ref('123456')
const loading = ref(false)
const error = ref('')

const roles = [
  { label: '管理员', phone: '13800000000' },
  { label: '业主', phone: '13800000001' },
  { label: '供电所', phone: '13800000002' },
  { label: '调度员', phone: '13800000003' },
]
const fill = (p: string) => {
  phone.value = p
  password.value = '123456'
  error.value = ''
}

const submit = async () => {
  loading.value = true
  error.value = ''
  try {
    await auth.login(phone.value, password.value)
    ui.success(`欢迎，${auth.user?.name}`)
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : '登录失败，请检查网络'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-6 relative overflow-hidden">
    <div class="absolute inset-0 pointer-events-none opacity-40 bg-grid" />
    <div class="absolute -top-32 -right-32 w-96 h-96 rounded-full bg-pv/10 blur-3xl" />

    <div class="relative w-full max-w-md">
      <div class="text-center mb-8">
        <div class="inline-flex items-center gap-3 mb-4">
          <div class="w-12 h-12 rounded-lg bg-pv/15 border border-pv/40 flex items-center justify-center shadow-glow">
            <span class="text-pv font-display text-2xl font-bold">PV</span>
          </div>
        </div>
        <h1 class="font-display text-2xl text-fog tracking-wide">分布式光伏消纳管理系统</h1>
        <p class="text-xs text-fog-dim font-mono mt-1 tracking-[0.3em]">GRID CONSUMPTION CONTROL</p>
      </div>

      <form class="panel p-6 space-y-4" @submit.prevent="submit">
        <label class="block">
          <span class="label-tag">手机号</span>
          <input v-model="phone" class="field mt-1.5 font-mono" placeholder="13800000000" />
        </label>
        <label class="block">
          <span class="label-tag">密码</span>
          <input v-model="password" type="password" class="field mt-1.5 font-mono" placeholder="••••••" />
        </label>

        <p v-if="error" class="text-sm text-danger-glow bg-danger-glow/10 border border-danger-glow/30 rounded-md px-3 py-2">
          {{ error }}
        </p>

        <button type="submit" class="btn-primary w-full" :disabled="loading">
          {{ loading ? '登录中…' : '登 录' }}
        </button>

        <div class="pt-2 border-t border-ink-600/60">
          <p class="text-[11px] text-fog-dim mb-2 font-mono">演示账号（密码均为 123456）</p>
          <div class="grid grid-cols-4 gap-2">
            <button
              v-for="r in roles"
              :key="r.phone"
              type="button"
              class="btn-ghost text-xs py-1.5"
              :class="phone === r.phone && 'border-pv/60 text-pv'"
              @click="fill(r.phone)"
            >{{ r.label }}</button>
          </div>
        </div>
      </form>
    </div>
  </div>
</template>
