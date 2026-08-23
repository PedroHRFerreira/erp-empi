import type { IStockPurchase } from '../../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')
  return $fetch<IStockPurchase>(`${config.apiBase}/api/stock/purchases/${event.context.params?.id}`, {
    headers: authorization ? { Authorization: authorization } : undefined
  })
})
