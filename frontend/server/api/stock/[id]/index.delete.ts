export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')

  await $fetch(`${config.apiBase}/api/stock/${event.context.params?.id}`, {
    method: 'DELETE',
    headers: authorization
      ? {
          Authorization: authorization
        }
      : undefined
  })

  return null
})
