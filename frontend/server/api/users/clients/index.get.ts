import type { IPaginated, IUser } from '../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)

  return $fetch<IPaginated<IUser>>(`${config.apiBase}/api/users/clients`, {
    query: getQuery(event),
    headers: useRequestHeaders(['authorization'])
  })
})
