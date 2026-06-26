import type { IReceipt, IUser } from '../../server/contracts/types'
import { buildReceiptDocument } from './receiptDocument'

type PdfText = {
  kind: 'text'
  text: string
  x: number
  y: number
  size?: number
  font?: 'F1' | 'F2'
}

type PdfRule = {
  kind: 'rule'
  x1: number
  x2: number
  y: number
}

type PdfElement = PdfText | PdfRule

export function receiptWhatsAppMessage(receipt: IReceipt, company: IUser | null = null) {
  const document = buildReceiptDocument(receipt, company)
  const itemLines = document.lines.map((line) => `- ${line.quantity}x ${line.description}: ${line.totalLabel}`)

  return [
    document.receiptNumber,
    document.company.name,
    ...document.company.lines,
    `Cliente: ${document.customer.name}`,
    `Veículo: ${document.vehicle.name}`,
    ...document.vehicle.lines,
    'Itens:',
    ...itemLines,
    `Pagamento: ${document.payment.methodLabel}`,
    ...document.summaryRows.map((row) => `${row.label}: ${row.valueLabel}`),
    document.legalNotice
  ].join('\n')
}

export async function shareReceiptPdf(receipt: IReceipt, company: IUser | null = null) {
  const file = buildReceiptPdfFile(receipt, company)
  const text = receiptWhatsAppMessage(receipt, company)
  const document = buildReceiptDocument(receipt, company)
  const shareData = {
    title: `${document.receiptNumber} - ${receipt.user.name}`,
    text,
    files: [file]
  }

  if (navigator.canShare?.(shareData)) {
    try {
      await navigator.share(shareData)
      return true
    } catch {
      downloadReceiptPdf(file)
      return false
    }
  }

  downloadReceiptPdf(file)
  return false
}

export function buildReceiptPdfFile(receipt: IReceipt, company: IUser | null = null) {
  const document = buildReceiptDocument(receipt, company)
  const bytes = buildReceiptPdfBytes(receipt, company)
  const filename = `${document.receiptNumber.toLowerCase().replace(/\s+/g, '-')}.pdf`
  return new File([bytes], filename, { type: 'application/pdf' })
}

function buildReceiptPdfBytes(receipt: IReceipt, company: IUser | null = null) {
  const document = buildReceiptDocument(receipt, company)
  const elements: PdfElement[] = []
  let y = 790

  addLine(elements, document.company.name, 48, y, 16, 'F2')
  addLine(elements, document.receiptNumber, 330, y, 24, 'F2')
  y -= 18
  for (const line of document.company.lines.slice(0, 3)) {
    addLine(elements, truncate(line, 42), 48, y, 9)
    y -= 12
  }
  addLine(elements, document.issuedAtLabel, 430, 772, 10)
  y -= 10
  addRule(elements, y)
  y -= 30

  addLine(elements, 'Dados do cliente', 48, y, 12, 'F2')
  y -= 18
  addLine(elements, document.customer.name, 48, y, 10, 'F2')
  addLine(elements, document.vehicle.name, 330, y, 10, 'F2')
  y -= 14
  addLine(elements, truncate(document.customer.lines.join(' | ') || '-', 46), 48, y, 9)
  addLine(elements, truncate(document.vehicle.lines.join(' | '), 32), 330, y, 9)
  y -= 34

  addLine(elements, 'Itens', 48, y, 12, 'F2')
  y -= 20
  addLine(elements, 'Item', 48, y, 10, 'F2')
  addLine(elements, 'Qtd.', 270, y, 10, 'F2')
  addLine(elements, 'Preco', 350, y, 10, 'F2')
  addLine(elements, 'Total', 485, y, 10, 'F2')
  y -= 8
  addRule(elements, y)
  y -= 16

  for (const line of document.lines) {
    addLine(elements, truncate(line.description, 30), 48, y)
    addLine(elements, line.quantity, 275, y)
    addLine(elements, line.priceLabel, 350, y)
    addLine(elements, line.totalLabel, 470, y, 10, 'F2')
    y -= 16
  }

  y -= 4
  addRule(elements, y)
  y -= 26

  for (const row of document.summaryRows) {
    addLine(elements, row.label, 335, y, row.strong ? 13 : 10, row.strong ? 'F2' : 'F1')
    addLine(elements, row.valueLabel, 470, y, row.strong ? 13 : 10, row.strong ? 'F2' : 'F1')
    y -= row.strong ? 20 : 16
  }

  y -= 18
  addRule(elements, y)
  y -= 28
  addLine(elements, 'Detalhes do pagamento', 48, y, 12, 'F2')
  y -= 18
  addLine(elements, document.payment.dateLabel, 48, y)
  addLine(elements, document.payment.methodLabel, 180, y)
  addLine(elements, document.payment.amountLabel, 330, y, 10, 'F2')
  y -= 44
  addLine(elements, document.thankYouTitle, 48, y, 12, 'F2')
  y -= 16
  addLine(elements, document.thankYouMessage, 48, y)
  addLine(elements, document.legalNotice, 48, Math.max(y - 46, 54), 10, 'F2')

  const stream = elements.map((element) => drawElement(element)).join('\n')
  return createPdf(stream)
}

