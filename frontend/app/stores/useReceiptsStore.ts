import { defineStore } from 'pinia'
import type { IPaginated, IReceipt } from '../../server/contracts/types'
import { formatCurrency } from '../utils/format'
import { receiptWhatsAppMessage, shareReceiptPdf } from '../utils/receiptPdf'
import { isPlate, onlyDigits } from '../utils/validation'
import type { IStoreActionResult } from './types'

export type ReceiptPaymentMethod = 'credit_card' | 'debit_card' | 'pix' | 'cash'

export type ReceiptServiceExpenseForm = {
  description: string
  category: string
  amountCents: number
  spentAt: string
  notes: string
}

export const receiptWizardSteps = [
  { key: 'client', title: 'Informações do cliente' },
  { key: 'vehicle', title: 'Informações do veículo' },
  { key: 'services', title: 'Serviços' },
  { key: 'products', title: 'Produtos usados' },
  { key: 'serviceExpenses', title: 'Gastos do serviço' },
  { key: 'finish', title: 'Finalizar recibo' }
] as const

export type ReceiptWizardStepKey = (typeof receiptWizardSteps)[number]['key']

export type ReceiptForm = {
  client: {
    name: string
    phone: string
  }
  vehicleModel: string
  vehicleYear: number
  vehiclePlate: string
  services: string
  laborPriceCents: number
  priceCents: number
  cardFeePercent: number
  machineFeePercent: number
  installmentFeePercent: number
  paymentMethod: ReceiptPaymentMethod
  installments: number
  notes: string
  items: Array<{ stockItemId: string; quantity: number }>
  serviceExpenses: ReceiptServiceExpenseForm[]
}

export function makeReceiptForm(defaultFees: { machineFeePercent?: number; installmentFeePercent?: number } = {}): ReceiptForm {
  const machineFeePercent = Number(defaultFees.machineFeePercent || 0)
  const installmentFeePercent = Number(defaultFees.installmentFeePercent || 0)

  return {
    client: { name: '', phone: '' },
    vehicleModel: '',
    vehicleYear: new Date().getFullYear(),
    vehiclePlate: '',
    services: '',
    laborPriceCents: 0,
    priceCents: 0,
    cardFeePercent: 0,
    machineFeePercent,
    installmentFeePercent,
    paymentMethod: 'cash',
    installments: 1,
    notes: '',
    items: [],
    serviceExpenses: []
  }
}

export function makeReceiptServiceExpense(): ReceiptServiceExpenseForm {
  return {
    description: '',
    category: '',
    amountCents: 0,
    spentAt: toDateInputValue(new Date()),
    notes: ''
  }
}

