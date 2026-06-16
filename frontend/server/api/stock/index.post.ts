import type { IStockItem } from '../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')
  const body = await readBody(event)

  return $fetch<IStockItem>(`${config.apiBase}/api/stock`, {
    method: 'POST',
    body,
    headers: authorization
      ? {
          Authorization: authorization
        }
      : undefined
  })
})
