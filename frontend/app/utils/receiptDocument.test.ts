import { describe, expect, it, vi } from 'vitest'
import type { IReceipt, IUser } from '../../server/contracts/types'
import { buildReceiptDocument, buildReceiptInvoiceData, paymentMethodLabel, receiptPaymentParts } from './receiptDocument'
import { buildReceiptPdfBytes, receiptWhatsAppMessage } from './receiptPdf'
import { printReceiptDocument } from './print'

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

  it('preserves the stored historical total even if visible items differ', () => {
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

    expect(document.summaryRows.map((row) => row.valueCents)).toEqual([-1000, 21000])
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

  it('creates an A4 receipt without truncating information or exposing internal expenses', () => {
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


function splitReceipt(overrides: Partial<IReceipt> = {}): IReceipt {
  return makeReceipt({
    paymentMethod: 'mixed',
    subtotalCents: 10000,
    priceCents: 10200,
    cardFeeCents: 200,
    status: 'paid',
    payments: [
      { paymentMethod: 'pix', baseAmountCents: 6000, cardFeePercent: 0, cardFeeCents: 0, amountCents: 6000, installments: 1 },
      { paymentMethod: 'credit_card', baseAmountCents: 4000, cardFeePercent: 5, cardFeeCents: 200, amountCents: 4200, installments: 3 },
    ],
    ...overrides,
  })
}

describe('split payment documents', () => {
  it('shows ordered amounts, credit installments and only the card fee without duplicating the total', () => {
    const document = buildReceiptDocument(splitReceipt({ expenses: [], items: [] }))
    expect(document.payment.methodLabel).toBe('Pix + Cartão de crédito (3x)')
    expect(document.payment.parts.map((part) => part.amountCents)).toEqual([6000, 4200])
    expect(document.payment.parts[0]?.cardFeeLabel).toBe('')
    expect(document.payment.parts[1]?.cardFeeLabel.replaceAll('\u00a0', ' ')).toBe('Base R$ 40,00 + taxa R$ 2,00 (5%)')
    expect(document.payment.parts[1]?.installmentLabel.replaceAll('\u00a0', ' ')).toBe('3x de R$ 14,00')
    expect(document.summaryRows.map((row) => row.valueCents)).toEqual([10200])
  })

  it('preserves stored historical fees and divides cents exactly with the final installment', () => {
    const document = buildReceiptDocument(makeReceipt({ priceCents: 10001, cardFeePercent: 5, cardFeeCents: 1, installments: 3 }))
    expect(document.payment.parts[0]?.baseAmountCents).toBe(10000)
    expect(document.payment.parts[0]?.cardFeeCents).toBe(1)
    expect(document.payment.parts[0]?.installmentLabel.replaceAll('\u00a0', ' ')).toBe('2x de R$ 33,33 + última de R$ 33,35')
    expect(document.summaryRows.at(-1)?.valueCents).toBe(10001)
  })

  it('does not describe missing mixed details as cash or parcel debit', () => {
    expect(paymentMethodLabel(splitReceipt({ payments: [] }))).toBe('Pagamento dividido (detalhes indisponíveis)')
    const debit = receiptPaymentParts(makeReceipt({ paymentMethod: 'debit_card', installments: 12, payments: [] }))[0]!
    expect(debit.installments).toBe(1)
    expect(debit.installmentLabel).toBe('')
    expect(debit.methodLabel).toBe('Cartão de débito')
  })

  it('includes both parts in WhatsApp and the printable PDF', async () => {
    const receipt = splitReceipt()
    const message = receiptWhatsAppMessage(receipt).replaceAll('\u00a0', ' ')
    expect(message).toContain('- Pix: R$ 60,00')
    expect(message).toContain('- Cartão de crédito (3x): R$ 42,00')
    expect(message).toContain('3x de R$ 14,00')
    expect(message).toContain('taxa R$ 2,00 (5%)')
    expect(message).toContain('*TOTAL PAGO: R$ 102,00*')
    const source = new TextDecoder('latin1').decode(buildReceiptPdfBytes(receipt))
    expect(source).toContain('Pix: R$ 60,00')
    expect(source).toContain('R$ 42,00')
    expect(source).toContain('3x de R$ 14,00')
    expect(source).toContain('taxa R$ 2,00')

    let printedFile: Blob | undefined
    const createObjectURL = vi.spyOn(URL, 'createObjectURL').mockImplementation((file) => {
      printedFile = file as Blob
      return 'blob:receipt-print'
    })
    const open = vi.spyOn(window, 'open').mockReturnValue({} as Window)
    const timeout = vi.spyOn(window, 'setTimeout').mockReturnValue({} as ReturnType<typeof window.setTimeout>)
    try {
      printReceiptDocument(receipt)
      expect(open).toHaveBeenCalledWith('blob:receipt-print', '_blank', 'noopener,noreferrer')
      expect(printedFile?.type).toBe('application/pdf')
      const printed = new TextDecoder('latin1').decode(await printedFile!.arrayBuffer())
      expect(printed).toContain('Pix: R$ 60,00')
      expect(printed).toContain('3x de R$ 14,00')
    } finally {
      createObjectURL.mockRestore()
      open.mockRestore()
      timeout.mockRestore()
    }
  })

  it('wraps and paginates long receipts while retaining payment details', () => {
    const source = new TextDecoder('latin1').decode(buildReceiptPdfBytes(splitReceipt({
      services: 'Serviço detalhado de manutenção '.repeat(150),
      notes: 'Observação extensa para o cliente '.repeat(100),
    })))
    expect(Number(source.match(/\/Count (\d+)/)?.[1])).toBeGreaterThan(1)
    expect(source).toContain('Pix: R$ 60,00')
    expect(source).toContain('3x de R$ 14,00')
    expect(source).toContain('taxa R$ 2,00')
    expect(source).not.toContain('...')
  })
})
