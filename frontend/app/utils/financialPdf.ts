import type { IExpense, IFinancialSummary } from '../../server/contracts/types'
import { formatCurrency } from './format'

type PdfLine = {
  text: string
  x: number
  y: number
  size?: number
  font?: 'F1' | 'F2'
}

export function downloadFinancialReportPdf(summary: IFinancialSummary, expenses: IExpense[]) {
  const bytes = buildFinancialReportPdfBytes(summary, expenses)
  const file = new File([bytes], `relatorio-financeiro-${summary.startDate}-${summary.endDate}.pdf`, {
    type: 'application/pdf'
  })
  const href = URL.createObjectURL(file)
  const link = document.createElement('a')
  link.href = href
  link.download = file.name
  link.click()
  URL.revokeObjectURL(href)
}

function buildFinancialReportPdfBytes(summary: IFinancialSummary, expenses: IExpense[]) {
  const lines: PdfLine[] = []
  let y = 790

  addLine(lines, 'EMPI Autocenter', 48, y, 18, 'F2')
  addLine(lines, 'Relatório financeiro', 360, y, 18, 'F2')
  y -= 20
  addLine(lines, `Período: ${formatDate(summary.startDate)} até ${formatDate(summary.endDate)}`, 360, y, 9)
  y -= 34

  addLine(lines, `Status: ${healthLabel(summary.healthStatus)}`, 48, y, 13, 'F2')
  addLine(lines, `Margem líquida: ${formatPercent(summary.netMarginPercent)}`, 360, y, 11, 'F2')
  y -= 30

  const summaryRows = [
    ['Receita recebida', summary.revenuePaidCents],
    ['Custo dos produtos', summary.productCostCents],
    ['Taxas de cartão', summary.cardFeesCents],
    ['Lucro bruto', summary.grossProfitCents],
    ['Gastos operacionais', summary.operationalExpensesCents],
    ['Lucro operacional', summary.operationalProfitCents],
    ['Lucro líquido', summary.netProfitCents]
  ]

  addLine(lines, 'Demonstrativo', 48, y, 12, 'F2')
  y -= 18
  for (const [label, value] of summaryRows) {
    addMoneyLine(lines, label as string, value as number, y)
    y -= 15
  }
  y -= 12

  addLine(lines, 'Gastos por categoria', 48, y, 12, 'F2')
  y -= 18
  if (summary.expensesByCategory.length) {
    for (const item of summary.expensesByCategory.slice(0, 8)) {
      addLine(lines, `${item.category} (${item.count})`, 48, y)
      addLine(lines, formatCurrency(item.amountCents), 360, y, 10, 'F2')
      y -= 14
    }
  } else {
    addLine(lines, 'Nenhum gasto registrado no período.', 48, y)
    y -= 14
  }
  y -= 16

  addLine(lines, 'Gastos lançados', 48, y, 12, 'F2')
  y -= 18
  if (expenses.length) {
    for (const expense of expenses.slice(0, 12)) {
      addLine(lines, truncate(`${formatDate(expense.spentAt)} - ${expense.description}`, 48), 48, y)
      addLine(lines, truncate(expense.category, 18), 300, y)
      addLine(lines, formatCurrency(expense.amountCents), 430, y, 10, 'F2')
      y -= 14
    }
  } else {
    addLine(lines, 'Nenhum gasto encontrado.', 48, y)
  }

  const stream = lines.map((line) => drawText(line)).join('\n')
  return createPdf(stream)
}

function addLine(lines: PdfLine[], text: string, x: number, y: number, size = 10, font: 'F1' | 'F2' = 'F1') {
  lines.push({ text, x, y, size, font })
}

function addMoneyLine(lines: PdfLine[], label: string, value: number, y: number) {
  addLine(lines, label, 48, y)
  addLine(lines, formatCurrency(value), 360, y, 10, 'F2')
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

function formatDate(value: string) {
  if (!value) return '-'
  if (/^\d{4}-\d{2}-\d{2}$/.test(value)) {
    const parts = value.split('-').map(Number)
    const year = parts[0] || 0
    const month = parts[1] || 1
    const day = parts[2] || 1
    return new Intl.DateTimeFormat('pt-BR').format(new Date(year, month - 1, day))
  }
  return new Intl.DateTimeFormat('pt-BR').format(new Date(value))
}

function formatPercent(value: number) {
  return `${(Number.isFinite(value) ? value : 0).toFixed(1)}%`
}

function healthLabel(status: IFinancialSummary['healthStatus']) {
  if (status === 'red') return 'Vermelho'
  if (status === 'yellow') return 'Amarelo'
  return 'Verde'
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
