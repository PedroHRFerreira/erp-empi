import { defineStore } from 'pinia'
import type { IPayableAlert, IPayableInstallment, IStockPurchase, PayableMethod, PaymentMethod } from '../../server/contracts/types'

export interface PurchaseDraft {
  supplierName: string
  items: Array<{ stockItemId: string; quantity: number; unitCostCents: number }>
  installments: Array<{ amountCents: number; dueDate: string; plannedMethod: PayableMethod }>
}

export const usePurchasesStore = defineStore('purchases', {
  state: () => ({
    purchases: [] as IStockPurchase[],
    payables: [] as IPayableInstallment[],
    alerts: [] as IPayableAlert[],
    loading: false,
    alertsLoading: false,
    error: ''
  }),
  getters: {
    urgentCount: (state) => state.alerts.filter((alert) => alert.kind !== 'early_payment').length
  },
  actions: {
    async loadPurchases() {
      this.loading = true
      const { data, status } = await useApiFetch<IStockPurchase[]>('/stock/purchases')
      this.loading = false
      if (status.value === 'error') {
        this.error = 'Não foi possível carregar as compras.'
        return false
      }
      this.purchases = Array.isArray(data.value) ? data.value : []
      this.error = ''
      return true
    },
    async loadPayables() {
      const { data, status } = await useApiFetch<IPayableInstallment[]>('/payables')
      if (status.value === 'error') {
        this.error = 'Não foi possível carregar as contas a pagar.'
        return false
      }
      this.payables = Array.isArray(data.value) ? data.value : []
      return true
    },
    async loadAlerts(forceRefresh = false) {
      if (forceRefresh) invalidateApiCache(['/payables/alerts'])
      this.alertsLoading = true
      const { data, status } = await useApiFetch<IPayableAlert[]>('/payables/alerts')
      this.alertsLoading = false
      if (status.value === 'error') return false
      this.alerts = Array.isArray(data.value) ? data.value : []
      return true
    },
    async createPurchase(input: PurchaseDraft) {
      const { status } = await useApiFetch('/stock/purchases', { method: 'POST', body: input })
      if (status.value === 'error') {
        this.error = 'Revise os itens e confirme se a soma das parcelas corresponde ao total.'
        return false
      }
      invalidateApiCache(['/stock', '/stock/purchases', '/payables', '/payables/alerts', '/metrics/summary'])
      await Promise.all([this.loadPurchases(), this.loadPayables(), this.loadAlerts(true)])
      return true
    },
    async cancelPurchase(id: string) {
      const { status } = await useApiFetch(`/stock/purchases/${id}`, { method: 'DELETE' })
      if (status.value === 'error') {
        this.error = 'A compra só pode ser cancelada sem pagamentos e com todas as unidades disponíveis.'
        return false
      }
      invalidateApiCache(['/stock', '/stock/purchases', '/payables', '/payables/alerts', '/metrics/summary'])
      await Promise.all([this.loadPurchases(), this.loadPayables(), this.loadAlerts(true)])
      return true
    },
    async payInstallment(id: string, paymentMethod: PaymentMethod, paidAt: string) {
      const { status } = await useApiFetch(`/payables/${id}/pay`, { method: 'POST', body: { paymentMethod, paidAt } })
      if (status.value === 'error') {
        this.error = paymentMethod === 'cash'
          ? 'Abra o caixa antes de pagar uma parcela em dinheiro.'
          : 'Não foi possível quitar esta parcela. Confira a data e tente novamente.'
        return false
      }
      invalidateApiCache(['/payables', '/payables/alerts', '/cash/current', '/cash/daily-entries', '/financial/expenses', '/financial/summary'])
      await Promise.all([this.loadPayables(), this.loadAlerts(true)])
      return true
    }
  }
})
