import { describe, expect, it } from 'vitest'
import {
  addMonthsStable,
  divideCents,
  generateInstallmentSchedule,
  validateInstallmentSchedule,
} from './purchaseInstallments'

describe('purchase installment helpers', () => {
  it('keeps the exact total when cents do not divide evenly', () => {
    const installments = divideCents(10_000, 3)

    expect(installments).toEqual([3_334, 3_333, 3_333])
    expect(installments.reduce((sum, amount) => sum + amount, 0)).toBe(10_000)
  })

  it('redistributes the total across the requested remaining rows', () => {
    expect(divideCents(19_999, 2)).toEqual([10_000, 9_999])
  })

  it('always returns at least one installment', () => {
    expect(divideCents(2_500, 0)).toEqual([2_500])
  })

  it('handles non-finite inputs without producing an invalid array', () => {
    expect(divideCents(Number.NaN, Number.NaN)).toEqual([0])
  })

  it('keeps the original day while generating monthly due dates', () => {
    expect(generateInstallmentSchedule(10_000, 3, '2026-08-15')).toEqual([
      { amountCents: 3_334, dueDate: '2026-08-15' },
      { amountCents: 3_333, dueDate: '2026-09-15' },
      { amountCents: 3_333, dueDate: '2026-10-15' },
    ])
  })

  it('clamps month-end dates without drifting subsequent installments', () => {
    expect(generateInstallmentSchedule(3_000, 3, '2026-01-31')).toEqual([
      { amountCents: 1_000, dueDate: '2026-01-31' },
      { amountCents: 1_000, dueDate: '2026-02-28' },
      { amountCents: 1_000, dueDate: '2026-03-31' },
    ])
    expect(addMonthsStable('2024-01-31', 1)).toBe('2024-02-29')
  })

  it('rejects invalid first due dates when generating a schedule', () => {
    expect(generateInstallmentSchedule(1_000, 2, '2026-02-30')).toEqual([])
    expect(addMonthsStable('not-a-date', 1)).toBe('')
  })

  it('validates exact totals, positive integer cents, and ISO dates', () => {
    expect(validateInstallmentSchedule(10_000, [
      { amountCents: 5_000, dueDate: '2026-08-31' },
      { amountCents: 5_000, dueDate: '2026-09-30' },
    ])).toEqual({ valid: true, errors: [] })

    expect(validateInstallmentSchedule(10_000, [
      { amountCents: 0, dueDate: '2026-02-30' },
      { amountCents: 9_999, dueDate: '31/03/2026' },
    ])).toEqual({
      valid: false,
      errors: ['invalid_amount', 'invalid_date', 'total_mismatch'],
    })
  })

  it('rejects an empty schedule', () => {
    expect(validateInstallmentSchedule(0, [])).toEqual({
      valid: false,
      errors: ['empty'],
    })
  })
})
