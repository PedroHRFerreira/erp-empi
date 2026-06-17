export interface IStoreActionResult<T = unknown> {
  status: 'success' | 'error'
  errors?: string | Record<string, string>
  message?: string
  data?: T
}
