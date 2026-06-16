export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)

  await $fetch(`${config.apiBase}/api/stock/${event.context.params?.id}`, {
    method: 'DELETE',
    headers: useRequestHeaders(['authorization'])
  })

  return null
})
