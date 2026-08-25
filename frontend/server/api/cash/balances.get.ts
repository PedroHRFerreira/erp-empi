import type { ICashBalances } from '../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')
  return $fetch<ICashBalances>(`${config.apiBase}/api/cash/balances`, {
    headers: authorization ? { Authorization: authorization } : undefined
  })
})
