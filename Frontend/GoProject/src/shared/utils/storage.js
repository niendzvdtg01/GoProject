const ACCESS_TOKEN_KEY = 'goproject.accessToken'
const USER_KEY = 'goproject.user'

export const authStorage = {
  getAccessToken() {
    return window.localStorage.getItem(ACCESS_TOKEN_KEY)
  },

  setAccessToken(token) {
    window.localStorage.setItem(ACCESS_TOKEN_KEY, token)
  },

  getUser() {
    const value = window.localStorage.getItem(USER_KEY)
    return value ? JSON.parse(value) : null
  },

  setUser(user) {
    window.localStorage.setItem(USER_KEY, JSON.stringify(user))
  },

  clear() {
    window.localStorage.removeItem(ACCESS_TOKEN_KEY)
    window.localStorage.removeItem(USER_KEY)
  },
}
