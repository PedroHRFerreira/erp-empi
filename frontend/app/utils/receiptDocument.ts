import type { IReceipt, IUser } from '../../server/contracts/types'
import { formatCurrency } from './format'

const FALLBACK_COMPANY_NAME = 'EMPI Autocenter'
export const COMPANY_CNPJ = '46377137000160'
export const NFSE_SERVICE_CITY = 'Governador Valadares/MG'
export const NFSE_TAX_CODE = '31.01.03 - Serviços técnicos em mecânica e congêneres.'
const EMPTY_VALUE = '-'

export interface IReceiptDocumentCompany {
  name: string
  initials: string
  cnpj: string
  cnpjLabel: string
  address: string
  email: string
  phone: string
  lines: string[]
}

export interface IReceiptDocumentParty {
  name: string
  lines: string[]
}

export interface IReceiptDocumentLine {
  description: string
  quantity: string
  priceCents: number
  priceLabel: string
  taxLabel: string
  totalCents: number
  totalLabel: string
}

export interface IReceiptDocumentMoneyRow {
  label: string
  valueCents: number
  valueLabel: string
  strong?: boolean
}

export interface IReceiptInvoicePortalRow {
  label: string
  value: string
}

export interface IReceiptDocumentPayment {
  dateLabel: string
  methodLabel: string
  amountLabel: string
  statusLabel: string
}

export interface IReceiptDocument {
  receiptNumber: string
  issuedAtLabel: string
  company: IReceiptDocumentCompany
  customer: IReceiptDocumentParty
  vehicle: IReceiptDocumentParty
  lines: IReceiptDocumentLine[]
  summaryRows: IReceiptDocumentMoneyRow[]
  payment: IReceiptDocumentPayment
  thankYouTitle: string
  thankYouMessage: string
  legalNotice: string
}

export interface IReceiptInvoiceDataDocument {
  receipt: IReceiptDocument
  title: string
  notice: string
  portalRows: IReceiptInvoicePortalRow[]
  providerRows: IReceiptDocumentParty[]
  customerRows: IReceiptDocumentParty[]
  serviceRows: IReceiptDocumentLine[]
  summaryRows: IReceiptDocumentMoneyRow[]
}

export function buildReceiptDocument(receipt: IReceipt, company: IUser | null = null): IReceiptDocument {
  const lines = buildReceiptLines(receipt)
  const subtotalCents = receipt.subtotalCents || sumLineTotals(lines)
  const cardFeeCents = receipt.cardFeeCents || 0
  const totalCents = receipt.priceCents || subtotalCents + cardFeeCents

  return {
    receiptNumber: receiptNumber(receipt),
    issuedAtLabel: formatDate(receipt.createdAt),
    company: receiptCompany(company),
    customer: receiptCustomer(receipt.user),
    vehicle: receiptVehicle(receipt),
    lines,
    summaryRows: buildSummaryRows(receipt, subtotalCents, cardFeeCents, totalCents),
    payment: {
      dateLabel: formatDate(receipt.paidAt || receipt.createdAt),
      methodLabel: paymentMethodLabel(receipt),
      amountLabel: receipt.status === 'paid' ? formatCurrency(totalCents) : statusLabel(receipt.status),
      statusLabel: statusLabel(receipt.status)
    },
    thankYouTitle: 'Obrigado',
    thankYouMessage: 'Tenha um ótimo dia!',
    legalNotice: 'Este recibo não é uma nota fiscal.'
  }
}

export function buildReceiptInvoiceData(receipt: IReceipt, company: IUser | null = null): IReceiptInvoiceDataDocument {
  const document = buildReceiptDocument(receipt, company)

  return {
    receipt: document,
    title: `Dados para nota fiscal - ${document.receiptNumber}`,
    notice: 'Dados auxiliares para emissão de nota fiscal. Este documento não substitui uma nota fiscal.',
    portalRows: buildInvoicePortalRows(receipt, document),
    providerRows: [
      {
        name: 'Prestador',
        lines: partyLines(document.company.name, document.company.lines)
      }
    ],
    customerRows: [
      {
        name: 'Tomador',
        lines: partyLines(document.customer.name, document.customer.lines)
      },
      {
        name: 'Veículo',
        lines: partyLines(document.vehicle.name, document.vehicle.lines)
      }
    ],
    serviceRows: document.lines,
    summaryRows: document.summaryRows
  }
}

