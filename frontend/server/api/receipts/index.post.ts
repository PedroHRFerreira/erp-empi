import type { IReceipt } from '../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')
  const body = await readBody(event)

  return $fetch<IReceipt>(`${config.apiBase}/api/receipts`, {
    method: 'POST',
    body,
    headers: authorization
      ? {
          Authorization: authorization
        }
      : undefined
  })
})
