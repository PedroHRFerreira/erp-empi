import type { Ref } from 'vue'

export type ApiMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'

export type ApiFetchStatus = 'idle' | 'pending' | 'success' | 'error'

export type ApiFetchError = {
  data?: {
    message?: string
    errors?: Record<string, string[]>
  }
  response?: {
    status?: number
  }
  statusCode?: number
}

export type ApiFetchResult<T> = {
  data: Ref<T | null>
  error: Ref<ApiFetchError | null>
  status: Ref<ApiFetchStatus>
}

export type ApiFetchRequestOptions = {
  method: ApiMethod
  body?: BodyInit | Record<string, unknown> | null
  query?: unknown
  headers: Record<string, string>
}

export type ApiFetchClient = <T>(url: string, options: ApiFetchRequestOptions) => Promise<T>
