import { beforeEach, describe, expect, it, vi } from 'vitest'
import { clearApiCache, getCachedApiResponse, invalidateApiCache } from './apiCache'

describe('api cache', () => {
  beforeEach(() => {
    clearApiCache()
  })

  it('reuses a successful response for the same request', async () => {
    const request = vi.fn().mockResolvedValue({ id: 'receipt-1' })

    await getCachedApiResponse('user:/receipts:page-1', '/receipts', request)
    const result = await getCachedApiResponse('user:/receipts:page-1', '/receipts', request)

    expect(request).toHaveBeenCalledTimes(1)
    expect(result).toEqual({ id: 'receipt-1' })
  })

  it('shares simultaneous requests with the same key', async () => {
    let resolveRequest: (value: { id: string }) => void = () => undefined
    const request = vi.fn(() => new Promise<{ id: string }>((resolve) => {
      resolveRequest = resolve
    }))

    const first = getCachedApiResponse('user:/stock:page-1', '/stock', request)
    const second = getCachedApiResponse('user:/stock:page-1', '/stock', request)
    resolveRequest({ id: 'stock-1' })

    await expect(Promise.all([first, second])).resolves.toEqual([{ id: 'stock-1' }, { id: 'stock-1' }])
    expect(request).toHaveBeenCalledTimes(1)
  })

  it('does not cache failures', async () => {
    const request = vi.fn()
      .mockRejectedValueOnce(new Error('offline'))
      .mockResolvedValueOnce({ id: 'expense-1' })

    await expect(getCachedApiResponse('user:/expenses:page-1', '/expenses', request)).rejects.toThrow('offline')
    await expect(getCachedApiResponse('user:/expenses:page-1', '/expenses', request)).resolves.toEqual({ id: 'expense-1' })

    expect(request).toHaveBeenCalledTimes(2)
  })

  it('invalidates resource lists and their details', async () => {
    const request = vi.fn().mockResolvedValue({ ok: true })

    await getCachedApiResponse('user:/receipts:list', '/receipts', request)
    await getCachedApiResponse('user:/receipts:detail', '/receipts/receipt-1', request)
    invalidateApiCache(['/receipts'])
    await getCachedApiResponse('user:/receipts:list', '/receipts', request)
    await getCachedApiResponse('user:/receipts:detail', '/receipts/receipt-1', request)

    expect(request).toHaveBeenCalledTimes(4)
  })

  it('does not cache a request that finished after invalidation', async () => {
    let resolveRequest: (value: { id: string }) => void = () => undefined
    const request = vi.fn()
      .mockImplementationOnce(() => new Promise<{ id: string }>((resolve) => {
        resolveRequest = resolve
      }))
      .mockResolvedValueOnce({ id: 'fresh-stock' })

    const staleRequest = getCachedApiResponse('user:/stock:list', '/stock', request)
    invalidateApiCache(['/stock'])
    resolveRequest({ id: 'stale-stock' })
    await staleRequest
    await getCachedApiResponse('user:/stock:list', '/stock', request)

    expect(request).toHaveBeenCalledTimes(2)
  })
})
