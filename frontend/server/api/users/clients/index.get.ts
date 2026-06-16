import type { IPaginated, IUser } from '../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')

  return $fetch<IPaginated<IUser>>(`${config.apiBase}/api/users/clients`, {
    query: getQuery(event),
    headers: authorization
      ? {
          Authorization: authorization
        }
      : undefined
  })
})
