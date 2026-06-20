<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { roleLabel } from '@/lib/format'
import type { Role } from '@/api/types'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

interface NavItem {
  name: string
  label: string
  icon: string
  roles?: Role[]
}
const navItems: NavItem[] = [
  { name: 'dashboard', label: '运行驾驶舱', icon: 'grid' },
  { name: 'areas', label: '台区管理', icon: 'area' },
  { name: 'declarations', label: '并网申报', icon: 'doc' },
  { name: 'alarms', label: '反送电告警', icon: 'bolt' },
  { name: 'limits', label: '限发执行', icon: 'gauge' },
  { name: 'analysis', label: '消纳分析', icon: 'chart' },
]

const visibleNav = computed(() => navItems.filter((n) => !n.roles || (auth.role !== '' && n.roles.includes(auth.role))))
const title = computed(() => (route.meta.title as string) || '分布式光伏消纳管理')
const clock = ref('')
let timer = 0
const tick = () => {
  const d = new Date()
  const p = (n: number) => String(n).padStart(2, '0')
  clock.value = `${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
}
onMounted(() => {
  tick()
  timer = window.setInterval(tick, 1000)
})
onBeforeUnmount(() => clearInterval(timer))

const logout = () => {
  auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="min-h-screen flex">
    <aside class="w-56 shrink-0 border-r border-ink-600/60 flex flex-col bg-ink-900/60 backdrop-blur">
      <div class="px-5 py-5 border-b border-ink-600/60">
        <div class="flex items-center gap-2.5">
          <div class="w-8 h-8 rounded bg-pv/15 border border-pv/40 flex items-center justify-center">
            <span class="text-pv font-display text-lg font-bold">PV</span>
          </div>
          <div>
            <p class="font-display text-sm text-fog tracking-wide leading-tight">光伏消纳</p>
            <p class="text-[10px] text-fog-dim font-mono">GRID CONTROL</p>
          </div>
        </div>
      </div>
      <nav class="flex-1 py-3 px-2.5 space-y-1">
        <RouterLink
          v-for="item in visibleNav"
          :key="item.name"
          :to="{ name: item.name }"
          class="nav-link"
          active-class="nav-link-active"
        >
          <span class="nav-ico">{{ item.icon }}</span>
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>
      <div class="px-3 py-3 border-t border-ink-600/60">
        <div class="panel p-3">
          <p class="text-xs text-fog">{{ auth.user?.name }}</p>
          <p class="text-[11px] text-fog-dim font-mono">{{ auth.user?.phone }}</p>
          <span class="chip border border-cyan-glow/30 text-cyan-glow mt-1.5 inline-block">{{ auth.role ? roleLabel[auth.role] : '' }}</span>
        </div>
        <button class="btn-ghost w-full mt-2 text-xs" @click="logout">退出登录</button>
      </div>
    </aside>

    <div class="flex-1 flex flex-col min-w-0">
      <header class="h-14 shrink-0 border-b border-ink-600/60 flex items-center justify-between px-6 bg-ink-900/40 backdrop-blur">
        <div>
          <h1 class="font-display text-base text-fog tracking-wide">{{ title }}</h1>
        </div>
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-2 text-xs font-mono text-fog-dim">
            <span class="w-1.5 h-1.5 rounded-full bg-pv animate-pulse" />
            <span>系统在线</span>
          </div>
          <div class="font-mono text-sm text-pv tabular-nums">{{ clock }}</div>
        </div>
      </header>

      <main class="flex-1 overflow-y-auto p-6">
        <RouterView v-slot="{ Component }">
          <Transition name="fade" mode="out-in">
            <component :is="Component" />
          </Transition>
        </RouterView>
      </main>
    </div>
  </div>
</template>

<style scoped>
.nav-link {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.5rem 0.75rem;
  border-radius: 0.375rem;
  font-size: 0.8125rem;
  color: var(--fog-dim, #5e7299);
  transition: all 0.18s ease;
  border: 1px solid transparent;
}
.nav-link:hover {
  background: rgba(30, 48, 80, 0.4);
  color: var(--fog, #8aa0bd);
}
.nav-link-active {
  background: rgba(46, 230, 166, 0.08);
  color: var(--pv, #2ee6a6);
  border-color: rgba(46, 230, 166, 0.3);
}
.nav-ico {
  width: 1rem;
  height: 1rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 0.7rem;
  opacity: 0.8;
}
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.18s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
