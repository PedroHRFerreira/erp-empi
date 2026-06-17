export function maskCpf(value: string): string {
  return onlyDigits(value)
    .slice(0, 11)
    .replace(/(\d{3})(\d)/, '$1.$2')
    .replace(/(\d{3})(\d)/, '$1.$2')
    .replace(/(\d{3})(\d{1,2})$/, '$1-$2')
}

export function maskPhone(value: string): string {
  const digits = onlyDigits(value).slice(0, 11)

  if (digits.length <= 2) {
    return digits
  }

  const area = digits.slice(0, 2)
  const prefixLength = digits.length <= 10 ? 4 : 5
  const prefix = digits.slice(2, 2 + prefixLength)
  const suffix = digits.slice(2 + prefixLength)

  return suffix ? `(${area}) ${prefix}-${suffix}` : `(${area}) ${prefix}`
}

export function maskVehiclePlate(value: string): string {
  return value.replace(/[^a-zA-Z0-9]/g, '').slice(0, 7).toUpperCase()
}

export function maskCurrency(value: string): string {
  return formatCentsAsCurrency(currencyMaskToCents(value))
}

export function currencyMaskToCents(value: string): number {
  return Number(onlyDigits(value) || 0)
}

export function formatCentsAsCurrency(valueInCents: number): string {
  return new Intl.NumberFormat('pt-BR', {
    style: 'currency',
    currency: 'BRL'
  }).format((valueInCents || 0) / 100)
}

function onlyDigits(value: string): string {
  return value.replace(/\D/g, '')
}
