import { defineStore } from 'pinia'
import type { IUser } from '../../server/contracts/types'

type Credentials = {
  cpf: string
  password: string
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null as IUser | null,
    loading: false,
    loaded: false,
    error: ''
  }),
  actions: {
    async login(credentials: Credentials) {
      this.loading = true
      this.error = ''
      try {
        const { data, error } = await useApiFetch<{
          user: IUser
          tokens: { accessToken: string; refreshToken: string }
        }>('/auth/login', {
          method: 'POST',
          body: credentials
        })
        if (error.value || !data.value) {
          throw new Error('invalid credentials')
        }
        useAuthToken().setTokenCookie({
          access_token: data.value.tokens.accessToken,
          token_type: 'Bearer',
          expires_in: 60 * 15
        })
        this.user = data.value.user
        this.loaded = true
      } catch {
        this.error = 'CPF ou senha invalidos.'
        throw new Error(this.error)
      } finally {
        this.loading = false
      }
    },
    async bootstrap() {
      if (this.loaded) {
        return
      }
      this.loading = true
      try {
        if (!useAuthToken().getAuthorizationHeader()) {
          this.user = null
          return
        }
        const { data, error } = await useApiFetch<IUser>('/auth/me')
        this.user = error.value ? null : data.value
      } catch {
        this.user = null
      } finally {
        this.loading = false
        this.loaded = true
      }
    },
    async logout() {
      useAuthToken().removeTokenCookie()
      this.user = null
      this.loaded = true
    }
  }
})
