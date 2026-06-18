import type { IExpense } from '../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')
  const body = await readBody(event)

  return $fetch<IExpense>(`${config.apiBase}/api/expenses/${event.context.params?.id}`, {
    method: 'PUT',
    body,
    headers: authorization
      ? {
          Authorization: authorization
        }
      : undefined
  })
})
