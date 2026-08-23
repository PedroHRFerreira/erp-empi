import { describe, expect, it } from 'vitest'
import { realizedExpenseOriginLabel, realizedExpensePaymentMethodLabel } from './presentation'

describe('ExpensesTable presentation', () => {
  it('identifica a origem das saídas realizadas', () => {
    expect(realizedExpenseOriginLabel({ origin: 'operational' })).toBe('Operacional')
    expect(realizedExpenseOriginLabel({ origin: 'stock' })).toBe('Estoque')
  })

  it('traduz os meios de pagamento', () => {
    expect(realizedExpensePaymentMethodLabel('pix')).toBe('PIX')
    expect(realizedExpensePaymentMethodLabel('debit_card')).toBe('Cartão de débito')
    expect(realizedExpensePaymentMethodLabel('credit_card')).toBe('Cartão de crédito')
    expect(realizedExpensePaymentMethodLabel('cash')).toBe('Dinheiro')
    expect(realizedExpensePaymentMethodLabel('legacy')).toBe('Forma não registrada — legado')
    expect(realizedExpensePaymentMethodLabel()).toBe('Não informado')
  })
})
