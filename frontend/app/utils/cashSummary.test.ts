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
  it('separates the opening fund, available balance, and daily result', () => {
    const summary = calculateCashSummary([
      entry('cash', 4_500),
      entry('pix', -12_995),
    ], 127_138)

    expect(summary.totals).toEqual({
      cash: 4_500,
      pix: -12_995,
      debit_card: 0,
      credit_card: 0,
    })
    expect(summary.dailyResultCents).toBe(-8_495)
    expect(summary.availableBalanceCents).toBe(118_643)
  })
})
