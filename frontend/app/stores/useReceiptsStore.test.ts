import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { buildReceiptPayload, makeReceiptForm, receiptToForm, useReceiptsStore, useRecoveryReceiptsStore } from './useReceiptsStore'
import type { IReceipt } from '../../server/contracts/types'

describe('useReceiptsStore receipt creation', () => {
  beforeEach(() => setActivePinia(createPinia()))
  afterEach(() => vi.unstubAllGlobals())

  it('keeps the split receipt pending and explains the cash requirement when payment fails', async () => {
    const store = useReceiptsStore()
    store.receipts = [{
      id: 'split-cash', status: 'pending', paymentMethod: 'mixed',
      payments: [{ paymentMethod: 'pix' }, { paymentMethod: 'cash' }],
    } as unknown as IReceipt]
    vi.stubGlobal('useApiFetch', vi.fn().mockResolvedValue({
      error: { value: { data: { message: 'Conflito ao processar a solicitação.' } } },
      status: { value: 'error' },
    }))

    const result = await store.markPaid('split-cash')

    expect(result.status).toBe('error')
    expect(store.error).toContain('caixa está aberto')
    expect(store.receipts[0]!.status).toBe('pending')
  })

  it.each([false, true])('blocks a concurrent submission when quick is %s', async (quick) => {
    const store = useReceiptsStore()
    store.creating = true

    const result = await store.create(makeReceiptForm({}, quick))

    expect(result).toEqual({
      status: 'error',
      errors: 'O recibo já está sendo salvo.',
      message: 'O recibo já está sendo salvo.',
    })
  })

  it('validates split payments at finalization after products and discounts', () => {
    const store = useReceiptsStore()
    const form = makeReceiptForm({}, true)
    form.services = 'Troca de óleo'
    form.laborPriceCents = 10_000
    form.productsTotalCents = 5000
    form.discountCents = 1000
    form.payments = [
      { paymentMethod: 'pix', baseAmountCents: 6000, cardFeePercent: 0, installments: 1 },
      { paymentMethod: 'credit_card', baseAmountCents: 0, cardFeePercent: 5, installments: 4 },
    ]
    store.setReceiptDraft(form)
    expect(store.validateReceiptStep(0)).toBe(true)
    expect(store.validate(form)).toBe(true)
    form.payments[0]!.baseAmountCents = 14_001
    expect(store.validate(form)).toBe(false)
    expect(store.fieldErrors['payments.1.baseAmountCents']).toBeTruthy()
  })

  it('sends derived remaining amount and card installments with fees only on card', () => {
    const form = makeReceiptForm({}, true)
    form.services = ' Serviço '
    form.laborPriceCents = 11_000
    form.discountCents = 1000
    form.payments = [
      { paymentMethod: 'pix', baseAmountCents: 6000, cardFeePercent: 9, installments: 1 },
      { paymentMethod: 'credit_card', baseAmountCents: 111, cardFeePercent: 5, installments: 4 },
    ]
    const payload = buildReceiptPayload(form)
    expect(payload.payments).toEqual([
      { paymentMethod: 'pix', baseAmountCents: 6000, cardFeePercent: 0, installments: 1 },
      { paymentMethod: 'credit_card', baseAmountCents: 4000, cardFeePercent: 5, installments: 4 },
    ])
    expect(payload).toMatchObject({ priceCents: 10_200, paymentMethod: 'mixed', cardFeePercent: 0, installments: 1 })
  })

  it('preserves both allocations and custom installment settings when editing', () => {
    const receipt = {
      paymentMethod: 'mixed', priceCents: 10_200, cardFeeCents: 200, cardFeePercent: 0, items: [],
      payments: [
        { paymentMethod: 'pix', baseAmountCents: 6000, cardFeePercent: 0, cardFeeCents: 0, amountCents: 6000, installments: 1 },
        { paymentMethod: 'credit_card', baseAmountCents: 4000, cardFeePercent: 5, cardFeeCents: 200, amountCents: 4200, installments: 4 },
      ],
    } as unknown as IReceipt
    const form = receiptToForm(receipt)
    expect(form.payments).toEqual([
      { paymentMethod: 'pix', baseAmountCents: 6000, cardFeePercent: 0, installments: 1 },
      { paymentMethod: 'credit_card', baseAmountCents: 4000, cardFeePercent: 5, installments: 4 },
    ])
    form.payments[1]!.installments = 12
    expect(receipt.payments![1]!.installments).toBe(4)
  })

  it('hydrates an old receipt using the historical exact total and fee', () => {
    const receipt = { paymentMethod: 'credit_card', priceCents: 10_321, cardFeeCents: 321, cardFeePercent: 3.2, installments: 3, items: [] } as unknown as IReceipt
    expect(receiptToForm(receipt).payments).toEqual([
      { paymentMethod: 'credit_card', baseAmountCents: 10_000, cardFeePercent: 3.2, installments: 3 },
    ])
  })
})

