import { describe, expect, it } from 'vitest'
import type { IReceipt, IUser } from '../../server/contracts/types'
import { buildReceiptDocument, buildReceiptInvoiceData } from './receiptDocument'
import { buildReceiptPdfBytes, receiptWhatsAppMessage } from './receiptPdf'

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
  updatedAt: '2025-02-09T12:00:00.000Z',
}

const companyUser: IUser = {
  ...baseUser,
  id: 'admin-1',
  name: 'EMPI Oficina',
  type: 'admin',
  email: 'contato@empi.test',
  phone: '33999998888',
  address: 'Av. Principal, 456',
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
          updatedAt: '2025-02-09T12:00:00.000Z',
        },
      },
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
        updatedAt: '2025-02-09T12:00:00.000Z',
      },
    ],
    createdAt: '2025-02-09T12:00:00.000Z',
    updatedAt: '2025-02-09T12:00:00.000Z',
    ...overrides,
  }
}

describe('receipt document helpers', () => {
  it('uses the logged profile as company data', () => {
    const document = buildReceiptDocument(makeReceipt(), companyUser)

    expect(document.receiptNumber).toBe('Recibo 2025-ABC123')
    expect(document.company.name).toBe('EMPI AUTO CENTER')
    expect(document.company.cnpjLabel).toBe('46.377.137/0001-60')
    expect(document.company.lines).toContain('CNPJ: 46.377.137/0001-60')
    expect(document.company.lines).toContain('Av. Principal, 456')
    expect(document.company.lines).toContain('Tel: 33999998888')
  })

  it('uses the EMPI AUTO CENTER brand without a company profile', () => {
    const document = buildReceiptDocument(makeReceipt(), null)

    expect(document.company.name).toBe('EMPI AUTO CENTER')
    expect(document.company.initials).toBe('EA')
  })

  it('builds customer-facing service, product and financial rows without internal expenses', () => {
    const document = buildReceiptDocument(makeReceipt(), companyUser)

    expect(document.lines.map((line) => line.description)).toEqual(['Higienização interna', 'Produto premium'])
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
        priceCents: 18900,
      }),
      companyUser,
    )

    expect(document.summaryRows.map((row) => row.label)).toEqual(['Desconto', 'Total'])
    expect(document.summaryRows.at(0)?.valueCents).toBe(-5000)
    expect(document.summaryRows.at(0)?.valueLabel).toContain('-')
    expect(document.summaryRows.at(1)?.valueCents).toBe(18900)
  })

  it('recalculates the customer total from visible items when a stored total has the discount inverted', () => {
    const document = buildReceiptDocument(
      makeReceipt({
        laborPriceCents: 5000,
        productsTotalCents: 15000,
        discountCents: 1000,
        cardFeeCents: 0,
        priceCents: 21000,
        expenses: [],
        items: [
          {
            ...makeReceipt().items[0]!,
            quantity: 1,
            unitResaleCents: 15000,
          },
        ],
      }),
      companyUser,
    )

    expect(document.summaryRows.map((row) => row.valueCents)).toEqual([-1000, 19000])
  })

  it('builds a quick receipt without customer or vehicle data', () => {
    const quickReceipt = makeReceipt({
      userId: null,
      user: null,
      vehicleModel: '',
      vehicleYear: 0,
      vehiclePlate: '',
    })
    const document = buildReceiptDocument(quickReceipt, companyUser)
    const invoice = buildReceiptInvoiceData(quickReceipt, companyUser)

    expect(document.customer.name).toBe('Recibo rápido')
    expect(document.vehicle.name).toBe('Sem veículo')
    expect(document.vehicle.lines).toEqual(['Placa: -'])
    expect(invoice.portalRows).toContainEqual({
      label: 'CPF/CNPJ do tomador',
      value: 'Preencher no portal com o documento do cliente',
    })
  })

  it('builds the invoice helper notice without replacing a fiscal document', () => {
    const document = buildReceiptInvoiceData(makeReceipt({ status: 'paid' }), companyUser)

    expect(document.title).toBe('Dados para nota fiscal - Recibo 2025-ABC123')
    expect(document.notice).toContain('não substitui uma nota fiscal')
    expect(document.summaryRows.map((row) => row.label)).toEqual(['Total'])
    expect(document.portalRows).toContainEqual({
      label: 'CNPJ do prestador',
      value: '46.377.137/0001-60',
    })
    expect(document.portalRows).toContainEqual({
      label: 'Município da prestação',
      value: 'Governador Valadares/MG',
    })
    expect(document.portalRows).toContainEqual({
      label: 'CPF/CNPJ do tomador',
      value: '529.982.247-25',
    })
  })

  it('creates an A4 receipt without truncating information or exposing card fees and internal expenses', () => {
    const receipt = makeReceipt({
      services: 'Troca completa das buchas da suspensão dianteira com revisão do sistema de direção',
      notes: 'Garantia de três meses para os serviços executados conforme orientação do fabricante.',
    })
    const source = new TextDecoder('latin1').decode(buildReceiptPdfBytes(receipt, companyUser))

    expect(source).toContain('/MediaBox [0 0 595.28 841.89]')
    expect(source).toContain('/Count 1')
    expect(source).toContain('Troca completa das')
    expect(source).toContain('Garantia de três meses')
    expect(source).not.toContain('...')
    expect(source).not.toContain('Taxa do cartão')
    expect(source).not.toContain('Juros da maquininha')
    expect(source).not.toContain('Deslocamento')
  })

  it('matches the customer receipt hierarchy with service and product subtotals', () => {
    const source = new TextDecoder('latin1').decode(
      buildReceiptPdfBytes(
        makeReceipt({ discountCents: 5000, priceCents: 19150, status: 'paid' }),
        companyUser,
      ),
    )

    expect(source).toContain('DESCRIÇÃO DOS SERVIÇOS')
    expect(source).toContain('PEÇAS E MATERIAIS')
    expect(source).toContain('Subtotal Serviços:')
    expect(source).toContain('Subtotal Peças:')
    expect(source).toContain('TOTAL PAGO:')
  })

  it('builds a concise WhatsApp message with the receipt summary', () => {
    const message = receiptWhatsAppMessage(makeReceipt({ status: 'paid' }), companyUser)

    expect(message).toContain('Olá, Gia Bruno!')
    expect(message).toContain('*SERVIÇOS*')
    expect(message).toContain('*PEÇAS E MATERIAIS*')
    expect(message).toContain('*TOTAL PAGO: R$')
    expect(message).toContain('PDF segue em anexo')
    expect(message).not.toContain('Deslocamento')
  })
})
