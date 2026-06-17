import { defineStore } from 'pinia'
import type { IMetricsSummary } from '../../server/contracts/types'
import type { IStoreActionResult } from './types'

export const useMetricsStore = defineStore('metrics', {
  state: () => {
    return {
      summary: null as IMetricsSummary | null,
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
    async load(): Promise<IStoreActionResult<IMetricsSummary>> {
      this.setLoading(true)
      const { data, status } = await useApiFetch<IMetricsSummary>('/metrics/summary')
      this.setLoading(false)

      if (status.value === 'error' || !data.value) {
        this.error = 'Não foi possível carregar as métricas.'
        return { status: 'error', errors: this.error, message: this.error }
      }

      const summary: IMetricsSummary = {
        ...data.value,
        receiptsCancelled: data.value.receiptsCancelled || 0,
        revenuePendingCents: data.value.revenuePendingCents || 0,
        averageTicketPaidCents: data.value.averageTicketPaidCents || 0,
        stockItemsTotal: data.value.stockItemsTotal || 0,
        stockUnitsAvailableTotal: data.value.stockUnitsAvailableTotal || 0,
        stockUnitsUsedTotal: data.value.stockUnitsUsedTotal || 0,
        lastReceipt: data.value.lastReceipt || null,
        lastStockItem: data.value.lastStockItem || null,
        topProducts: Array.isArray(data.value.topProducts) ? data.value.topProducts : [],
        lowStockProducts: Array.isArray(data.value.lowStockProducts) ? data.value.lowStockProducts : [],
        recentClients: Array.isArray(data.value.recentClients) ? data.value.recentClients : [],
        pendingReceipts: Array.isArray(data.value.pendingReceipts) ? data.value.pendingReceipts : [],
        paidReceipts: Array.isArray(data.value.paidReceipts) ? data.value.paidReceipts : []
      }

      this.summary = summary
      this.error = ''

      return { status: 'success', data: summary }
    }
  }
})
