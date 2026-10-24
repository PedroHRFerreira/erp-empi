export type ReceiptPaymentForm = {
  paymentMethod: 'cash' | 'pix' | 'credit_card' | 'debit_card'
  baseAmountCents: number
  installments: number
  cardFeePercent: number
}

export type ReceiptPaymentCalculated = ReceiptPaymentForm & {
  cardFeeCents: number
  amountCents: number
}

export type ReceiptPaymentsCalculation = {
  payments: ReceiptPaymentCalculated[]
  baseAmountCents: number
  cardFeeCents: number
  amountCents: number
}

const paymentMethods = new Set(['cash', 'pix', 'credit_card', 'debit_card'])

function isCard(paymentMethod: string): boolean {
  return paymentMethod === 'credit_card' || paymentMethod === 'debit_card'
}

/** Derives the second amount from the subtotal; callers must validate before saving. */
export function calculateReceiptPayments(
  payments: readonly ReceiptPaymentForm[],
  subtotalCents: number,
): ReceiptPaymentsCalculation {
  const calculated = payments.map((payment, index) => {
    const baseAmountCents = payments.length === 1
      ? subtotalCents
      : index === 1 ? subtotalCents - payments[0]!.baseAmountCents : payment.baseAmountCents
    const cardFeePercent = isCard(payment.paymentMethod) ? payment.cardFeePercent : 0
    const cardFeeCents = isCard(payment.paymentMethod)
      ? Math.trunc(baseAmountCents * (cardFeePercent / 100))
      : 0
    return {
      ...payment,
      baseAmountCents,
      cardFeePercent,
      cardFeeCents,
      amountCents: baseAmountCents + cardFeeCents,
    }
  })

  return {
    payments: calculated,
    baseAmountCents: calculated.reduce((sum, payment) => sum + payment.baseAmountCents, 0),
    cardFeeCents: calculated.reduce((sum, payment) => sum + payment.cardFeeCents, 0),
    amountCents: calculated.reduce((sum, payment) => sum + payment.amountCents, 0),
  }
}

/** Validates editable inputs using the same derived remainder shown in the UI. */
export function validateReceiptPayments(
  payments: readonly ReceiptPaymentForm[],
  subtotalCents: number,
): Record<string, string> {
  const errors: Record<string, string> = {}
  if (!Number.isSafeInteger(subtotalCents) || subtotalCents < 0) {
    errors.subtotalCents = 'O subtotal deve ser um valor válido em centavos.'
  }
  if (payments.length < 1 || payments.length > 2) {
    errors.payments = 'Selecione uma ou duas formas de pagamento.'
    return errors
  }

  const calculation = calculateReceiptPayments(payments, subtotalCents)
  const methods = new Set<string>()
  payments.forEach((payment, index) => {
    const prefix = `payments.${index}`
    if (!paymentMethods.has(payment.paymentMethod)) {
      errors[`${prefix}.paymentMethod`] = 'Selecione uma forma de pagamento válida.'
    } else if (methods.has(payment.paymentMethod)) {
      errors[`${prefix}.paymentMethod`] = 'Escolha duas formas de pagamento diferentes.'
    }
    methods.add(payment.paymentMethod)

    const amount = calculation.payments[index]!.baseAmountCents
    if (!Number.isSafeInteger(amount) || amount < 0 || (payments.length === 2 && amount === 0)) {
      errors[`${prefix}.baseAmountCents`] = payments.length === 2
        ? 'Cada forma de pagamento deve ter um valor maior que zero e dentro do subtotal.'
        : 'Informe um valor válido em centavos.'
    }
    if (!Number.isFinite(payment.cardFeePercent) || payment.cardFeePercent < 0) {
      errors[`${prefix}.cardFeePercent`] = 'A taxa deve ser um percentual válido maior ou igual a zero.'
    } else if (!Number.isSafeInteger(calculation.payments[index]!.amountCents)) {
      errors[`${prefix}.cardFeePercent`] = 'O valor calculado com a taxa está fora do limite permitido.'
    }

    const installments = payment.installments
    if (!Number.isInteger(installments)
      || installments < 1
      || installments > (payment.paymentMethod === 'credit_card' ? 12 : 1)) {
      errors[`${prefix}.installments`] = payment.paymentMethod === 'credit_card'
        ? 'Selecione de 1 a 12 parcelas.'
        : 'Esta forma de pagamento deve ser à vista.'
    }
  })

  return errors
}
