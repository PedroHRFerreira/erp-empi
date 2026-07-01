import { describe, expect, it } from 'vitest'
import type { IReceipt, IUser } from '../../server/contracts/types'
import { buildReceiptDocument, buildReceiptInvoiceData } from './receiptDocument'

const baseUser: IUser = {
  id: 'user-1',
  name: 'Gia Bruno',
  cpf: '52998224725',
  type: 'client',
  email: 'gia@example.com',
  phone: '33987351922',
  markupPercent: 10,
  machineFeePercent: 0,
  installmentFeePercent: 0,
  address: 'Rua Cliente, 123',
  notes: '',
  createdAt: '2025-02-09T12:00:00.000Z',
  updatedAt: '2025-02-09T12:00:00.000Z'
}

const companyUser: IUser = {
  ...baseUser,
  id: 'admin-1',
  name: 'EMPI Oficina',
  type: 'admin',
  email: 'contato@empi.test',
  phone: '33999998888',
  address: 'Av. Principal, 456'
}

function makeReceipt(overrides: Partial<IReceipt> = {}): IReceipt {
  return {
    id: 'abc123ef-0000-4000-9000-123456789000',
    userId: baseUser.id,
    user: baseUser,
    vehicleModel: 'Civic',
    vehicleYear: 2020,
    vehiclePlate: 'ABC1D23',
    services: 'Higienização interna',
    laborPriceCents: 10000,
    discountCents: 0,
    productsTotalCents: 10000,
    subtotalCents: 23000,
    cardFeePercent: 5,
    cardFeeCents: 1150,
    paymentMethod: 'credit_card',
    installments: 2,
    priceCents: 24150,
    status: 'pending',
    notes: '',
    items: [
      {
        id: 'item-1',
        stockItemId: 'stock-1',
        quantity: 2,
        unitCostCents: 2500,
        unitResaleCents: 5000,
        markupPercent: 100,
        stockItem: {
          id: 'stock-1',
          name: 'Produto premium',
          description: '',
          costCents: 2500,
          markupPercent: 100,
          resalePriceCents: 5000,
          quantity: 10,
          usedQuantity: 0,
          active: true,
          createdAt: '2025-02-09T12:00:00.000Z',
          updatedAt: '2025-02-09T12:00:00.000Z'
        }
      }
    ],
    expenses: [
      {
        id: 'expense-1',
        receiptId: 'abc123ef-0000-4000-9000-123456789000',
        description: 'Deslocamento',
        category: 'Operacional',
        amountCents: 3000,
        spentAt: '2025-02-09',
        notes: '',
        createdAt: '2025-02-09T12:00:00.000Z',
        updatedAt: '2025-02-09T12:00:00.000Z'
      }
    ],
    createdAt: '2025-02-09T12:00:00.000Z',
    updatedAt: '2025-02-09T12:00:00.000Z',
    ...overrides
  }
}

describe('receipt document helpers', () => {
  it('uses the logged profile as company data', () => {
    const document = buildReceiptDocument(makeReceipt(), companyUser)

    expect(document.receiptNumber).toBe('Recibo 2025-ABC123')
    expect(document.company.name).toBe('EMPI Oficina')
    expect(document.company.cnpjLabel).toBe('46.377.137/0001-60')
    expect(document.company.lines).toContain('CNPJ: 46.377.137/0001-60')
    expect(document.company.lines).toContain('Av. Principal, 456')
    expect(document.company.lines).toContain('contato@empi.test')
    expect(document.company.lines).toContain('33999998888')
  })

  it('falls back to EMPI Autocenter without a company profile', () => {
    const document = buildReceiptDocument(makeReceipt(), null)

    expect(document.company.name).toBe('EMPI Autocenter')
    expect(document.company.initials).toBe('EA')
  })

  it('builds service, product, expense and financial rows', () => {
    const document = buildReceiptDocument(makeReceipt(), companyUser)

    expect(document.lines.map((line) => line.description)).toEqual([
      'Higienização interna',
      'Produto premium',
      'Gasto do serviço: Deslocamento'
    ])
    expect(document.lines.at(0)?.taxLabel).toBe('-')
    expect(document.summaryRows.map((row) => row.label)).toEqual(['Total'])
    expect(document.summaryRows).not.toContainEqual(expect.objectContaining({ label: 'Taxa do cartão' }))
    expect(document.summaryRows).not.toContainEqual(expect.objectContaining({ label: 'Total pendente' }))
    expect(document.summaryRows).not.toContainEqual(expect.objectContaining({ label: 'Total pago' }))
    expect(document.summaryRows).not.toContainEqual(expect.objectContaining({ label: 'Total cancelado' }))
    expect(document.summaryRows.at(0)?.valueCents).toBe(24150)
  })

  it('shows labor discount as a negative financial row', () => {
    const document = buildReceiptDocument(
      makeReceipt({
        discountCents: 5000,
        subtotalCents: 18000,
        cardFeeCents: 900,
        priceCents: 18900
      }),
      companyUser
    )

    expect(document.summaryRows.map((row) => row.label)).toEqual(['Desconto', 'Total'])
    expect(document.summaryRows.at(0)?.valueCents).toBe(-5000)
    expect(document.summaryRows.at(0)?.valueLabel).toContain('-')
    expect(document.summaryRows.at(1)?.valueCents).toBe(18900)
  })

  it('builds the invoice helper notice without replacing a fiscal document', () => {
    const document = buildReceiptInvoiceData(makeReceipt({ status: 'paid' }), companyUser)

    expect(document.title).toBe('Dados para nota fiscal - Recibo 2025-ABC123')
    expect(document.notice).toContain('não substitui uma nota fiscal')
    expect(document.summaryRows.map((row) => row.label)).toEqual(['Total'])
    expect(document.portalRows).toContainEqual({ label: 'CNPJ do prestador', value: '46.377.137/0001-60' })
    expect(document.portalRows).toContainEqual({ label: 'Município da prestação', value: 'Governador Valadares/MG' })
    expect(document.portalRows).toContainEqual({ label: 'CPF/CNPJ do tomador', value: '529.982.247-25' })
  })
})
