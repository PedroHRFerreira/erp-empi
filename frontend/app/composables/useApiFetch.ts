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
  const { baseApiURL, headers } = getApiOptions(
    path,
    options.body,
    options.headers as Record<string, string>,
    useBaseApiURL
  )

  const data = shallowRef(null) as Ref<T | null>
  const error = ref<ApiFetchError | null>(null)
  const status = ref<ApiFetchStatus>('pending')
  const apiFetch = $fetch as ApiFetchClient

  return apiFetch<T>(baseApiURL + path, {
    method: method as ApiMethod,
    body: options.body as BodyInit | Record<string, unknown> | null | undefined,
    query: options.query,
    headers: {
      ...headers,
      ...(options.headers as Record<string, string>)
    }
  } as ApiFetchRequestOptions)
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
