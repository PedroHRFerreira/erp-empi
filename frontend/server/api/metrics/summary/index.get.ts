import type { IMetricsSummary } from '../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')

  return $fetch<IMetricsSummary>(`${config.apiBase}/api/metrics/summary`, {
    headers: authorization
      ? {
          Authorization: authorization
        }
      : undefined
  })
})
