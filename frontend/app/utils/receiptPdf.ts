import type { IReceipt } from '../../server/contracts/types'
import { formatCpf, formatCurrency, formatDateTime } from './format'

type PdfLine = {
  text: string
  x: number
  y: number
  size?: number
  font?: 'F1' | 'F2'
}

export function receiptWhatsAppMessage(receipt: IReceipt) {
  const productLines = receipt.items.length
    ? ['Produtos utilizados:', ...receipt.items.map((item) => `- ${item.quantity}x ${item.stockItem?.name || item.stockItemId}`)]
    : ['Produtos utilizados: nenhum produto vinculado.']

  return [
    'Recibo EMPI Autocenter',
    `Cliente: ${receipt.user.name}`,
    `Veículo: ${receipt.vehicleModel} ${receipt.vehicleYear}`,
    `Placa: ${receipt.vehiclePlate}`,
    `Serviços: ${receipt.services}`,
    ...productLines,
    `Pagamento: ${paymentMethodLabel(receipt)}`,
    `Valor total: ${formatCurrency(receipt.priceCents)}`,
    'Este recibo não é uma nota fiscal.'
  ].join('\n')
}

export async function shareReceiptPdf(receipt: IReceipt) {
  const file = buildReceiptPdfFile(receipt)
  const text = receiptWhatsAppMessage(receipt)
  const shareData = {
    title: `Recibo EMPI Autocenter - ${receipt.user.name}`,
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

export function buildReceiptPdfFile(receipt: IReceipt) {
  const bytes = buildReceiptPdfBytes(receipt)
  const filename = `recibo-empi-${receipt.id.slice(0, 8)}.pdf`
  return new File([bytes], filename, { type: 'application/pdf' })
}

function buildReceiptPdfBytes(receipt: IReceipt) {
  const lines: PdfLine[] = []
  let y = 790

  addLine(lines, 'EMPI Autocenter', 48, y, 18, 'F2')
  addLine(lines, 'Recibo de serviço', 395, y, 18, 'F2')
  y -= 20
  addLine(lines, `Emissão: ${formatDateTime(receipt.createdAt)}`, 395, y, 9)
  y -= 30

  addLine(lines, 'Cliente', 48, y, 12, 'F2')
  addLine(lines, 'Veículo', 320, y, 12, 'F2')
  y -= 18
  addLine(lines, receipt.user.name, 48, y, 11, 'F2')
  addLine(lines, `${receipt.vehicleModel} ${receipt.vehicleYear}`, 320, y, 11, 'F2')
  y -= 15
  addLine(lines, `CPF: ${receipt.user.cpf ? formatCpf(receipt.user.cpf) : '-'}`, 48, y)
  addLine(lines, `Placa: ${receipt.vehiclePlate}`, 320, y)
  y -= 15
  addLine(lines, `Telefone: ${receipt.user.phone || '-'}`, 48, y)
  addLine(lines, `Status: ${statusLabel(receipt.status)}`, 320, y)
  y -= 32

  addLine(lines, 'Serviços', 48, y, 12, 'F2')
  y -= 16
  for (const line of wrapText(receipt.services, 88)) {
    addLine(lines, line, 48, y)
    y -= 14
  }
  if (receipt.notes) {
    y -= 4
    for (const line of wrapText(`Observações: ${receipt.notes}`, 88)) {
      addLine(lines, line, 48, y, 9)
      y -= 12
    }
  }
  y -= 14

  addLine(lines, 'Produtos utilizados', 48, y, 12, 'F2')
  y -= 18
  addLine(lines, 'Produto', 48, y, 9, 'F2')
  addLine(lines, 'Qtd.', 450, y, 9, 'F2')
  y -= 14

  if (receipt.items.length) {
    for (const item of receipt.items) {
      addLine(lines, truncate(item.stockItem?.name || item.stockItemId, 58), 48, y)
      addLine(lines, String(item.quantity), 450, y)
      y -= 14
    }
  } else {
    addLine(lines, 'Nenhum produto vinculado.', 48, y)
    y -= 14
  }
  y -= 22

  addLine(lines, 'Resumo financeiro', 48, y, 12, 'F2')
  y -= 18
  addLine(lines, 'Pagamento', 48, y)
  addLine(lines, paymentMethodLabel(receipt), 360, y, 10, 'F2')
  y -= 18
  addMoneyLine(lines, 'Valor total', receipt.priceCents, y, 13)
  y -= 50
  addLine(lines, 'Este recibo não é uma nota fiscal.', 48, Math.max(y, 54), 10, 'F2')

  const stream = lines.map((line) => drawText(line)).join('\n')
  return createPdf(stream)
}

function addLine(lines: PdfLine[], text: string, x: number, y: number, size = 10, font: 'F1' | 'F2' = 'F1') {
  lines.push({ text, x, y, size, font })
}

function addMoneyLine(lines: PdfLine[], label: string, value: number, y: number, size = 10) {
  addLine(lines, label, 48, y, size)
  addLine(lines, formatCurrency(value), 360, y, size, 'F2')
}

function drawText(line: PdfLine) {
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

function paymentMethodLabel(receipt: IReceipt) {
  if (receipt.paymentMethod === 'credit_card') return `Cartão de crédito (${receipt.installments || 1}x)`
  if (receipt.paymentMethod === 'debit_card') return 'Cartão de débito'
  if (receipt.paymentMethod === 'pix') return 'Pix'
  return 'Dinheiro'
}

function statusLabel(status: IReceipt['status']) {
  if (status === 'paid') return 'Pago'
  if (status === 'cancelled') return 'Cancelado'
  return 'Pendente'
}

function wrapText(value: string, maxLength: number) {
  const words = value.split(/\s+/)
  const lines: string[] = []
  let current = ''

  for (const word of words) {
    if (`${current} ${word}`.trim().length > maxLength) {
      if (current) lines.push(current)
      current = word
    } else {
      current = `${current} ${word}`.trim()
    }
  }
  if (current) lines.push(current)
  return lines.length ? lines : ['-']
}

function truncate(value: string, maxLength: number) {
  return value.length > maxLength ? `${value.slice(0, maxLength - 1)}…` : value
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
