import { defineStore } from 'pinia'
import type { IPaginated, IReceipt } from '../../server/contracts/types'
import { formatCurrency } from '../utils/format'
import { receiptWhatsAppMessage, shareReceiptPdf } from '../utils/receiptPdf'
import { isCpf, isPlate, onlyDigits } from '../utils/validation'
import type { IStoreActionResult } from './types'

export type ReceiptPaymentMethod = 'credit_card' | 'debit_card' | 'pix' | 'cash'

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
  laborPriceCents: number
  priceCents: number
  paymentMethod: ReceiptPaymentMethod
  installments: number
  notes: string
  items: Array<{ stockItemId: string; quantity: number }>
}

export const useReceiptsStore = defineStore('receipts', {
  state: () => {
    return {
      receipts: [] as IReceipt[],
      total: 0,
      limit: 10,
      offset: 0,
      isLoading: false,
      loading: false,
      error: '',
      fieldErrors: {} as Record<string, string>
    }
  },
  actions: {
    validate(form: ReceiptForm): boolean {
      this.fieldErrors = {}

      const clientCpf = onlyDigits(form.client.cpf)
      const clientPhone = onlyDigits(form.client.phone)

      if (!form.client.name.trim()) this.fieldErrors['client.name'] = 'Informe o nome do cliente.'
      if (clientCpf && !isCpf(clientCpf)) this.fieldErrors['client.cpf'] = 'Informe um CPF válido.'
      if (!clientCpf && !clientPhone) {
        this.fieldErrors['client.phone'] = 'Informe um telefone quando o CPF não for preenchido.'
      }
      if (clientPhone && ![10, 11].includes(clientPhone.length)) {
        this.fieldErrors['client.phone'] = 'Informe um telefone com DDD.'
      }
      if (!form.vehicleModel.trim()) this.fieldErrors.vehicleModel = 'Informe o veículo.'
      if (form.vehicleYear < 1950) this.fieldErrors.vehicleYear = 'Informe um ano válido.'
      if (!isPlate(form.vehiclePlate)) this.fieldErrors.vehiclePlate = 'Informe uma placa válida.'
      if (!form.services.trim()) this.fieldErrors.services = 'Informe os serviços.'
      if (form.laborPriceCents < 0) this.fieldErrors.laborPriceCents = 'Informe um valor válido para a mão de obra.'
      if (!['credit_card', 'debit_card', 'pix', 'cash'].includes(form.paymentMethod)) {
        this.fieldErrors.paymentMethod = 'Informe a forma de pagamento.'
      }
      if (form.paymentMethod === 'credit_card' && (form.installments < 1 || form.installments > 12)) {
        this.fieldErrors.installments = 'Informe entre 1 e 12 parcelas.'
      }
      if (form.items.some((item) => !item.stockItemId || item.quantity <= 0)) {
        this.fieldErrors.items = 'Revise os itens utilizados.'
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
    async load(offset = 0, receiptStatus = ''): Promise<IStoreActionResult<IPaginated<IReceipt>>> {
      this.setLoading(true)
      this.offset = offset
      const { data, status } = await useApiFetch<IPaginated<IReceipt>>('/receipts', {
        query: { limit: this.limit, offset, status: receiptStatus }
      })
      this.setLoading(false)

      if (status.value === 'error' || !data.value) {
        this.error = 'Não foi possível carregar os recibos.'
        return { status: 'error', errors: this.error, message: this.error }
      }

      const receipts = Array.isArray(data.value.data) ? data.value.data : []

      this.receipts = receipts
      this.total = data.value.total || receipts.length
      this.error = ''

      return {
        status: 'success',
        data: {
          ...data.value,
          data: receipts,
          total: this.total
        }
      }
    },
    async create(form: ReceiptForm): Promise<IStoreActionResult> {
      if (!this.validate(form)) {
        return { status: 'error', errors: this.fieldErrors, message: this.error }
      }
      const { error, status } = await useApiFetch('/receipts', {
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

      if (status.value === 'error') {
        this.error = getReceiptErrorMessage(error.value?.data?.message, 'Não foi possível salvar o recibo.')
        return { status: 'error', errors: this.error, message: this.error }
      }

      this.error = ''
      this.fieldErrors = {}
      const loadResult = await this.load(0)

      if (loadResult.status === 'error') {
        return loadResult
      }

      return { status: 'success' }
    },
    async markPaid(id: string): Promise<IStoreActionResult> {
      const { error, status } = await useApiFetch(`/receipts/${id}/pay`, { method: 'POST' })

      if (status.value === 'error') {
        this.error = getReceiptErrorMessage(error.value?.data?.message, 'Não foi possível marcar o recibo como pago.')
        return { status: 'error', errors: this.error, message: this.error }
      }

      const loadResult = await this.load(this.offset)

      if (loadResult.status === 'error') {
        return loadResult
      }

      return { status: 'success' }
    },
    async shareWhatsApp(receipt: IReceipt) {
      const text = receiptWhatsAppMessage(receipt)
      const shared = await shareReceiptPdf(receipt)

      if (!shared) {
        window.open(`https://wa.me/?text=${encodeURIComponent(text)}`, '_blank', 'noopener,noreferrer')
      }
    },
    async copyInstagramText(receipt: IReceipt) {
      const text = `Recibo EMPI: ${receipt.services} - ${formatCurrency(receipt.priceCents)}`
      await navigator.clipboard.writeText(text)
    }
  }
})

function getReceiptErrorMessage(message: string | undefined, fallback: string) {
  if (message === 'insufficient stock' || message === 'Estoque insuficiente.') {
    return 'Estoque insuficiente para os produtos selecionados.'
  }
  if (message === 'reserved stock' || message === 'Produto reservado em outro recibo pendente.') {
    return 'Produto indisponível: ele já está reservado em outro recibo pendente.'
  }
  if (message === 'invalid input' || message === 'Dados inválidos.') {
    return 'Revise os dados do recibo antes de salvar.'
  }
  return message || fallback
}
