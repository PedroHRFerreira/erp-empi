import type { IStockItem } from "../../../contracts/types"

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, "authorization")
  return $fetch<IStockItem>(`${config.apiBase}/api/stock/${event.context.params?.id}`, {
    headers: authorization ? { Authorization: authorization } : undefined
  })
})
