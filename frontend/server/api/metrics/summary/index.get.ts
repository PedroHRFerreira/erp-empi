import type { IMetricsSummary } from '../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)

  return $fetch<IMetricsSummary>(`${config.apiBase}/api/metrics/summary`, {
    headers: useRequestHeaders(['authorization'])
  })
})
