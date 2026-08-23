export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')
  return $fetch<unknown>(`${config.apiBase}/api/payables/${event.context.params?.id}/pay`, {
    method: 'POST',
    body: await readBody(event),
    headers: authorization ? { Authorization: authorization } : undefined
  })
})
