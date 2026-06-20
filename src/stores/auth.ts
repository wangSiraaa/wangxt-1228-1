import { defineStore } from 'pinia'
import { api } from '@/api/client'
import { clearToken, getStoredUser, getToken, setStoredUser, setToken } from '@/api/client'
import type { Role, User } from '@/api/types'

interface AuthState {
  token: string
  user: User | null
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: getToken(),
    user: getStoredUser(),
  }),
  getters: {
    isLoggedIn: (s) => !!s.token,
    role: (s): Role | '' => s.user?.role ?? '',
  },
  actions: {
    async login(phone: string, password: string) {
      const resp = await api.login(phone, password)
      this.token = resp.token
      this.user = resp.user
      setToken(resp.token)
      setStoredUser(resp.user)
    },
    logout() {
      this.token = ''
      this.user = null
      clearToken()
    },
  },
})
