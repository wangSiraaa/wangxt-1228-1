import { defineStore } from 'pinia'

export interface Toast {
  id: number
  type: 'success' | 'error' | 'info'
  message: string
}

let seq = 0

export const useUiStore = defineStore('ui', {
  state: () => ({ toasts: [] as Toast[] }),
  actions: {
    push(message: string, type: Toast['type'] = 'info') {
      const id = ++seq
      this.toasts.push({ id, type, message })
      setTimeout(() => this.dismiss(id), 3600)
    },
    success(message: string) {
      this.push(message, 'success')
    },
    error(message: string) {
      this.push(message, 'error')
    },
    dismiss(id: number) {
      this.toasts = this.toasts.filter((t) => t.id !== id)
    },
  },
})
