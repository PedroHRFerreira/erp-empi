import type { ICashBalances, ICashEntry, IReceipt } from '../../server/contracts/types'

type PaymentMethod = IReceipt['paymentMethod']

const paymentMethods: PaymentMethod[] = ['cash', 'pix', 'debit_card', 'credit_card']

const emptyBalances: ICashBalances = {
  pixCents: 0,
  debitCardCents: 0,
  creditCardCents: 0,
}

export function calculateCashSummary(
  entries: ICashEntry[],
  openingCashCents = 0,
  balances: ICashBalances = emptyBalances,
) {
  const dailyTotals = Object.fromEntries(paymentMethods.map((method) => [method, 0])) as Record<PaymentMethod, number>

  for (const entry of entries) {
    dailyTotals[entry.paymentMethod] += entry.amountCents
  }

  const dailyResultCents = Object.values(dailyTotals).reduce((sum, value) => sum + value, 0)
  const totals: Record<PaymentMethod, number> = {
    cash: dailyTotals.cash,
    pix: balances.pixCents,
    debit_card: balances.debitCardCents,
    credit_card: balances.creditCardCents,
  }
  const nonCashBalanceCents = balances.pixCents + balances.debitCardCents + balances.creditCardCents

  return {
    totals,
    dailyTotals,
    dailyResultCents,
    availableBalanceCents: openingCashCents + dailyTotals.cash + nonCashBalanceCents,
  }
}
