import type { IUser } from '../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')
  const body = await readBody(event)

  return $fetch<IUser>(`${config.apiBase}/api/users/profile`, {
    method: 'PUT',
    body,
    headers: authorization
      ? {
          Authorization: authorization
        }
      : undefined
  })
})