export function paymentMethodLabel(receipt: Pick<IReceipt, 'paymentMethod' | 'installments'>) {
  if (receipt.paymentMethod === 'credit_card') return `Cartão de crédito (${receipt.installments || 1}x)`
  if (receipt.paymentMethod === 'debit_card') return 'Cartão de débito'
  if (receipt.paymentMethod === 'pix') return 'Pix'
  return 'Dinheiro'
}

export function statusLabel(status: IReceipt['status']) {
  if (status === 'paid') return 'Pago'
  if (status === 'cancelled') return 'Cancelado'
  return 'Pendente'
}

function buildReceiptLines(receipt: IReceipt): IReceiptDocumentLine[] {
  const lines: IReceiptDocumentLine[] = [
    moneyLine(receipt.services || 'Serviços', '1', receipt.laborPriceCents || 0, receipt.laborPriceCents || 0)
  ]

  for (const item of Array.isArray(receipt.items) ? receipt.items : []) {
    const quantity = Number(item.quantity || 0)
    const unitPriceCents = Number(item.unitResaleCents || item.stockItem?.resalePriceCents || 0)
    lines.push(
      moneyLine(item.stockItem?.name || item.stockItemId || 'Produto', String(quantity), unitPriceCents, unitPriceCents * quantity)
    )
  }

  for (const expense of Array.isArray(receipt.expenses) ? receipt.expenses : []) {
    lines.push(moneyLine(`Gasto do serviço: ${expense.description}`, '1', expense.amountCents || 0, expense.amountCents || 0))
  }

  return lines
}

function buildSummaryRows(receipt: IReceipt, subtotalCents: number, cardFeeCents: number, totalCents: number): IReceiptDocumentMoneyRow[] {
  const rows: IReceiptDocumentMoneyRow[] = [moneyRow('Subtotal', subtotalCents)]

  if (cardFeeCents > 0) {
    rows.push(moneyRow('Taxa do cartão', cardFeeCents))
  }

  rows.push(moneyRow('Total', totalCents, true))

  if (receipt.status === 'paid') {
    rows.push(moneyRow('Total pago', totalCents))
  } else if (receipt.status === 'cancelled') {
    rows.push(moneyRow('Total cancelado', totalCents))
  } else {
    rows.push(moneyRow('Total pendente', totalCents))
  }

  return rows
}

function moneyLine(description: string, quantity: string, priceCents: number, totalCents: number): IReceiptDocumentLine {
  return {
    description: normalizeText(description) || EMPTY_VALUE,
    quantity,
    priceCents,
    priceLabel: formatCurrency(priceCents),
    taxLabel: EMPTY_VALUE,
    totalCents,
    totalLabel: formatCurrency(totalCents)
  }
}

function moneyRow(label: string, valueCents: number, strong = false): IReceiptDocumentMoneyRow {
  return {
    label,
    valueCents,
    valueLabel: formatCurrency(valueCents),
    strong
  }
}

function receiptCompany(company: IUser | null): IReceiptDocumentCompany {
  const name = normalizeText(company?.name) || FALLBACK_COMPANY_NAME
  const cnpjLabel = formatCnpj(COMPANY_CNPJ)
  const address = normalizeText(company?.address)
  const email = normalizeText(company?.email)
  const phone = normalizeText(company?.phone)

  return {
    name,
    initials: initials(name),
    cnpj: COMPANY_CNPJ,
    cnpjLabel,
    address,
    email,
    phone,
    lines: [`CNPJ: ${cnpjLabel}`, address, email, phone].filter(Boolean)
  }
}

function receiptCustomer(user: IUser): IReceiptDocumentParty {
  const name = normalizeText(user?.name) || 'Cliente'
  return {
    name,
    lines: [normalizeText(user?.address), normalizeText(user?.email), normalizeText(user?.phone)].filter(Boolean)
  }
}

