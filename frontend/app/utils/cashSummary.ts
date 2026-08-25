import type { ICashEntry, IReceipt } from '../../server/contracts/types'

type PaymentMethod = IReceipt['paymentMethod']

const paymentMethods: PaymentMethod[] = ['cash', 'pix', 'debit_card', 'credit_card']

export function calculateCashSummary(entries: ICashEntry[], openingCashCents = 0) {
  const totals = Object.fromEntries(paymentMethods.map((method) => [method, 0])) as Record<PaymentMethod, number>

  for (const entry of entries) {
    totals[entry.paymentMethod] += entry.amountCents
  }

  const dailyResultCents = Object.values(totals).reduce((sum, value) => sum + value, 0)

  return {
    totals,
    dailyResultCents,
    availableBalanceCents: openingCashCents + dailyResultCents,
  }
}
