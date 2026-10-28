import { defineStore } from 'pinia'
import type { IPaginated, IStockItem, IStockEntryInput, IStockSupplier, IStockSupplierDetail } from '../../server/contracts/types'
import type { IStoreActionResult } from './types'

export type StockForm = {
  id?: string
  existingStockItemId?: string
  entry?: IStockEntryInput
  name: string
  description: string
  costCents: number
  markupPercent: number
  quantity: number
}

export const useStockStore = defineStore('stock', {
  state: () => {
    return {
      items: [] as IStockItem[],
      productOptions: [] as IStockItem[],
      suppliers: [] as IStockSupplier[],
      supplierDetail: null as IStockSupplierDetail | null,
      saving: false,
      total: 0,
      limit: 10,
      offset: 0,
      searchQuery: '',
      listRequestVersion: 0,
      isLoading: false,
      loading: false,
      error: '',
      fieldErrors: {} as Record<string, string>
    }
  },
  actions: {
    validate(form: StockForm): boolean {
      this.fieldErrors = {}

      if (!form.existingStockItemId) {
        if (!form.name.trim()) this.fieldErrors.name = 'Informe o produto.'
        if (!Number.isSafeInteger(form.costCents) || form.costCents <= 0) this.fieldErrors.costCents = 'Informe um custo válido.'
        if (!Number.isFinite(form.markupPercent) || form.markupPercent < 0) this.fieldErrors.markupPercent = 'Margem de revenda não pode ser negativa.'
      } else if (!form.entry) {
        this.fieldErrors.entry = 'Informe a quantidade a adicionar.'
      }
      if (form.entry) {
        if (!form.entry.supplierName.trim()) this.fieldErrors.supplierName = 'Informe o fornecedor.'
        if (!Number.isSafeInteger(form.entry.quantity) || form.entry.quantity <= 0) this.fieldErrors.quantity = 'Informe uma quantidade inteira maior que zero.'
        if (!Number.isSafeInteger(form.entry.unitCostCents) || form.entry.unitCostCents <= 0) this.fieldErrors.unitCostCents = 'Informe um custo válido para a entrada.'
        const installments = form.entry.installments
        if (!installments.length || installments.some((part) => !Number.isSafeInteger(part.amountCents) || part.amountCents <= 0 || !part.dueDate)) {
          this.fieldErrors.installments = 'Informe parcelas válidas e suas datas de vencimento.'
        } else if (installments.reduce((sum, part) => sum + part.amountCents, 0) !== form.entry.quantity * form.entry.unitCostCents) {
          this.fieldErrors.installments = 'A soma das parcelas deve ser igual ao total da entrada.'
        }
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
    async searchProducts(q: string, supplier = ''): Promise<IStockItem[]> {
      const { data, status } = await useApiFetch<IPaginated<IStockItem>>('/stock', {
        query: { q, supplier, limit: 30, offset: 0 }
      })
      return status.value === 'error' ? [] : data.value?.data || []
    },
    async loadOptions(): Promise<boolean> {
      const products: IStockItem[] = []
      this.productOptions = []
      try {
        let total = 0
        do {
          const { data, status } = await useApiFetch<IPaginated<IStockItem>>('/stock', {
            query: { limit: 100, offset: products.length }
          })
          if (status.value === 'error' || !data.value) throw new Error('stock options unavailable')
          const page = Array.isArray(data.value.data) ? data.value.data : []
          total = data.value.total ?? page.length
          if (!page.length && products.length < total) throw new Error('incomplete stock options')
          products.push(...page)
        } while (products.length < total)
        this.productOptions = products
        this.error = ''
        return true
      } catch {
        this.error = 'Não foi possível carregar os produtos disponíveis. Atualize a página para tentar novamente.'
        return false
      }
    },
    async loadItem(id: string): Promise<IStockItem | null> {
      const { data, status } = await useApiFetch<IStockItem>(`/stock/${id}`)
      return status.value === 'error' ? null : data.value || null
    },
    async loadSuppliers(): Promise<boolean> {
      this.setLoading(true)
      const { data, status } = await useApiFetch<IStockSupplier[]>('/stock/suppliers')
      this.setLoading(false)
      if (status.value === 'error') {
        this.error = 'Não foi possível carregar os fornecedores.'
        return false
      }
      this.suppliers = Array.isArray(data.value) ? data.value : []
      this.error = ''
      return true
    },
    async loadSupplier(supplierName: string): Promise<boolean> {
      this.setLoading(true)
      const { data, status } = await useApiFetch<IStockSupplierDetail>('/stock/suppliers/detail', { query: { supplierName } })
      this.setLoading(false)
      if (status.value === 'error' || !data.value) {
        this.error = 'Não foi possível carregar o histórico do fornecedor.'
        return false
      }
      this.supplierDetail = data.value
      this.error = ''
      return true
    },
    async applySearch(query: string): Promise<IStoreActionResult<IPaginated<IStockItem>>> {
      this.searchQuery = query.trim()
      invalidateApiCache(['/stock'])
      return this.load(0)
    },
    async load(offset?: number): Promise<IStoreActionResult<IPaginated<IStockItem>>> {
      offset ??= this.offset
      if (offset !== this.offset) invalidateApiCache(['/stock'])
      this.setLoading(true)
      this.offset = offset
      const version = ++this.listRequestVersion
      const { data, status } = await useApiFetch<IPaginated<IStockItem>>('/stock', {
        query: { limit: this.limit, offset, q: this.searchQuery }
      })
      if (version !== this.listRequestVersion) return { status: 'error', message: 'Busca substituída por uma mais recente.' }
      this.setLoading(false)

      if (status.value === 'error' || !data.value) {
        this.error = 'Não foi possível carregar o estoque.'
        return { status: 'error', errors: this.error, message: this.error }
      }

      const stockItems = Array.isArray(data.value.data) ? data.value.data : []
      const total = data.value.total ?? stockItems.length
      if (offset > 0 && offset >= total) return this.load(Math.max(0, Math.floor((total - 1) / this.limit) * this.limit))

      this.items = stockItems
      this.total = total
      this.error = ''

      return {
        status: 'success',
        data: {
          ...data.value,
          data: stockItems,
          total: this.total
        }
      }
    },
    async save(form: StockForm): Promise<IStoreActionResult> {
      if (this.saving) return { status: 'error', message: 'Aguarde o salvamento em andamento.' }
      if (!this.validate(form)) {
        return { status: 'error', errors: this.fieldErrors, message: this.error }
      }
      this.saving = true
      try {
        const method = form.id ? 'PUT' : 'POST'
        const url = form.id ? `/stock/${form.id}` : '/stock'
        const { data, status } = await useApiFetch<IStockItem>(url, { method, body: form })
        if (status.value === 'error') {
          this.error = 'Não foi possível salvar. Confira a quantidade, o fornecedor e a soma das parcelas.'
          return { status: 'error', errors: this.error, message: this.error }
        }
        this.error = ''
        this.fieldErrors = {}
        invalidateApiCache(['/stock', '/stock/suppliers', '/stock/purchases', '/payables', '/payables/alerts', '/metrics/summary'])
        await this.load(this.offset)
        return { status: 'success', data: data.value || undefined }
      } finally {
        this.saving = false
      }
    },
    async remove(id: string): Promise<IStoreActionResult> {
      const { status } = await useApiFetch(`/stock/${id}`, { method: 'DELETE' })

      if (status.value === 'error') {
        this.error = 'Não foi possível remover o produto.'
        return { status: 'error', errors: this.error, message: this.error }
      }

      invalidateApiCache(['/stock', '/stock/purchases', '/payables', '/payables/alerts', '/metrics/summary'])
      const loadResult = await this.load(this.offset)

      if (loadResult.status === 'error') {
        return loadResult
      }

      return { status: 'success' }
    },
    exportCsv() {
      const rows = [
        ['Produto', 'Custo', 'Revenda', 'Quantidade', 'Usados'],
        ...this.items.map((item) => [
          item.name,
          String(item.costCents / 100),
          String(item.resalePriceCents / 100),
          String(item.quantity),
          String(item.usedQuantity)
        ])
      ]
      downloadCsv('estoque-empi.csv', rows)
    }
  }
})

function downloadCsv(filename: string, rows: string[][]) {
  const csv = rows.map((row) => row.map((cell) => `"${cell.replaceAll('"', '""')}"`).join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = filename
  link.click()
  URL.revokeObjectURL(link.href)
}
