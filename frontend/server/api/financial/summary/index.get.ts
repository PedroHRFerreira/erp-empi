import type { IFinancialSummary } from '../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')

  return $fetch<IFinancialSummary>(`${config.apiBase}/api/financial/summary`, {
    query: getQuery(event),
    headers: authorization
      ? {
          Authorization: authorization
        }
      : undefined
  })
})
