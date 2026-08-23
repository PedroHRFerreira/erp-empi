import type { IPayableAlert } from '../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')
  return $fetch<IPayableAlert[]>(`${config.apiBase}/api/payables/alerts`, {
    headers: authorization ? { Authorization: authorization } : undefined
  })
})
