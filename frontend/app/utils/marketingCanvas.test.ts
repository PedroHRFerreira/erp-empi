import { describe, expect, it, vi } from 'vitest'
import { contrastingText, marketingDimensions, wrapMarketingText, renderMarketingCanvas } from './marketingCanvas'
import type { MarketingCreationInput, MarketingTemplate } from '../types/marketing'

describe('marketing artwork layout', () => {
  it('exports feed and story at their native dimensions', () => {
    expect(marketingDimensions('feed')).toEqual({ width: 1080, height: 1350 })
    expect(marketingDimensions('story')).toEqual({ width: 1080, height: 1920 })
    expect(marketingDimensions('square')).toEqual({ width: 1080, height: 1080 })
  })
  it('wraps long words and paragraphs without exceeding the available width', () => {
    const lines = wrapMarketingText('Uma oferta especial\nhttps://um-link-muito-longo.example', 10, value => value.length)
    expect(lines.every(line => line.length <= 10)).toBe(true)
    expect(lines.join('').replaceAll(' ', '')).toBe('Umaofertaespecialhttps://um-link-muito-longo.example')
  })
  it('keeps text readable on light and dark brand colors', () => {
    expect(contrastingText('#ffffff')).toBe('#101820')
    expect(contrastingText('#000000')).toBe('#ffffff')
    expect(contrastingText('#f59e0b')).toBe('#101820')
  })
})

describe('original Instagram artwork', () => {
  it('preserves original pixels and dimensions rather than generating an unrelated layout', async () => {
    const requested: string[] = []
    vi.stubGlobal('document', { fonts: { load: async () => [] } })
    vi.stubGlobal('Image', class {
      naturalWidth = 1254; naturalHeight = 1254
      onload?: () => void
      set src(url: string) { requested.push(url); queueMicrotask(() => this.onload?.()) }
    })
    const ctx = { drawImage: vi.fn(), fillRect: vi.fn(), save: vi.fn(), restore: vi.fn(), translate: vi.fn(), scale: vi.fn(), fillStyle: '' }
    const canvas = { width: 0, height: 0, getContext: () => ctx } as unknown as HTMLCanvasElement
    try {
      const template = { previewUrl: '/original-fidelity-test.jpg', backgroundUrl: '/generic-background.png' } as MarketingTemplate
      await renderMarketingCanvas(canvas, { format: 'original' } as MarketingCreationInput, '/replacement.jpg', '/logo.jpg', template)
      expect(requested).toEqual(['/original-fidelity-test.jpg'])
      expect([canvas.width, canvas.height]).toEqual([1254, 1254])
      expect(ctx.drawImage).toHaveBeenCalledWith(expect.anything(), 0, 0, 1254, 1254)
    } finally { vi.unstubAllGlobals() }
  })
})
