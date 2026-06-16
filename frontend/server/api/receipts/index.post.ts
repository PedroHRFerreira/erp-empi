import type { IReceipt } from '../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const body = await readBody(event)

  return $fetch<IReceipt>(`${config.apiBase}/api/receipts`, {
    method: 'POST',
    body,
    headers: useRequestHeaders(['authorization'])
  })
})
