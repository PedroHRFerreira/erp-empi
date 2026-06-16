import { defineStore } from 'pinia'
import type { IUser } from '../../server/contracts/types'
import { isCpf, onlyDigits } from '../utils/validation'

export const useProfileStore = defineStore('profile', {
  state: () => ({
    saving: false,
    error: ''
  }),
  actions: {
    validate(user: Partial<IUser>): string {
      if (!user.name?.trim()) return 'Informe o nome.'
      if (!isCpf(user.cpf || '')) return 'Informe um CPF valido.'
      if ((user.markupPercent || 0) < 0) return 'Markup nao pode ser negativo.'
      return ''
    },
    async update(user: IUser) {
      const error = this.validate(user)
      if (error) {
        this.error = error
        throw new Error(error)
      }
      this.saving = true
      try {
        const { data, error } = await useApiFetch<IUser>('/users/profile', {
          method: 'PUT',
          body: {
            ...user,
            cpf: onlyDigits(user.cpf),
            phone: onlyDigits(user.phone)
          }
        })
        if (error.value || !data.value) {
          throw new Error('profile update failed')
        }
        return data.value
      } finally {
        this.saving = false
      }
    }
  }
})
