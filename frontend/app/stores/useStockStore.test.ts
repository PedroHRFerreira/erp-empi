import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useStockStore, type StockForm } from './useStockStore'

const entry = () => ({ supplierName: 'Fornecedor', quantity: 2, unitCostCents: 1000, installments: [{ amountCents: 2000, dueDate: '2026-10-27', plannedMethod: 'pix' as const }] })
const form = (): StockForm => ({ name: 'Filtro', description: '', costCents: 1000, markupPercent: 50, quantity: 0, entry: entry() })

describe('stock atomic entry form', () => {
  beforeEach(() => { setActivePinia(createPinia()); vi.unstubAllGlobals() })

  it('validates installments against the entry total', () => {
    const store = useStockStore()
    const draft = form()
    draft.entry!.installments[0]!.amountCents = 1000
    expect(store.validate(draft)).toBe(false)
    expect(store.fieldErrors.installments).toContain('total da entrada')
  })

  it('permits explicit existing item replenishment without catalog metadata', () => {
    const store = useStockStore()
    expect(store.validate({ ...form(), name: '', costCents: 0, existingStockItemId: 'existing' })).toBe(true)
    expect(store.validate({ ...form(), existingStockItemId: 'existing', entry: undefined })).toBe(false)
  })

  it('loads every sale option without replacing the paginated stock list', async () => {
    const store = useStockStore()
    store.items = [{ id: 'current-page' }] as typeof store.items
    const first = Array.from({ length: 100 }, (_, index) => ({ id: `item-${index}` }))
    const fetch = vi.fn()
      .mockResolvedValueOnce({ data: { value: { data: first, total: 101 } }, status: { value: 'success' } })
      .mockResolvedValueOnce({ data: { value: { data: [{ id: 'older-product' }], total: 101 } }, status: { value: 'success' } })
    vi.stubGlobal('useApiFetch', fetch)
    expect(await store.loadOptions()).toBe(true)
    expect(store.productOptions).toHaveLength(101)
    expect(store.productOptions[100]?.id).toBe('older-product')
    expect(store.items[0]?.id).toBe('current-page')
    expect(fetch.mock.calls[1]?.[1].query.offset).toBe(100)
  })

  it('filters the whole catalog and preserves the search during paging and reloads', async () => {
    const store = useStockStore()
    vi.stubGlobal('invalidateApiCache', vi.fn())
    const fetch = vi.fn().mockResolvedValue({ data: { value: { data: [{ id: 'matched' }], total: 12 } }, status: { value: 'success' } })
    vi.stubGlobal('useApiFetch', fetch)
    store.offset = 20
    await store.applySearch('  Filtro de óleo  ')
    expect(store.offset).toBe(0)
    await store.load(10)
    await store.load()
    expect(fetch.mock.lastCall?.[1].query).toMatchObject({ q: 'Filtro de óleo', offset: 10 })
    await store.applySearch('')
    expect(store.searchQuery).toBe('')
    expect(store.offset).toBe(0)
  })

  it('keeps the newest search when an earlier response arrives late', async () => {
    const store = useStockStore()
    vi.stubGlobal('invalidateApiCache', vi.fn())
    let finish!: (value: unknown) => void
    vi.stubGlobal('useApiFetch', vi.fn()
      .mockReturnValueOnce(new Promise((resolve) => { finish = resolve }))
      .mockResolvedValueOnce({ data: { value: { data: [{ id: 'new' }], total: 1 } }, status: { value: 'success' } }))
    const earlier = store.applySearch('old')
    await store.applySearch('new')
    finish({ data: { value: { data: [{ id: 'old' }], total: 1 } }, status: { value: 'success' } })
    await earlier
    expect(store.items[0]?.id).toBe('new')
    expect(store.searchQuery).toBe('new')
  })

  it('saves product and entry in one request and blocks a second submit', async () => {
    const store = useStockStore()
    let complete!: (value: unknown) => void
    const pending = new Promise((resolve) => { complete = resolve })
    const fetch = vi.fn().mockReturnValue(pending)
    vi.stubGlobal('useApiFetch', fetch)
    vi.stubGlobal('invalidateApiCache', vi.fn())
    vi.spyOn(store, 'load').mockResolvedValue({ status: 'success' })
    const draft = form()
    const first = store.save(draft)
    expect(store.saving).toBe(true)
    expect((await store.save(draft)).status).toBe('error')
    expect(fetch).toHaveBeenCalledTimes(1)
    expect(fetch.mock.calls[0]![1].body.entry).toEqual(draft.entry)
    complete({ data: { value: { id: 'saved' } }, status: { value: 'success' } })
    expect((await first).status).toBe('success')
    expect(store.saving).toBe(false)
  })
})
