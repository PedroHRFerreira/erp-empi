export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')
  return $fetch<unknown>(`${config.apiBase}/api/stock/purchases/${event.context.params?.id}`, {
    method: 'DELETE',
    headers: authorization ? { Authorization: authorization } : undefined
  })
})
