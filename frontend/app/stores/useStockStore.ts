import { defineStore } from 'pinia'
import type { IPaginated, IStockItem } from '../../server/contracts/types'

export type StockForm = {
  id?: string
  name: string
  description: string
  costCents: number
  markupPercent: number
  quantity: number
}

export const useStockStore = defineStore('stock', {
  state: () => ({
    items: [] as IStockItem[],
    total: 0,
    limit: 10,
    offset: 0,
    loading: false,
    error: ''
  }),
  actions: {
    validate(form: StockForm): string {
      if (!form.name.trim()) return 'Informe o produto.'
      if (form.costCents <= 0) return 'Informe um custo valido.'
      if (form.markupPercent < 0) return 'Markup nao pode ser negativo.'
      if (form.quantity < 0) return 'Quantidade nao pode ser negativa.'
      return ''
    },
    async load(offset = 0) {
      this.loading = true
      this.offset = offset
      try {
        const { data, error } = await useApiFetch<IPaginated<IStockItem>>('/stock', {
          query: { limit: this.limit, offset }
        })
        if (error.value || !data.value) {
          throw new Error('stock load failed')
        }
        this.items = data.value.data
        this.total = data.value.total
      } finally {
        this.loading = false
      }
    },
    async save(form: StockForm) {
      const error = this.validate(form)
      if (error) {
        this.error = error
        throw new Error(error)
      }
      const method = form.id ? 'PUT' : 'POST'
      const url = form.id ? `/stock/${form.id}` : '/stock'
      const { error: requestError } = await useApiFetch(url, { method, body: form })
      if (requestError.value) {
        throw new Error('stock save failed')
      }
      await this.load(this.offset)
    },
    async remove(id: string) {
      const { error } = await useApiFetch(`/stock/${id}`, { method: 'DELETE' })
      if (error.value) {
        throw new Error('stock delete failed')
      }
      await this.load(this.offset)
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
