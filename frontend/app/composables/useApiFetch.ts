import type { UseFetchOptions } from 'nuxt/app'
import { ref, shallowRef } from 'vue'
import type {
  ApiFetchClient,
  ApiFetchError,
  ApiFetchRequestOptions,
  ApiFetchResult,
  ApiFetchStatus,
  ApiMethod
} from '~/types/api'
import { getCachedApiResponse } from '~/utils/apiCache'

const getApiOptions = (
  path: string,
  body?: unknown,
  optionHeaders: Record<string, string> = {},
  useBaseApiURL = true
) => {
  const isPublicAuthRoute = path === '/auth/login' || path === '/login'
  const config = useRuntimeConfig()
  const authToken = useAuthToken()
  const authorization = authToken.getAuthorizationHeader()
  const baseApiURL = useBaseApiURL ? config.public.apiBaseUrl || '/api' : ''
  const isFormData = typeof FormData !== 'undefined' && body instanceof FormData

  let headers: Record<string, string> = {
    Accept: 'application/json'
  }

  if (authorization && !isPublicAuthRoute) {
    headers.Authorization = authorization
  }

  if (import.meta.server) {
    headers = {
      ...headers,
      ...useRequestHeaders(['referer', 'cookie'])
    }
  }

  if (!isFormData) {
    headers['Content-Type'] = 'application/json'
  }

  return {
    baseApiURL,
    authorization,
    headers: {
      ...headers,
      ...optionHeaders
    }
  }
}

export function useApiFetch<T>(
  path: string,
  options: UseFetchOptions<T> = {},
  useBaseApiURL = true
): Promise<ApiFetchResult<T>> {
  const method = String(options.method || 'GET').toUpperCase()
  const { baseApiURL, authorization, headers } = getApiOptions(
    path,
    options.body,
    options.headers as Record<string, string>,
    useBaseApiURL
  )

  const data = shallowRef(null) as Ref<T | null>
  const error = ref<ApiFetchError | null>(null)
  const status = ref<ApiFetchStatus>('pending')
  const apiFetch = $fetch as ApiFetchClient

  const request = () => {
    return apiFetch<T>(baseApiURL + path, {
      method: method as ApiMethod,
      body: options.body as BodyInit | Record<string, unknown> | null | undefined,
      query: options.query,
      headers: {
        ...headers,
        ...(options.headers as Record<string, string>)
      }
    } as ApiFetchRequestOptions)
  }

  const canCache = import.meta.client && method === 'GET'
  const cacheKey = `${authorization || 'anonymous'}:${path}:${stableSerialize(options.query || {})}`

  return (canCache ? getCachedApiResponse(cacheKey, path, request) : request())
    .then((response) => {
      data.value = response as T
      status.value = 'success'

      return {
        data,
        error,
        status
      }
    })
    .catch(async (fetchError: ApiFetchError) => {
      error.value = fetchError
      status.value = 'error'

      return {
        data,
        error,
        status
      }
    })
}

function stableSerialize(value: unknown): string {
  if (Array.isArray(value)) return `[${value.map(stableSerialize).join(',')}]`
  if (value && typeof value === 'object') {
    return `{${Object.entries(value as Record<string, unknown>)
      .sort(([left], [right]) => left.localeCompare(right))
      .map(([key, item]) => `${key}:${stableSerialize(item)}`)
      .join(',')}}`
  }
  return JSON.stringify(value)
}
