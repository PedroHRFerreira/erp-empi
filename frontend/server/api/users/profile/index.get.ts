import type { IUser } from '../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)

  return $fetch<IUser>(`${config.apiBase}/api/users/profile`, {
    headers: useRequestHeaders(['authorization'])
  })
})
