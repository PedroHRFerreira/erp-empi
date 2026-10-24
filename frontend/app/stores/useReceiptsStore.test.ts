import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { buildReceiptPayload, makeReceiptForm, receiptToForm, useReceiptsStore } from './useReceiptsStore'
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