export const useReceiptsStore = defineStore('receipts', {
  state: () => {
    return {
      receipts: [] as IReceipt[],
      receiptOptions: [] as IReceipt[],
      receiptDraft: makeReceiptForm(),
      receiptWizardStep: 0,
      total: 0,
      limit: 10,
      offset: 0,
      isLoading: false,
      loading: false,
      optionsLoading: false,
      error: '',
      fieldErrors: {} as Record<string, string>
    }
  },
  actions: {
    resetReceiptWizard(defaultFees: { machineFeePercent?: number; installmentFeePercent?: number } = {}) {
      this.receiptDraft = makeReceiptForm(defaultFees)
      this.receiptWizardStep = 0
      this.error = ''
      this.fieldErrors = {}
    },
    applyReceiptDefaultFees(defaultFees: { machineFeePercent?: number; installmentFeePercent?: number }) {
      this.receiptDraft.machineFeePercent = Number(defaultFees.machineFeePercent || 0)
      this.receiptDraft.installmentFeePercent = Number(defaultFees.installmentFeePercent || 0)
    },
    previousReceiptStep() {
      this.receiptWizardStep = Math.max(this.receiptWizardStep - 1, 0)
    },
    nextReceiptStep() {
      this.receiptWizardStep = Math.min(this.receiptWizardStep + 1, receiptWizardSteps.length - 1)
    },
    validateReceiptStep(stepIndex?: number): boolean {
      const step = receiptWizardSteps[stepIndex ?? this.receiptWizardStep]?.key
      const form = this.receiptDraft

      this.clearReceiptStepErrors(step)

      if (step === 'client') {
        this.validateClientFields(form)
      }
      if (step === 'vehicle') {
        this.validateVehicleFields(form)
      }
      if (step === 'services') {
        this.validateServiceFields(form)
      }
      if (step === 'products') {
        this.validateProductFields(form)
      }
      if (step === 'serviceExpenses') {
        this.validateServiceExpenseFields(form)
      }
      if (step === 'finish') {
        return this.validate(form)
      }

      this.error = Object.values(this.fieldErrors)[0] || ''
      return !this.error
    },
    validate(form: ReceiptForm): boolean {
      this.fieldErrors = {}

      this.validateClientFields(form)
      this.validateVehicleFields(form)
      this.validateServiceFields(form)
      this.validateProductFields(form)
      this.validateServiceExpenseFields(form)

      this.error = Object.values(this.fieldErrors)[0] || ''
      return Object.keys(this.fieldErrors).length === 0
    },
    validateClientFields(form: ReceiptForm) {
      const clientPhone = onlyDigits(form.client.phone)

      if (!form.client.name.trim()) this.fieldErrors['client.name'] = 'Informe o nome do cliente.'
      if (!clientPhone) {
        this.fieldErrors['client.phone'] = 'Informe o telefone do cliente.'
      }
      if (clientPhone && ![10, 11].includes(clientPhone.length)) {
        this.fieldErrors['client.phone'] = 'Informe um telefone com DDD.'
      }
    },
    validateVehicleFields(form: ReceiptForm) {
      if (!form.vehicleModel.trim()) this.fieldErrors.vehicleModel = 'Informe o veículo.'
      if (form.vehicleYear < 1950) this.fieldErrors.vehicleYear = 'Informe um ano válido.'
      if (!isPlate(form.vehiclePlate)) this.fieldErrors.vehiclePlate = 'Informe uma placa válida.'
    },
    validateServiceFields(form: ReceiptForm) {
      if (!form.services.trim()) this.fieldErrors.services = 'Informe os serviços.'
      if (form.laborPriceCents < 0) this.fieldErrors.laborPriceCents = 'Informe um valor válido para a mão de obra.'
      if (!['credit_card', 'debit_card', 'pix', 'cash'].includes(form.paymentMethod)) {
        this.fieldErrors.paymentMethod = 'Informe a forma de pagamento.'
      }
      if (form.paymentMethod === 'credit_card' && (form.installments < 1 || form.installments > 12)) {
        this.fieldErrors.installments = 'Informe entre 1 e 12 parcelas.'
      }
      if ((form.paymentMethod === 'credit_card' || form.paymentMethod === 'debit_card') && form.cardFeePercent < 0) {
        this.fieldErrors.cardFeePercent = 'Informe um juros válido.'
      }
    },
    validateProductFields(form: ReceiptForm) {
      if (form.items.some((item) => !item.stockItemId || item.quantity <= 0)) {
        this.fieldErrors.items = 'Revise os itens utilizados.'
      }
    },
    validateServiceExpenseFields(form: ReceiptForm) {
      if (
        form.serviceExpenses.some((expense) => {
          return (
            !expense.description.trim() ||
            !expense.category.trim() ||
            expense.amountCents <= 0 ||
            !expense.spentAt
          )
        })
      ) {
        this.fieldErrors.serviceExpenses = 'Revise os gastos do serviço.'
      }
    },
    clearReceiptStepErrors(step: ReceiptWizardStepKey | undefined) {
      const fieldsByStep: Record<ReceiptWizardStepKey, string[]> = {
        client: ['client.name', 'client.phone'],
        vehicle: ['vehicleModel', 'vehicleYear', 'vehiclePlate'],
        services: ['services', 'laborPriceCents', 'paymentMethod', 'installments', 'cardFeePercent'],
        products: ['items'],
        serviceExpenses: ['serviceExpenses'],
        finish: []
      }

      if (!step) return
      for (const field of fieldsByStep[step]) {
        delete this.fieldErrors[field]
      }
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
    async loadOptions(): Promise<IStoreActionResult<IReceipt[]>> {
      this.optionsLoading = true
      const { data, status } = await useApiFetch<IPaginated<IReceipt>>('/receipts', {
        query: { limit: 100, offset: 0 }
      })
      this.optionsLoading = false

      if (status.value === 'error' || !data.value) {
        this.error = 'Não foi possível carregar os recibos.'
        return { status: 'error', errors: this.error, message: this.error }
      }

      this.receiptOptions = Array.isArray(data.value.data) ? data.value.data : []
      return { status: 'success', data: this.receiptOptions }
    },
    async create(form: ReceiptForm): Promise<IStoreActionResult> {
      if (!this.validate(form)) {
        return { status: 'error', errors: this.fieldErrors, message: this.error }
      }
      const { error, status } = await useApiFetch('/receipts', {
        method: 'POST',
        body: {
          client: {
            name: form.client.name.trim(),
            phone: onlyDigits(form.client.phone)
          },
          vehicleModel: form.vehicleModel.trim(),
          vehicleYear: form.vehicleYear,
          vehiclePlate: form.vehiclePlate.toUpperCase(),
          services: form.services.trim(),
          laborPriceCents: form.laborPriceCents,
          priceCents: form.priceCents,
          cardFeePercent: ['credit_card', 'debit_card'].includes(form.paymentMethod) ? form.cardFeePercent : 0,
          paymentMethod: form.paymentMethod,
          installments: form.paymentMethod === 'credit_card' ? form.installments : 1,
          notes: form.notes.trim(),
          items: form.items,
          serviceExpenses: form.serviceExpenses
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
      const whatsappUrl = buildReceiptWhatsAppUrl(receipt, text)
      const whatsappWindow = window.open(whatsappUrl, '_blank')

      if (whatsappWindow) {
        whatsappWindow.opener = null
      } else {
        window.location.href = whatsappUrl
      }

      try {
        await shareReceiptPdf(receipt)
      } catch {
        this.error = 'Não foi possível gerar o PDF do recibo.'
      }
    },
    async copyInstagramText(receipt: IReceipt) {
      const text = `Recibo EMPI: ${receipt.services} - ${formatCurrency(receipt.priceCents)}`
      await navigator.clipboard.writeText(text)
    }
  }
})

function toDateInputValue(date: Date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function buildReceiptWhatsAppUrl(receipt: IReceipt, text: string) {
  const phone = whatsappPhoneNumber(receipt.user?.phone || '')
  const encodedText = encodeURIComponent(text)

  if (!phone) {
    return `https://wa.me/?text=${encodedText}`
  }

  return `https://wa.me/${phone}?text=${encodedText}`
}

function whatsappPhoneNumber(phone: string) {
  const digits = onlyDigits(phone)

  if ([12, 13].includes(digits.length) && digits.startsWith('55')) {
    return digits
  }
  if ([10, 11].includes(digits.length)) {
    return `55${digits}`
  }
  return ''
}

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
