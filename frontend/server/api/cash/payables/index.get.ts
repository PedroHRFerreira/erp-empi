import type { IPayableInstallment } from '../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')
  return $fetch<IPayableInstallment[]>(`${config.apiBase}/api/cash/payables`, {
    headers: authorization ? { Authorization: authorization } : undefined
  })
})