describe('receipt filters', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.stubGlobal('invalidateApiCache', vi.fn())
  })
  afterEach(() => vi.unstubAllGlobals())

  it('isolates recovery filters and rows from the receipts screen', () => {
    const receipts = useReceiptsStore()
    const recovery = useRecoveryReceiptsStore()
    receipts.filters.q = 'Cliente escolhido'
    receipts.filters.paymentMethod = 'pix'
    recovery.filters.status = 'cancelled'
    recovery.receipts = [{ id: 'recovery-only' } as IReceipt]
    expect(receipts.filters.status).toBe('')
    expect(recovery.filters.q).toBe('')
    expect(recovery.filters.paymentMethod).toBe('')
    expect(receipts.receipts).toEqual([])
  })

  it('returns to the last valid filtered page after the final row leaves a page', async () => {
    const fetch = vi.fn()
      .mockResolvedValueOnce({ data: { value: { data: [], total: 10 } }, status: { value: 'success' } })
      .mockResolvedValueOnce({ data: { value: { data: [{ id: 'remaining' }], total: 10 } }, status: { value: 'success' } })
    vi.stubGlobal('useApiFetch', fetch)
    const store = useReceiptsStore()
    store.filters.status = 'pending'
    await store.load(10)
    expect(store.offset).toBe(0)
    expect(store.receipts[0]?.id).toBe('remaining')
    expect(fetch.mock.calls[1]?.[1].query.status).toBe('pending')
  })

  it('keeps all filters across pagination and mutation reloads and clears explicitly', async () => {
    const fetch = vi.fn().mockResolvedValue({ data: { value: { data: [], total: 20 } }, status: { value: 'success' } })
    vi.stubGlobal('useApiFetch', fetch)
    const store = useReceiptsStore()
    store.offset = 20
    await store.applyFilters({ q: ' Alice ', status: 'pending', startDate: '2026-10-26', endDate: '2026-10-26', paymentMethod: 'pix', type: 'normal' })
    expect(store.offset).toBe(0)
    expect(fetch.mock.lastCall?.[1].query).toMatchObject({ q: 'Alice', status: 'pending', paymentMethod: 'pix', type: 'normal', offset: 0 })
    await store.load(10)
    await store.markPaid('receipt-id')
    expect(fetch.mock.lastCall?.[1].query).toMatchObject({ q: 'Alice', startDate: '2026-10-26', endDate: '2026-10-26', paymentMethod: 'pix', status: 'pending', type: 'normal', offset: 10 })
    await store.load()
    expect(store.offset).toBe(10)
    await store.clearFilters()
    expect(store.filters).toEqual({ q: '', startDate: '', endDate: '', status: '', paymentMethod: '', type: '' })
    expect(store.offset).toBe(0)
  })

  it('retains legacy status argument compatibility', async () => {
    vi.stubGlobal('useApiFetch', vi.fn().mockResolvedValue({ data: { value: { data: [], total: 0 } }, status: { value: 'success' } }))
    const store = useReceiptsStore()
    await store.load(0, 'pending')
    expect(store.filters.status).toBe('pending')
    expect(store.currentStatusFilter).toBe('pending')
  })

  it('does not let an old request replace a new filter result', async () => {
    let resolveFirst!: (value: unknown) => void
    const first = new Promise((resolve) => { resolveFirst = resolve })
    vi.stubGlobal('useApiFetch', vi.fn().mockReturnValueOnce(first).mockResolvedValueOnce({ data: { value: { data: [{ id: 'new' }], total: 1 } }, status: { value: 'success' } }))
    const store = useReceiptsStore()
    const pending = store.applyFilters({ q: 'old' })
    await store.applyFilters({ q: 'new' })
    resolveFirst({ data: { value: { data: [{ id: 'old' }], total: 10 } }, status: { value: 'success' } })
    await pending
    expect(store.receipts[0]!.id).toBe('new')
    expect(store.total).toBe(1)
    expect(store.filters.q).toBe('new')
    expect(store.loading).toBe(false)
  })

  it('surfaces invalid filters and ends loading on request failure', async () => {
    vi.stubGlobal('useApiFetch', vi.fn().mockResolvedValue({ data: { value: null }, status: { value: 'error' }, error: { value: { data: { message: 'Dados inválidos.' } } } }))
    const store = useReceiptsStore()
    expect((await store.applyFilters({ startDate: 'invalid' })).status).toBe('error')
    expect(store.error).toContain('período')
    expect(store.loading).toBe(false)
  })
})
