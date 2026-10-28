import type { IStockSupplier } from "../../../contracts/types"

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, "authorization")
  return $fetch<IStockSupplier[]>(`${config.apiBase}/api/stock/suppliers`, {
    headers: authorization ? { Authorization: authorization } : undefined
  })
})
