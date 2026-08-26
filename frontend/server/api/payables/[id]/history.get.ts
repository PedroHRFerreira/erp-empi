export default defineEventHandler(async (event): Promise<unknown> => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')
  return $fetch<unknown>(`${config.apiBase}/api/payables/${event.context.params?.id}/history`, {
    headers: authorization ? { Authorization: authorization } : undefined,
  })
})
