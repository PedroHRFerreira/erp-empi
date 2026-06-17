import { defineStore } from 'pinia'
import type { IUser } from '../../server/contracts/types'
import { isCpf, onlyDigits } from '../utils/validation'
import type { IStoreActionResult } from './types'

export const useProfileStore = defineStore('profile', {
  state: () => {
    return {
      isLoading: false,
      saving: false,
      error: '',
      fieldErrors: {} as Record<string, string>
    }
  },
  actions: {
    validate(user: Partial<IUser>): boolean {
      this.fieldErrors = {}

      if (!user.name?.trim()) this.fieldErrors.name = 'Informe o nome.'
      if (!isCpf(user.cpf || '')) this.fieldErrors.cpf = 'Informe um CPF válido.'
      if (user.phone && ![10, 11].includes(onlyDigits(user.phone).length)) {
        this.fieldErrors.phone = 'Informe um telefone com DDD.'
      }
      if ((user.markupPercent || 0) < 0) this.fieldErrors.markupPercent = 'Markup não pode ser negativo.'
      if ((user.machineFeePercent || 0) < 0) this.fieldErrors.machineFeePercent = 'Juros da maquininha não pode ser negativo.'

      this.error = Object.values(this.fieldErrors)[0] || ''
      return Object.keys(this.fieldErrors).length === 0
    },
    setLoading(isLoading: boolean) {
      this.isLoading = isLoading
      this.saving = isLoading
    },
    clearFieldError(field: string) {
      delete this.fieldErrors[field]
      this.error = Object.values(this.fieldErrors)[0] || ''
    },
    async update(user: IUser): Promise<IStoreActionResult<IUser>> {
      if (!this.validate(user)) {
        return { status: 'error', errors: this.fieldErrors, message: this.error }
      }

      this.setLoading(true)
      const { data, status } = await useApiFetch<IUser>('/users/profile', {
        method: 'PUT',
        body: {
          ...user,
          cpf: onlyDigits(user.cpf),
          phone: onlyDigits(user.phone)
        }
      })
      this.setLoading(false)

      if (status.value === 'error' || !data.value) {
        this.error = 'Não foi possível atualizar o perfil.'
        return { status: 'error', errors: this.error, message: this.error }
      }

      this.error = ''
      this.fieldErrors = {}

      return { status: 'success', data: data.value }
    }
  }
})
