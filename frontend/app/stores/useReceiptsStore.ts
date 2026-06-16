import { defineStore } from 'pinia'
import type { IPaginated, IReceipt } from '../../server/contracts/types'
import { formatCurrency } from '../utils/format'
import { isCpf, isPlate, onlyDigits } from '../utils/validation'

export type ReceiptForm = {
  client: {
    name: string
    cpf: string
    phone: string
    email: string
  }
  vehicleModel: string
  vehicleYear: number
  vehiclePlate: string
  services: string
  priceCents: number
  notes: string
  items: Array<{ stockItemId: string; quantity: number }>
}

export const useReceiptsStore = defineStore('receipts', {
  state: () => ({
    receipts: [] as IReceipt[],
    total: 0,
    limit: 10,
    offset: 0,
    loading: false,
    error: ''
  }),
  actions: {
    validate(form: ReceiptForm): string {
      if (!form.client.name.trim()) return 'Informe o nome do cliente.'
      if (!isCpf(form.client.cpf)) return 'Informe um CPF valido.'
      if (!form.vehicleModel.trim()) return 'Informe o veiculo.'
      if (form.vehicleYear < 1950) return 'Informe um ano valido.'
      if (!isPlate(form.vehiclePlate)) return 'Informe uma placa valida.'
      if (!form.services.trim()) return 'Informe os servicos.'
      if (form.priceCents <= 0) return 'Informe o valor do recibo.'
      if (form.items.some((item) => !item.stockItemId || item.quantity <= 0)) return 'Revise os itens utilizados.'
      return ''
    },
    async load(offset = 0, status = '') {
      this.loading = true
      this.offset = offset
      try {
        const { data, error } = await useApiFetch<IPaginated<IReceipt>>('/receipts', {
          query: { limit: this.limit, offset, status }
        })
        if (error.value || !data.value) {
          throw new Error('receipts load failed')
        }
        this.receipts = data.value.data
        this.total = data.value.total
      } finally {
        this.loading = false
      }
    },
    async create(form: ReceiptForm) {
      const error = this.validate(form)
      if (error) {
        this.error = error
        throw new Error(error)
      }
      const { error: requestError } = await useApiFetch('/receipts', {
        method: 'POST',
        body: {
          ...form,
          client: {
            ...form.client,
            cpf: onlyDigits(form.client.cpf),
            phone: onlyDigits(form.client.phone)
          },
          vehiclePlate: form.vehiclePlate.toUpperCase()
        }
      })
      if (requestError.value) {
        throw new Error('receipt create failed')
      }
      await this.load(0)
    },
    async markPaid(id: string) {
      const { error } = await useApiFetch(`/receipts/${id}/pay`, { method: 'POST' })
      if (error.value) {
        throw new Error('receipt payment failed')
      }
      await this.load(this.offset)
    },
    shareWhatsApp(receipt: IReceipt) {
      const text = [
        `Recibo EMPI Autocenter`,
        `Cliente: ${receipt.user.name}`,
        `Veiculo: ${receipt.vehicleModel} ${receipt.vehicleYear}`,
        `Placa: ${receipt.vehiclePlate}`,
        `Servicos: ${receipt.services}`,
        `Valor: ${formatCurrency(receipt.priceCents)}`
      ].join('\n')
      window.open(`https://wa.me/?text=${encodeURIComponent(text)}`, '_blank', 'noopener,noreferrer')
    },
    async copyInstagramText(receipt: IReceipt) {
      const text = `Recibo EMPI: ${receipt.services} - ${formatCurrency(receipt.priceCents)}`
      await navigator.clipboard.writeText(text)
    }
  }
})
