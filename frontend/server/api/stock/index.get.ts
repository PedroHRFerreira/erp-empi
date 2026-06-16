import type { IPaginated, IStockItem } from '../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)

  return $fetch<IPaginated<IStockItem>>(`${config.apiBase}/api/stock`, {
    query: getQuery(event),
    headers: useRequestHeaders(['authorization'])
  })
})
