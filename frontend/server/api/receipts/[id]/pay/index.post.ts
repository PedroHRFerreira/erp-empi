import type { IReceipt } from '../../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)

  return $fetch<IReceipt>(`${config.apiBase}/api/receipts/${event.context.params?.id}/pay`, {
    method: 'POST',
    headers: useRequestHeaders(['authorization'])
  })
})
