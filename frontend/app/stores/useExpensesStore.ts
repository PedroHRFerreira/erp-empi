import { defineStore } from 'pinia'
import type { IExpense, IExpenseForm, IFinancialSummary, IPaginated, IRealizedExpense, PayableMethod, RealizedExpenseOrigin } from '../../server/contracts/types'
import { downloadFinancialReportPdf } from '../utils/financialPdf'
import { validateInstallmentSchedule } from '../utils/purchaseInstallments'
import type { IStoreActionResult } from './types'

export const expenseCategories = [
  'combustível',
  'energia',
  'água',
  'internet',
  'aluguel',
  'salários',
  'manutenção',
  'peças/compra avulsa',
  'outros'
]

export type ExpenseForm = Omit<IExpenseForm, 'paymentMethod'> & {
  installments: Array<{ amountCents: number; dueDate: string; plannedMethod: PayableMethod }>
}

export const useExpensesStore = defineStore('expenses', {
  state: () => {
    const period = defaultPeriod()
    return {
      expenses: [] as IExpense[],
      realizedExpenses: [] as IRealizedExpense[],
      origin: 'all' as RealizedExpenseOrigin,
      summary: null as IFinancialSummary | null,
      total: 0,
      limit: 10,
      offset: 0,
      startDate: period.startDate,
      endDate: period.endDate,
      isLoading: false,
      loading: false,
      isSaving: false,
      saving: false,
      error: '',
      fieldErrors: {} as Record<string, string>
    }
  },
  actions: {
    setLoading(isLoading: boolean) {
      this.isLoading = isLoading
      this.loading = isLoading
    },
    setSaving(isSaving: boolean) {
      this.isSaving = isSaving
      this.saving = isSaving
    },
    setPeriod(startDate: string, endDate: string) {
      this.startDate = startDate
      this.endDate = endDate
    },
    clearFieldError(field: string) {
      delete this.fieldErrors[field]
      this.error = Object.values(this.fieldErrors)[0] || ''
    },
    validate(form: ExpenseForm): boolean {
      this.fieldErrors = {}

      if (!form.description.trim()) this.fieldErrors.description = 'Informe a descrição do gasto.'
      if (!form.category.trim()) this.fieldErrors.category = 'Informe a categoria.'
      if (form.amountCents <= 0) this.fieldErrors.amountCents = 'Informe um valor maior que zero.'
      if (!form.spentAt) this.fieldErrors.spentAt = 'Informe a data do gasto.'
      const schedule = validateInstallmentSchedule(form.amountCents, form.installments || [])
      if (schedule.errors.includes('empty')) {
        this.fieldErrors.installments = 'Gere ao menos uma parcela.'
      } else if (schedule.errors.includes('invalid_amount')) {
        this.fieldErrors.installments = 'Todas as parcelas devem ter valor maior que zero.'
      } else if (schedule.errors.includes('invalid_date')) {
        this.fieldErrors.installments = 'Informe uma data de vencimento válida em todas as parcelas.'
      } else if (schedule.errors.includes('total_mismatch')) {
        this.fieldErrors.installments = 'A soma das parcelas deve ser igual ao valor do gasto.'
      }
      if ((form.installments || []).some((installment) => !['boleto', 'cash', 'pix', 'debit_card', 'credit_card'].includes(installment.plannedMethod))) {
        this.fieldErrors.installments = 'Informe a forma prevista de todas as parcelas.'
      }

      this.error = Object.values(this.fieldErrors)[0] || ''
      return Object.keys(this.fieldErrors).length === 0
    },
    async load(offset = 0, forceRefresh = false): Promise<IStoreActionResult<IFinancialSummary>> {
      if (forceRefresh || offset !== this.offset) invalidateApiCache(['/expenses', '/financial/expenses', '/financial/summary'])
      this.setLoading(true)
      this.offset = offset
      const [expensesResult, realizedResult, summaryResult] = await Promise.all([
        this.loadExpenses(offset),
        this.loadRealizedExpenses(offset),
        this.loadSummary()
      ])
      this.setLoading(false)

      if (expensesResult.status === 'error') {
        return { status: 'error', errors: expensesResult.errors, message: expensesResult.message }
      }
      if (summaryResult.status === 'error') {
        return { status: 'error', errors: summaryResult.errors, message: summaryResult.message }
      }
      if (realizedResult.status === 'error') {
        return { status: 'error', errors: realizedResult.errors, message: realizedResult.message }
      }

      return { status: 'success', data: summaryResult.data }
    },
    async loadRealizedExpenses(offset = 0): Promise<IStoreActionResult<IPaginated<IRealizedExpense>>> {
      const { data, status } = await useApiFetch<IPaginated<IRealizedExpense>>('/financial/expenses', {
        query: {
          limit: this.limit,
          offset,
          origin: this.origin,
          startDate: this.startDate,
          endDate: this.endDate
        }
      })

      if (status.value === 'error' || !data.value) {
        this.error = 'Não foi possível carregar as saídas realizadas.'
        return { status: 'error', errors: this.error, message: this.error }
      }

      const rows = Array.isArray(data.value.data) ? data.value.data : []
      this.realizedExpenses = rows
      this.total = data.value.total || rows.length
      this.error = ''

      return { status: 'success', data: { ...data.value, data: rows, total: this.total } }
    },
    async loadExpenses(offset = 0): Promise<IStoreActionResult<IPaginated<IExpense>>> {
      const { data, status } = await useApiFetch<IPaginated<IExpense>>('/expenses', {
        query: {
          limit: 100,
          offset: 0,
          startDate: this.startDate,
          endDate: this.endDate
        }
      })

      if (status.value === 'error' || !data.value) {
        this.error = 'Não foi possível carregar os gastos.'
        return { status: 'error', errors: this.error, message: this.error }
      }

      const expenses = Array.isArray(data.value.data) ? data.value.data : []
      this.expenses = expenses
      this.error = ''

      return {
        status: 'success',
        data: {
          ...data.value,
          data: expenses,
          total: data.value.total || expenses.length
        }
      }
    },
    async loadSummary(): Promise<IStoreActionResult<IFinancialSummary>> {
      const { data, status } = await useApiFetch<IFinancialSummary>('/financial/summary', {
        query: {
          startDate: this.startDate,
          endDate: this.endDate
        }
      })

      if (status.value === 'error' || !data.value) {
        this.error = 'Não foi possível carregar o resumo financeiro.'
        return { status: 'error', errors: this.error, message: this.error }
      }

      const summary: IFinancialSummary = {
        ...data.value,
        expensesByCategory: Array.isArray(data.value.expensesByCategory) ? data.value.expensesByCategory : [],
        receiptCosts: Array.isArray(data.value.receiptCosts) ? data.value.receiptCosts : []
      }
      this.summary = summary
      this.error = ''

      return { status: 'success', data: summary }
    },
    async save(form: ExpenseForm): Promise<IStoreActionResult<IExpense>> {
      if (!this.validate(form)) {
        return { status: 'error', errors: this.fieldErrors, message: this.error }
      }

      this.setSaving(true)
      const { data, status } = await useApiFetch<IExpense>(form.id ? `/expenses/${form.id}` : '/expenses', {
        method: form.id ? 'PUT' : 'POST',
        body: form
      })
      this.setSaving(false)

      if (status.value === 'error' || !data.value) {
        this.error = 'Não foi possível salvar o gasto.'
        return { status: 'error', errors: this.error, message: this.error }
      }

      this.fieldErrors = {}
      this.error = ''
      invalidateApiCache(['/expenses', '/financial/expenses', '/financial/summary', '/metrics/summary', '/goals'])
      await this.load(form.id ? this.offset : 0)
      return { status: 'success', data: data.value }
    },
    async remove(id: string): Promise<IStoreActionResult> {
      const { status } = await useApiFetch(`/expenses/${id}`, { method: 'DELETE' })

      if (status.value === 'error') {
        this.error = 'Não foi possível remover o gasto.'
        return { status: 'error', errors: this.error, message: this.error }
      }

      invalidateApiCache(['/expenses', '/financial/expenses', '/financial/summary', '/metrics/summary', '/goals'])
      await this.load(this.offset)
      return { status: 'success' }
    },
    exportPdf(): IStoreActionResult {
      if (!this.summary) {
        this.error = 'Carregue o resumo financeiro antes de baixar o PDF.'
        return { status: 'error', errors: this.error, message: this.error }
      }

      downloadFinancialReportPdf(this.summary, this.realizedExpenses)
      return { status: 'success' }
    }
  }
})

function defaultPeriod() {
  const now = new Date()
  const start = new Date(now.getFullYear(), now.getMonth(), 1)
  const end = new Date(now.getFullYear(), now.getMonth() + 1, 0)

  return {
    startDate: toDateInputValue(start),
    endDate: toDateInputValue(end)
  }
}

function toDateInputValue(date: Date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}
