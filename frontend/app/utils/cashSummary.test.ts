import { describe, expect, it } from 'vitest'
import type { ICashEntry } from '../../server/contracts/types'
import { calculateCashSummary } from './cashSummary'

function entry(paymentMethod: ICashEntry['paymentMethod'], amountCents: number): ICashEntry {
  return {
    id: `${paymentMethod}-${amountCents}`,
    kind: 'adjustment',
    direction: amountCents < 0 ? 'out' : 'in',
    paymentMethod,
    amountCents,
    description: '',
    reason: '',
    occurredAt: '2026-08-24T12:00:00Z',
  }
}

describe('cash summary', () => {
  it('keeps the daily result while displaying accumulated non-cash balances', () => {
    const summary = calculateCashSummary([
      entry('cash', 4_500),
      entry('pix', -12_995),
    ], 127_138, {
      pixCents: 80_000,
      debitCardCents: 20_000,
      creditCardCents: 10_000,
    })

    expect(summary.totals).toEqual({
      cash: 4_500,
      pix: 80_000,
      debit_card: 20_000,
      credit_card: 10_000,
    })
    expect(summary.dailyTotals.pix).toBe(-12_995)
    expect(summary.dailyResultCents).toBe(-8_495)
    expect(summary.availableBalanceCents).toBe(241_638)
  })

  it('does not add todays non-cash entries on top of accumulated balances', () => {
    const summary = calculateCashSummary([
      entry('pix', 15_000),
      entry('debit_card', 5_000),
    ], 30_000, {
      pixCents: 40_000,
      debitCardCents: 25_000,
      creditCardCents: 10_000,
    })

    expect(summary.dailyResultCents).toBe(20_000)
    expect(summary.availableBalanceCents).toBe(105_000)
  })
})
