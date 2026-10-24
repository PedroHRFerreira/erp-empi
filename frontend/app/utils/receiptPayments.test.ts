import { describe, expect, it } from 'vitest'
import {
  calculateReceiptPayments,
  validateReceiptPayments,
  type ReceiptPaymentForm,
} from './receiptPayments'

function payment(
  paymentMethod: ReceiptPaymentForm['paymentMethod'],
  baseAmountCents = 0,
  cardFeePercent = 0,
  installments = 1,
): ReceiptPaymentForm {
  return { paymentMethod, baseAmountCents, cardFeePercent, installments }
}

describe('receipt payment calculations', () => {
  it('assigns the whole discounted subtotal to a single payment', () => {
    expect(calculateReceiptPayments([payment('pix', 1)], 8_000)).toEqual({
      payments: [{ ...payment('pix', 8_000), cardFeeCents: 0, amountCents: 8_000 }],
      baseAmountCents: 8_000,
      cardFeeCents: 0,
      amountCents: 8_000,
    })
  })

  it('adds the fee only to the card portion and preserves its installments', () => {
    const result = calculateReceiptPayments([
      payment('pix', 6_000), payment('credit_card', 0, 5, 4),
    ], 10_000)
    expect(result.payments).toEqual([
      { ...payment('pix', 6_000), cardFeeCents: 0, amountCents: 6_000 },
      { ...payment('credit_card', 4_000, 5, 4), cardFeeCents: 200, amountCents: 4_200 },
    ])
    expect(result.baseAmountCents).toBe(10_000)
    expect(result.cardFeeCents).toBe(200)
    expect(result.amountCents).toBe(10_200)
  })

  it('recalculates the remainder without changing editable inputs', () => {
    const payments = [payment('cash', 6_000), payment('pix', 99_999)]
    expect(calculateReceiptPayments(payments, 10_000).payments[1]!.baseAmountCents).toBe(4_000)
    expect(calculateReceiptPayments(payments, 8_000).payments[1]!.baseAmountCents).toBe(2_000)
    expect(calculateReceiptPayments([payment('cash', 3_000), payments[1]!], 8_000)
      .payments[1]!.baseAmountCents).toBe(5_000)
    expect(payments[1]!.baseAmountCents).toBe(99_999)
  })

  it('calculates separate credit and debit fees', () => {
    const result = calculateReceiptPayments([
      payment('credit_card', 6_000, 5, 12), payment('debit_card', 4_000, 2),
    ], 10_000)
    expect(result.payments.map(({ cardFeeCents }) => cardFeeCents)).toEqual([300, 80])
    expect(result.cardFeeCents).toBe(380)
    expect(result.amountCents).toBe(10_380)
  })

  it('preserves the backend truncation and operation order for fractional fee cents', () => {
    expect(calculateReceiptPayments([payment('credit_card', 0, 2.5)], 101).cardFeeCents).toBe(2)
    expect(calculateReceiptPayments([payment('debit_card', 0, 0.29)], 10_000).cardFeeCents)
      .toBe(Math.trunc(10_000 * (0.29 / 100)))
  })

  it('ignores fees on cash and pix', () => {
    const result = calculateReceiptPayments([payment('cash', 500, 10), payment('pix', 500, 20)], 1_000)
    expect(result.cardFeeCents).toBe(0)
    expect(result.payments.map(({ cardFeePercent }) => cardFeePercent)).toEqual([0, 0])
  })
})

describe('receipt payment validation', () => {
  it('accepts a single zero-value receipt using the subtotal', () => {
    expect(validateReceiptPayments([payment('cash', -1)], 0)).toEqual({})
  })

  it('accepts two distinct forms with a stale remainder', () => {
    expect(validateReceiptPayments([payment('pix', 6_000), payment('cash', -1)], 10_000)).toEqual({})
    expect(validateReceiptPayments([
      payment('credit_card', 6_000, 5, 12), payment('debit_card', 0, 2),
    ], 10_000)).toEqual({})
  })

  it('requires one or two payments', () => {
    expect(validateReceiptPayments([], 100).payments).toBeTruthy()
    expect(validateReceiptPayments([payment('cash'), payment('pix'), payment('debit_card')], 100)
      .payments).toBeTruthy()
  })

  it('rejects repeated and unknown methods', () => {
    expect(validateReceiptPayments([payment('pix', 50), payment('pix', 50)], 100)
      ['payments.1.paymentMethod']).toBeTruthy()
    expect(validateReceiptPayments([payment('mixed' as ReceiptPaymentForm['paymentMethod'])], 100)
      ['payments.0.paymentMethod']).toBeTruthy()
  })

  it.each([-1, 0, 100, 101, 0.5, Number.NaN, Number.POSITIVE_INFINITY])(
    'rejects an invalid first split amount %s without clamping it', (amount) => {
      const errors = validateReceiptPayments([payment('pix', amount), payment('cash')], 100)
      expect(errors['payments.0.baseAmountCents'] || errors['payments.1.baseAmountCents']).toBeTruthy()
    },
  )

  it.each([-1, 0.5, Number.NaN, Number.POSITIVE_INFINITY, Number.MAX_SAFE_INTEGER + 1])(
    'rejects an invalid subtotal %s', (subtotal) => {
      expect(validateReceiptPayments([payment('pix')], subtotal).subtotalCents).toBeTruthy()
    },
  )

  it.each([-1, Number.NaN, Number.POSITIVE_INFINITY])('rejects invalid fee %s', (fee) => {
    expect(validateReceiptPayments([payment('credit_card', 0, fee)], 100)
      ['payments.0.cardFeePercent']).toBeTruthy()
  })

  it('rejects a calculated amount outside safe integer precision', () => {
    expect(validateReceiptPayments([payment('credit_card', 0, Number.MAX_VALUE)], 100)
      ['payments.0.cardFeePercent']).toBeTruthy()
  })

  it.each([0, 13, 1.5, Number.NaN])('rejects invalid credit installments %s', (installments) => {
    expect(validateReceiptPayments([payment('credit_card', 0, 0, installments)], 100)
      ['payments.0.installments']).toBeTruthy()
  })

  it.each(['cash', 'pix', 'debit_card'] as const)('requires %s to use one installment', (method) => {
    expect(validateReceiptPayments([payment(method, 0, 0, 2)], 100)
      ['payments.0.installments']).toBeTruthy()
  })
})
