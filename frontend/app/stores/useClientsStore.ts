import { defineStore } from 'pinia'
import type { IClientDetail, IPaginated, IUser } from '../../server/contracts/types'
import type { IStoreActionResult } from './types'

export const useClientsStore = defineStore('clients', {
  state: () => {
    return {
      clients: [] as IUser[],
      detail: null as IClientDetail | null,
      total: 0,
      limit: 10,
      offset: 0,
      isLoading: false,
      loading: false,
      error: ''
    }
  },
  actions: {
    setLoading(isLoading: boolean) {
      this.isLoading = isLoading
      this.loading = isLoading
    },
    async load(offset = 0): Promise<IStoreActionResult<IPaginated<IUser>>> {
      this.setLoading(true)
      this.offset = offset
      const { data, status } = await useApiFetch<IPaginated<IUser>>('/users/clients', {
        query: { limit: this.limit, offset }
      })
      this.setLoading(false)

      if (status.value === 'error' || !data.value) {
        this.error = 'Não foi possível carregar os clientes.'
        return { status: 'error', errors: this.error, message: this.error }
      }

      const clients = Array.isArray(data.value.data) ? data.value.data : []

      this.clients = clients
      this.total = data.value.total || clients.length
      this.error = ''

      return {
        status: 'success',
        data: {
          ...data.value,
          data: clients,
          total: this.total
        }
      }
    },
    async loadDetail(id: string): Promise<IStoreActionResult<IClientDetail>> {
      this.setLoading(true)
      this.detail = null
      this.error = ''
      const { data, status } = await useApiFetch<IClientDetail>(`/users/clients/${id}/detail`)
      this.setLoading(false)

      if (status.value === 'error' || !data.value) {
        this.error = 'Não foi possível carregar os detalhes do cliente.'
        return { status: 'error', errors: this.error, message: this.error }
      }

      this.detail = {
        client: data.value.client,
        receipts: Array.isArray(data.value.receipts) ? data.value.receipts : []
      }
      this.error = ''

      return { status: 'success', data: this.detail }
    },
    async remove(id: string): Promise<IStoreActionResult> {
      const { status } = await useApiFetch(`/users/clients/${id}`, { method: 'DELETE' })

      if (status.value === 'error') {
        this.error = 'Não foi possível remover o cliente.'
        return { status: 'error', errors: this.error, message: this.error }
      }

      this.error = ''
      const loadResult = await this.load(this.offset)

      if (loadResult.status === 'error') {
        return loadResult
      }

      return { status: 'success' }
    }
  }
})