function addLine(elements: PdfElement[], text: string, x: number, y: number, size = 10, font: 'F1' | 'F2' = 'F1') {
  elements.push({ kind: 'text', text, x, y, size, font })
}

function addRule(elements: PdfElement[], y: number, x1 = 48, x2 = 547) {
  elements.push({ kind: 'rule', x1, x2, y })
}

function drawElement(element: PdfElement) {
  if (element.kind === 'rule') return `0.6 w ${element.x1} ${element.y} m ${element.x2} ${element.y} l S`
  return drawText(element)
}

function drawText(line: PdfText) {
  return `BT /${line.font || 'F1'} ${line.size || 10} Tf 1 0 0 1 ${line.x} ${line.y} Tm (${escapePdfString(line.text)}) Tj ET`
}

function createPdf(contentStream: string) {
  const objects = [
    '<< /Type /Catalog /Pages 2 0 R >>',
    '<< /Type /Pages /Kids [3 0 R] /Count 1 >>',
    '<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R /F2 5 0 R >> >> /Contents 6 0 R >>',
    '<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>',
    '<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>',
    `<< /Length ${latin1Length(contentStream)} >>\nstream\n${contentStream}\nendstream`
  ]

  const header = '%PDF-1.4\n'
  const chunks: string[] = [header]
  const offsets: number[] = [0]
  let length = latin1Length(header)

  objects.forEach((object, index) => {
    offsets.push(length)
    const chunk = `${index + 1} 0 obj\n${object}\nendobj\n`
    chunks.push(chunk)
    length += latin1Length(chunk)
  })

  const xrefOffset = length
  const xref = [
    `xref\n0 ${objects.length + 1}\n`,
    '0000000000 65535 f \n',
    ...offsets.slice(1).map((offset) => `${String(offset).padStart(10, '0')} 00000 n \n`),
    `trailer\n<< /Size ${objects.length + 1} /Root 1 0 R >>\nstartxref\n${xrefOffset}\n%%EOF`
  ].join('')

  chunks.push(xref)
  return latin1Bytes(chunks.join(''))
}

function downloadReceiptPdf(file: File) {
  const href = URL.createObjectURL(file)
  const link = document.createElement('a')
  link.href = href
  link.download = file.name
  link.click()
  URL.revokeObjectURL(href)
}

function truncate(value: string, maxLength: number) {
  return value.length > maxLength ? `${value.slice(0, maxLength - 1)}...` : value
}

function escapePdfString(value: string) {
  return value.replaceAll('\\', '\\\\').replaceAll('(', '\\(').replaceAll(')', '\\)')
}

function latin1Length(value: string) {
  return latin1Bytes(value).length
}

function latin1Bytes(value: string) {
  const bytes = new Uint8Array(value.length)
  for (let index = 0; index < value.length; index += 1) {
    const code = value.charCodeAt(index)
    bytes[index] = code <= 255 ? code : 63
  }
  return bytes
}
