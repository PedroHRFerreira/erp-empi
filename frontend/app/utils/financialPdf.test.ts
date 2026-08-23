import { describe, expect, it } from 'vitest'
import type { IFinancialSummary, IRealizedExpense } from '../../server/contracts/types'
import { buildFinancialReportPdfBytes } from './financialPdf'

describe('financialPdf', () => {
  it('exporta saídas realizadas com origem e meio de pagamento', () => {
    const summary: IFinancialSummary = {
      startDate: '2026-08-01',
      endDate: '2026-08-31',
      paidReceiptsCount: 0,
      expensesCount: 1,
      revenuePaidCents: 30000,
      productCostCents: 0,
      cardFeesCents: 0,
      grossProfitCents: 30000,
      operationalExpensesCents: 1000,
      stockExpensesCents: 5000,
      totalRealizedExpensesCents: 6000,
      stockPaymentsCount: 1,
      operationalProfitCents: 29000,
      netProfitCents: 24000,
      netMarginPercent: 80,
      healthStatus: 'green',
      expensesByCategory: [],
      receiptCosts: []
    }
    const expenses: IRealizedExpense[] = [
      {
        id: 'stock-1',
        origin: 'stock',
        description: 'Compra do fornecedor Central',
        category: 'Estoque',
        amountCents: 5000,
        occurredAt: '2026-08-17T12:00:00Z',
        paymentMethod: 'pix',
        editable: false
      }
    ]

    const pdf = new TextDecoder('latin1').decode(buildFinancialReportPdfBytes(summary, expenses))

    expect(pdf).toContain('Saídas realizadas')
    expect(pdf).toContain('Estoque')
    expect(pdf).toContain('PIX')
    expect(pdf).toContain('Compras de estoque pagas')
  })
})
