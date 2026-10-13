import { createError, defineEventHandler, getHeader, getMethod, getRouterParam, proxyRequest, setHeader } from 'h3'

export default defineEventHandler(async event => {
  const path = getRouterParam(event, 'path') || ''
  const method = getMethod(event)
  const valid = (path === 'templates' && method === 'GET')
    || (path === 'brand' && ['GET', 'PUT'].includes(method))
    || (path === 'creations' && ['GET', 'POST'].includes(method))
    || (/^creations\/[a-zA-Z0-9-]+$/.test(path) && ['PUT', 'DELETE'].includes(method))
    || (path === 'images' && method === 'POST')
    || (/^images\/[a-zA-Z0-9-]+$/.test(path) && method === 'GET')
  if (!valid) throw createError({ statusCode: 404, statusMessage: 'Not found' })
  const authorization = getHeader(event, 'authorization')
  if (!authorization) throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
  const contentLength = Number(getHeader(event, 'content-length') || 0)
  if (method === 'POST' && path === 'images' && contentLength > 6 * 1024 * 1024) {
    throw createError({ statusCode: 413, statusMessage: 'Image too large' })
  }
  setHeader(event, 'Cache-Control', 'no-store')
  const config = useRuntimeConfig(event)
  const query = event.node.req.url?.split('?')[1]
  return proxyRequest(event, `${config.apiBase}/api/marketing/${path}${query ? `?${query}` : ''}`, {
    headers: { Authorization: authorization, 'Cache-Control': 'no-store' }
  })
})
