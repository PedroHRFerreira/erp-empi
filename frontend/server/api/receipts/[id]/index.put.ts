import type { IReceipt } from '../../../contracts/types'

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const authorization = getHeader(event, 'authorization')
  const body = await readBody(event)

  try {
    return await $fetch<IReceipt>(`${config.apiBase}/api/receipts/${event.context.params?.id}`, {
      method: 'PUT',
      body,
      headers: authorization
        ? {
            Authorization: authorization
          }
        : undefined
    })
  } catch (error: any) {
    const message = error?.data?.message || error?.statusMessage || 'Não foi possível atualizar o recibo.'

    throw createError({
      statusCode: error?.statusCode || 500,
      statusMessage: message,
      message,
      data: {
        message
      }
    })
  }
})