function receiptVehicle(receipt: IReceipt): IReceiptDocumentParty {
  return {
    name: `${receipt.vehicleModel || EMPTY_VALUE} ${receipt.vehicleYear || ''}`.trim(),
    lines: [`Placa: ${receipt.vehiclePlate || EMPTY_VALUE}`]
  }
}

function receiptNumber(receipt: IReceipt) {
  const year = receiptYear(receipt.createdAt)
  const id = (receipt.id || '').replaceAll('-', '').slice(0, 6).toUpperCase() || '000000'
  return `Recibo ${year}-${id}`
}

function buildInvoicePortalRows(receipt: IReceipt, document: IReceiptDocument): IReceiptInvoicePortalRow[] {
  const totalRow = document.summaryRows.find((row) => row.label === 'Total')

  return [
    { label: 'Portal', value: 'NFS-e Nacional - Emissão completa' },
    { label: 'CNPJ do prestador', value: document.company.cnpjLabel },
    { label: 'Data de competência', value: document.issuedAtLabel },
    { label: 'País do tomador', value: 'Brasil' },
    { label: 'CPF/CNPJ do tomador', value: receipt.user?.cpf ? formatCpfCnpj(receipt.user.cpf) : 'Preencher no portal com o documento do cliente' },
    { label: 'Município da prestação', value: NFSE_SERVICE_CITY },
    { label: 'Código de Tributação Nacional', value: NFSE_TAX_CODE },
    { label: 'Caso de imunidade/exportação/não incidência do ISSQN', value: 'Não' },
    { label: 'Descrição do serviço', value: invoiceServiceDescription(receipt) },
    { label: 'Valor do serviço prestado', value: totalRow?.valueLabel || formatCurrency(receipt.priceCents) },
    { label: 'Regime especial de tributação', value: 'Nenhum' },
    { label: 'Exigibilidade do ISSQN suspensa', value: 'Não' },
    { label: 'Retenção do ISSQN', value: 'Não' },
    { label: 'Benefício municipal', value: 'Não' },
    { label: 'Tributos aproximados', value: 'Não informar nenhum valor estimado' }
  ]
}

function invoiceServiceDescription(receipt: IReceipt) {
  const services = normalizeText(receipt.services) || 'Prestação de serviço automotivo'
  const hasProducts = Array.isArray(receipt.items) && receipt.items.length > 0

  if (!hasProducts) return services
  return `${services}. Valor apresentado como prestação de serviço, com materiais/insumos embutidos quando aplicável, sem destaque de peças.`
}

function receiptYear(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return new Date().getFullYear()
  return date.getFullYear()
}

function formatDate(value: string) {
  if (!value) return EMPTY_VALUE
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return EMPTY_VALUE

  return new Intl.DateTimeFormat('pt-BR', {
    dateStyle: 'medium'
  }).format(date)
}

function initials(value: string) {
  const chars = value
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join('')
    .toUpperCase()

  return chars || 'EA'
}

function formatCnpj(value: string) {
  const digits = value.replace(/\D/g, '').slice(0, 14)
  return digits
    .replace(/^(\d{2})(\d)/, '$1.$2')
    .replace(/^(\d{2})\.(\d{3})(\d)/, '$1.$2.$3')
    .replace(/\.(\d{3})(\d)/, '.$1/$2')
    .replace(/(\d{4})(\d)/, '$1-$2')
}

function formatCpfCnpj(value: string) {
  const digits = value.replace(/\D/g, '')
  if (digits.length > 11) return formatCnpj(digits)
  return digits
    .slice(0, 11)
    .replace(/^(\d{3})(\d)/, '$1.$2')
    .replace(/^(\d{3})\.(\d{3})(\d)/, '$1.$2.$3')
    .replace(/\.(\d{3})(\d)/, '.$1-$2')
}

function partyLines(name: string, lines: string[]) {
  return [name, ...lines].filter(Boolean)
}

function normalizeText(value: string | undefined | null) {
  return String(value || '').trim()
}

function sumLineTotals(lines: IReceiptDocumentLine[]) {
  return lines.reduce((total, line) => total + line.totalCents, 0)
}
