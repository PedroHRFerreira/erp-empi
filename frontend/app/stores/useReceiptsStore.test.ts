import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { makeReceiptForm, useReceiptsStore } from './useReceiptsStore'

describe('useReceiptsStore receipt creation', () => {
  beforeEach(() => setActivePinia(createPinia()))

  it.each([false, true])('blocks a concurrent submission when quick is %s', async (quick) => {
    const store = useReceiptsStore()
    store.creating = true

    const result = await store.create(makeReceiptForm({}, quick))

    expect(result).toEqual({
      status: 'error',
      errors: 'O recibo já está sendo salvo.',
      message: 'O recibo já está sendo salvo.',
    })
  })
})
