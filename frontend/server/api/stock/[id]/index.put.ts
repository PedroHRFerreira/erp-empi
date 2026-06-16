import type { IStockItem } from '../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const body = await readBody(event)

  return $fetch<IStockItem>(`${config.apiBase}/api/stock/${event.context.params?.id}`, {
    method: 'PUT',
    body,
    headers: useRequestHeaders(['authorization'])
  })
})
