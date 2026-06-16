import type { ILoginResponse } from '../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const body = await readBody(event)

  return $fetch<ILoginResponse>(`${config.apiBase}/api/auth/login`, {
    method: 'POST',
    body
  })
})
