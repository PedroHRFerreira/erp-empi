import type { IReceipt } from '../../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')

  return $fetch<IReceipt>(`${config.apiBase}/api/receipts/${event.context.params?.id}/pay`, {
    method: 'POST',
    headers: authorization
      ? {
          Authorization: authorization
        }
      : undefined
  })
})
