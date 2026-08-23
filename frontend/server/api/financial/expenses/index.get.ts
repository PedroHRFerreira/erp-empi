import type { IPaginated, IRealizedExpense } from '../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')

  return $fetch<IPaginated<IRealizedExpense>>(`${config.apiBase}/api/financial/expenses`, {
    query: getQuery(event),
    headers: authorization
      ? {
          Authorization: authorization
        }
      : undefined
  })
})
