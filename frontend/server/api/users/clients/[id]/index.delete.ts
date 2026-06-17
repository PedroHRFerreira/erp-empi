import type { IUser } from '../../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')

  return $fetch<IUser>(`${config.apiBase}/api/users/clients/${event.context.params?.id}`, {
    method: 'DELETE',
    headers: authorization
      ? {
          Authorization: authorization
        }
      : undefined
  })
})
