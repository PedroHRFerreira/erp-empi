import type { ICashEntry } from '../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')
  return $fetch<ICashEntry[]>(`${config.apiBase}/api/cash/daily-entries`, {
    query: getQuery(event),
    headers: authorization ? { Authorization: authorization } : undefined
  })
})
