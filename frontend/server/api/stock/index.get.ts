import type { IPaginated, IStockItem } from '../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')

  return $fetch<IPaginated<IStockItem>>(`${config.apiBase}/api/stock`, {
    query: getQuery(event),
    headers: authorization
      ? {
          Authorization: authorization
        }
      : undefined
  })
})
