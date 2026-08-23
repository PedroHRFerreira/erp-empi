import type { IRealizedExpense, PaymentMethod } from '../../../../server/contracts/types'

export function realizedExpenseOriginLabel(expense: Pick<IRealizedExpense, 'origin'>) {
  return expense.origin === 'stock' ? 'Estoque' : 'Operacional'
}

export function realizedExpensePaymentMethodLabel(method?: PaymentMethod) {
  if (method === 'pix') return 'PIX'
  if (method === 'debit_card') return 'Cartão de débito'
  if (method === 'credit_card') return 'Cartão de crédito'
  if (method === 'cash') return 'Dinheiro'
  if (method === 'legacy') return 'Forma não registrada — legado'
  return 'Não informado'
}
