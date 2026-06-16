import type { IStockItem } from '../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const body = await readBody(event)

  return $fetch<IStockItem>(`${config.apiBase}/api/stock`, {
    method: 'POST',
    body,
    headers: useRequestHeaders(['authorization'])
  })
})
