import type { IClientDetail } from '../../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')

  return $fetch<IClientDetail>(`${config.apiBase}/api/users/clients/${event.context.params?.id}/detail`, {
    headers: authorization
      ? {
          Authorization: authorization
        }
      : undefined
  })
})
