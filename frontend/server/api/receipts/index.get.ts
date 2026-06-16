import type { IReceipt, IPaginated } from '../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')

  return $fetch<IPaginated<IReceipt>>(`${config.apiBase}/api/receipts`, {
    query: getQuery(event),
    headers: authorization
      ? {
          Authorization: authorization
        }
      : undefined
  })
})
