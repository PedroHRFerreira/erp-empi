export function onlyDigits(value: string): string {
  return value.replace(/\D/g, '')
}

export function isCpf(value: string): boolean {
  const cpf = onlyDigits(value)
  if (cpf.length !== 11 || /^(\d)\1+$/.test(cpf)) {
    return false
  }
  return validateCpfDigit(cpf, 9) && validateCpfDigit(cpf, 10)
}

function validateCpfDigit(cpf: string, position: number): boolean {
  let sum = 0
  for (let index = 0; index < position; index += 1) {
    sum += Number(cpf[index]) * (position + 1 - index)
  }
  let digit = (sum * 10) % 11
  if (digit === 10) {
    digit = 0
  }
  return digit === Number(cpf[position])
}

export function isPlate(value: string): boolean {
  return /^[A-Z]{3}[0-9][A-Z0-9][0-9]{2}$/i.test(value.replace(/[^a-z0-9]/gi, ''))
}
