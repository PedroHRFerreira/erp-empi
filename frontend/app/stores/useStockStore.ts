import { defineStore } from 'pinia'
import type { IPaginated, IStockItem } from '../../server/contracts/types'
import type { IStoreActionResult } from './types'

export type StockForm = {
  id?: string
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
    validate(form: StockForm): boolean {
      this.fieldErrors = {}

      if (!form.name.trim()) this.fieldErrors.name = 'Informe o produto.'
      if (form.costCents <= 0) this.fieldErrors.costCents = 'Informe um custo válido.'
      if (form.markupPercent < 0) this.fieldErrors.markupPercent = 'Margem de revenda não pode ser negativa.'
      if (form.quantity < 0) this.fieldErrors.quantity = 'Quantidade não pode ser negativa.'

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
    async load(offset = 0): Promise<IStoreActionResult<IPaginated<IStockItem>>> {
      this.setLoading(true)
      this.offset = offset
      const { data, status } = await useApiFetch<IPaginated<IStockItem>>('/stock', {
        query: { limit: this.limit, offset }
      })
      this.setLoading(false)

      if (status.value === 'error' || !data.value) {
        this.error = 'Não foi possível carregar o estoque.'
        return { status: 'error', errors: this.error, message: this.error }
      }

      const stockItems = Array.isArray(data.value.data) ? data.value.data : []

      this.items = stockItems
      this.total = data.value.total || stockItems.length
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
      if (!this.validate(form)) {
        return { status: 'error', errors: this.fieldErrors, message: this.error }
      }
      const method = form.id ? 'PUT' : 'POST'
      const url = form.id ? `/stock/${form.id}` : '/stock'
      const { status } = await useApiFetch(url, { method, body: form })

      if (status.value === 'error') {
        this.error = 'Não foi possível salvar o produto.'
        return { status: 'error', errors: this.error, message: this.error }
      }

      this.error = ''
      this.fieldErrors = {}
      const loadResult = await this.load(this.offset)

      if (loadResult.status === 'error') {
        return loadResult
      }

      return { status: 'success' }
    },
    async remove(id: string): Promise<IStoreActionResult> {
      const { status } = await useApiFetch(`/stock/${id}`, { method: 'DELETE' })

      if (status.value === 'error') {
        this.error = 'Não foi possível remover o produto.'
        return { status: 'error', errors: this.error, message: this.error }
      }

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
