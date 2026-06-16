import type { IReceipt } from '../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')

  return $fetch<IReceipt>(`${config.apiBase}/api/receipts/${event.context.params?.id}`, {
    headers: authorization
      ? {
          Authorization: authorization
        }
      : undefined
  })
})
