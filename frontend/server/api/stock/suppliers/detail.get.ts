import type { IStockSupplierDetail } from "../../../contracts/types"

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, "authorization")
  return $fetch<IStockSupplierDetail>(`${config.apiBase}/api/stock/suppliers/detail`, {
    query: getQuery(event),
    headers: authorization ? { Authorization: authorization } : undefined
  })
})
