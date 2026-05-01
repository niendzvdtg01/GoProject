import { create } from 'zustand'

export const useUiStore = create((set) => ({
  isSidebarOpen: true,
  theme: 'light',

  toggleSidebar() {
    set((state) => ({ isSidebarOpen: !state.isSidebarOpen }))
  },

  setTheme(theme) {
    set({ theme })
  },
}))
