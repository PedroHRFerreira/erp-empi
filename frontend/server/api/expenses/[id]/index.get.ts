import type { IExpense } from '../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')

  return $fetch<IExpense>(`${config.apiBase}/api/expenses/${event.context.params?.id}`, {
    headers: authorization ? { Authorization: authorization } : undefined
  })
})
