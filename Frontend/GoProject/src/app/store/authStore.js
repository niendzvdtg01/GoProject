import { create } from 'zustand'
import { authStorage } from '../../shared/utils/storage.js'

export const useAuthStore = create((set) => ({
  accessToken: authStorage.getAccessToken(),
  user: authStorage.getUser(),

  setSession(session) {
    authStorage.setAccessToken(session.accessToken)
    authStorage.setUser(session.user)
    set({
      accessToken: session.accessToken,
      user: session.user,
    })
  },

  clearSession() {
    authStorage.clear()
    set({
      accessToken: null,
      user: null,
    })
  },
}))
