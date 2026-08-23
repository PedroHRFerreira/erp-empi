import type { ICashSession } from '../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')
  return $fetch<ICashSession | null>(`${config.apiBase}/api/cash/current`, {
    headers: authorization ? { Authorization: authorization } : undefined
  })
})
