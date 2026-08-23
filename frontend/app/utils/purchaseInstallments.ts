export type InstallmentScheduleEntry = {
  amountCents: number
  dueDate: string
}

export type InstallmentScheduleValidation = {
  valid: boolean
  errors: Array<'empty' | 'invalid_amount' | 'invalid_date' | 'total_mismatch'>
}

export function divideCents(totalCents: number, installmentCount: number): number[] {
  const total = Number.isFinite(totalCents) ? Math.max(0, Math.trunc(totalCents)) : 0
  const count = Number.isFinite(installmentCount) ? Math.max(1, Math.trunc(installmentCount)) : 1
  const base = Math.floor(total / count)
  const remainder = total % count

  return Array.from({ length: count }, (_, index) => base + (index < remainder ? 1 : 0))
}

export function addMonthsStable(value: string, months: number): string {
  const date = parseIsoDate(value)
  if (!date || !Number.isInteger(months)) return ''

  const targetMonth = new Date(Date.UTC(date.year, date.month - 1 + months, 1))
  const year = targetMonth.getUTCFullYear()
  const month = targetMonth.getUTCMonth() + 1
  const lastDay = new Date(Date.UTC(year, month, 0)).getUTCDate()

  return formatIsoDate(year, month, Math.min(date.day, lastDay))
}

export function generateInstallmentSchedule(
  totalCents: number,
  installmentCount: number,
  firstDueDate: string,
): InstallmentScheduleEntry[] {
  if (!parseIsoDate(firstDueDate)) return []

  return divideCents(totalCents, installmentCount).map((amountCents, index) => ({
    amountCents,
    dueDate: addMonthsStable(firstDueDate, index),
  }))
}

export function validateInstallmentSchedule(
  totalCents: number,
  installments: readonly InstallmentScheduleEntry[],
): InstallmentScheduleValidation {
  const errors = new Set<InstallmentScheduleValidation['errors'][number]>()

  if (installments.length === 0) errors.add('empty')
  if (installments.some(({ amountCents }) => !Number.isInteger(amountCents) || amountCents <= 0)) {
    errors.add('invalid_amount')
  }
  if (installments.some(({ dueDate }) => !parseIsoDate(dueDate))) errors.add('invalid_date')

  const expectedTotal = Number.isFinite(totalCents) ? Math.trunc(totalCents) : -1
  const actualTotal = installments.reduce((sum, { amountCents }) => sum + amountCents, 0)
  if (expectedTotal < 0 || actualTotal !== expectedTotal) errors.add('total_mismatch')

  return { valid: errors.size === 0, errors: [...errors] }
}

function parseIsoDate(value: string): { year: number; month: number; day: number } | null {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value)
  if (!match) return null

  const year = Number(match[1])
  const month = Number(match[2])
  const day = Number(match[3])
  const date = new Date(Date.UTC(year, month - 1, day))
  if (
    date.getUTCFullYear() !== year
    || date.getUTCMonth() !== month - 1
    || date.getUTCDate() !== day
  ) return null

  return { year, month, day }
}

function formatIsoDate(year: number, month: number, day: number): string {
  return `${String(year).padStart(4, '0')}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`
}
