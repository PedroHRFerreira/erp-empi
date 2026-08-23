import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useExpensesStore } from './useExpensesStore'

describe('useExpensesStore installment validation', () => {
  beforeEach(() => setActivePinia(createPinia()))

  it('exige uma agenda de parcelas', () => {
    const store = useExpensesStore()
    const valid = store.validate({
      description: 'Energia',
      category: 'energia',
      amountCents: 1000,
      spentAt: '2026-08-17',
      notes: '',
      installments: []
    })

    expect(valid).toBe(false)
    expect(store.fieldErrors.installments).toBe('Gere ao menos uma parcela.')
  })

  it('aceita parcelas cuja soma corresponde ao gasto', () => {
    const store = useExpensesStore()
    expect(store.validate({
      description: 'Energia',
      category: 'energia',
      amountCents: 1000,
      spentAt: '2026-08-17',
      notes: '',
      installments: [
        { amountCents: 500, dueDate: '2026-08-20', plannedMethod: 'boleto' },
        { amountCents: 500, dueDate: '2026-09-20', plannedMethod: 'pix' }
      ]
    })).toBe(true)
  })

  it('rejeita parcelas com soma divergente', () => {
    const store = useExpensesStore()
    expect(store.validate({
      description: 'Energia',
      category: 'energia',
      amountCents: 1000,
      spentAt: '2026-08-17',
      notes: '',
      installments: [{ amountCents: 999, dueDate: '2026-08-20', plannedMethod: 'boleto' }]
    })).toBe(false)
    expect(store.fieldErrors.installments).toBe('A soma das parcelas deve ser igual ao valor do gasto.')
  })
})
