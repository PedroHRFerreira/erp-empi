import type { IGoalsSummary } from '../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')
  return $fetch<IGoalsSummary>(`${config.apiBase}/api/goals`, {
    query: getQuery(event),
    headers: authorization ? { Authorization: authorization } : undefined
  })
})
