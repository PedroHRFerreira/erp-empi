import type { IUser } from '../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const body = await readBody(event)

  return $fetch<IUser>(`${config.apiBase}/api/users/profile`, {
    method: 'PUT',
    body,
    headers: useRequestHeaders(['authorization'])
  })
})
