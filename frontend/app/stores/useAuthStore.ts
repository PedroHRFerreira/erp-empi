import { defineStore } from 'pinia'
import type { IUser } from '../../server/contracts/types'
import { isCpf } from '../utils/validation'
import type { IStoreActionResult } from './types'

type Credentials = {
  cpf: string
  password: string
}

export const useAuthStore = defineStore('auth', {
  state: () => {
    return {
      user: null as IUser | null,
      isLoading: false,
      loading: false,
      loaded: false,
      error: '',
      fieldErrors: {} as Record<string, string>
    }
  },
  actions: {
    validate(credentials: Credentials): boolean {
      this.fieldErrors = {}

      if (!isCpf(credentials.cpf)) {
        this.fieldErrors.cpf = 'Informe um CPF válido.'
      }

      if (!credentials.password.trim()) {
        this.fieldErrors.password = 'Informe a senha.'
      }

      this.error = Object.values(this.fieldErrors)[0] || ''
      return Object.keys(this.fieldErrors).length === 0
    },
    setLoading(isLoading: boolean) {
      this.isLoading = isLoading
      this.loading = isLoading
    },
    clearFieldError(field: string) {
      delete this.fieldErrors[field]
      this.error = Object.values(this.fieldErrors)[0] || ''
    },
    async login(credentials: Credentials): Promise<IStoreActionResult<IUser>> {
      if (!this.validate(credentials)) {
        return { status: 'error', errors: this.fieldErrors, message: this.error }
      }

      this.setLoading(true)
      this.error = ''
      const { data, status } = await useApiFetch<{
        user: IUser
        tokens: { accessToken: string; expiresIn: number }
      }>('/auth/login', {
        method: 'POST',
        body: credentials
      })
      this.setLoading(false)

      if (status.value === 'error' || !data.value) {
        this.error = 'CPF ou senha inválidos.'
        this.fieldErrors.password = this.error
        return { status: 'error', errors: this.fieldErrors, message: this.error }
      }

      useAuthToken().setTokenCookie({
        access_token: data.value.tokens.accessToken,
        token_type: 'Bearer',
        expires_in: data.value.tokens.expiresIn
      })
      this.user = data.value.user
      this.loaded = true

      return { status: 'success', data: this.user }
    },
    async bootstrap(): Promise<IStoreActionResult<IUser | null>> {
      if (this.loaded) {
        return { status: 'success', data: this.user }
      }

      this.setLoading(true)

      if (!useAuthToken().getAuthorizationHeader()) {
        this.user = null
        this.setLoading(false)
        this.loaded = true
        return { status: 'success', data: this.user }
      }

      const { data, status } = await useApiFetch<IUser>('/auth/me')
      this.user = status.value === 'error' ? null : data.value
      this.setLoading(false)
      this.loaded = true

      if (status.value === 'error') {
        return { status: 'error', message: 'Não foi possível carregar o usuário.', data: this.user }
      }

      return { status: 'success', data: this.user }
    },
    async logout(): Promise<IStoreActionResult> {
      useAuthToken().removeTokenCookie()
      this.user = null
      this.loaded = true
      this.fieldErrors = {}
      this.error = ''

      return { status: 'success' }
    }
  }
})
