import type { IGoalsSummary } from '../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')
  return $fetch<IGoalsSummary>(`${config.apiBase}/api/goals`, {
    method: 'PUT',
    query: getQuery(event),
    body: await readBody(event),
    headers: authorization ? { Authorization: authorization } : undefined
  })
})
