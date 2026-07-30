type CacheEntry = {
  path: string
  value: unknown
}

const responses = new Map<string, CacheEntry>()
const pendingRequests = new Map<string, Promise<unknown>>()
let version = 0

export async function getCachedApiResponse<T>(key: string, path: string, request: () => Promise<T>): Promise<T> {
  const cached = responses.get(key)
  if (cached) return cached.value as T

  const pending = pendingRequests.get(key)
  if (pending) return pending as Promise<T>

  const requestVersion = version
  const promise = request()
    .then((response) => {
      if (requestVersion === version) {
        responses.set(key, { path, value: response })
      }
      return response
    })
    .finally(() => {
      pendingRequests.delete(key)
    })

  pendingRequests.set(key, promise)
  return promise
}

export function invalidateApiCache(paths: string[]): void {
  version += 1
  for (const [key, entry] of responses) {
    if (paths.some((path) => entry.path === path || entry.path.startsWith(`${path}/`))) {
      responses.delete(key)
    }
  }
  pendingRequests.clear()
}

export function clearApiCache(): void {
  version += 1
  responses.clear()
  pendingRequests.clear()
}
