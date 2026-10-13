import { describe, expect, it } from 'vitest'
import type { MarketingBrand, MarketingCreation, MarketingTemplate } from '../types/marketing'
import { cloneMarketingInput, createFromMarketingTemplate, reopenMarketingCreation } from './marketingTemplates'

const template: MarketingTemplate = {
  id: 'empi-001', name: 'Revisão preventiva', category: 'Serviços', layout: 'photo-left',
  sourceUrl: 'https://www.instagram.com/p/reference/', previewUrl: '/marketing/references/reference.jpg',
  sourceKind: 'post', slide: 1, backgroundUrl: '/marketing/backgrounds/workshop.png',
  defaults: { title: 'A segurança começa nos detalhes.', description: 'Revisão com cuidado e precisão.',
    cta: 'Agende sua revisão', caption: 'Prepare seu carro com a EMPI.',
    primaryColor: '#07133D', secondaryColor: '#0066FF', kind: 'service', objective: 'service' },
}
const brand: MarketingBrand = {
  name: 'Minha oficina', contact: '(33) 1234-5678', logoImageId: 'logo-id',
  primaryColor: '#ff0000', secondaryColor: '#ffffff',
}

describe('reusable marketing templates', () => {
  it('creates a native-format snapshot with the template palette', () => {
    expect(createFromMarketingTemplate(template, brand)).toEqual({
      template: 'empi-001', format: 'original', ...template.defaults,
      priceCents: null, imageId: null, productId: null,
      brandName: 'Minha oficina', brandContact: '(33) 1234-5678', logoImageId: 'logo-id',
    })
  })
  it('isolates nested layers in catalog drafts and save submissions', () => {
    const layered = { ...template, layers: [{ id: 'headline', type: 'text' as const, x: 12, y: 24, width: 300, height: 70, text: 'Original', visible: true }] }
    const first = createFromMarketingTemplate(layered, brand)
    const second = createFromMarketingTemplate(layered, brand)
    const submitted = cloneMarketingInput(first)
    first.layers![0]!.text = 'Edited'
    expect(second.layers![0]!.text).toBe('Original')
    expect(layered.layers[0]!.text).toBe('Original')
    expect(submitted.layers![0]!.text).toBe('Original')
  })
  it('materializes missing legacy layers without changing saved text, photo, or format', () => {
    const creation: MarketingCreation = { ...createFromMarketingTemplate(template, brand), format: 'square', title: 'Minha oferta', imageId: 'uploaded-photo', id: 'saved', createdAt: '', updatedAt: '' }
    const layered = { ...template, layers: [{ id: 'photo', type: 'image' as const, x: 0, y: 0, width: 100, height: 100, visible: true }] }
    const reopened = reopenMarketingCreation(creation, layered)
    expect(reopened.format).toBe('square')
    expect(reopened.title).toBe('Minha oferta')
    expect(reopened.imageId).toBe('uploaded-photo')
    expect(reopened).not.toHaveProperty('id')
    reopened.layers![0]!.visible = false
    expect(layered.layers[0]!.visible).toBe(true)
    expect(creation.layers).toBeUndefined()
  })
  it('preserves edited and intentionally empty layers on reopening and duplication', () => {
    const creation: MarketingCreation = { ...createFromMarketingTemplate(template, brand), id: 'saved', createdAt: '', updatedAt: '', layers: [{ id: 'contact', type: 'text', x: 3, y: 4, width: 90, height: 30, text: 'Novo contato', visible: true }] }
    const reopened = reopenMarketingCreation(creation, { ...template, layers: [] })
    reopened.layers![0]!.text = 'Duplicated contact'
    expect(creation.layers![0]!.text).toBe('Novo contato')
    expect(reopenMarketingCreation({ ...creation, layers: [] }, { ...template, layers: creation.layers }).layers).toEqual([])
  })
  it('uses EMPI identity when brand details have not been configured', () => {
    const result = createFromMarketingTemplate(template, { ...brand, name: '  ', contact: '', logoImageId: null })
    expect(result.brandName).toBe('EMPI AutoCenter')
    expect(result.brandContact).toBe('(33) 98735-1922 · Rua Tupã, 587')
    expect(result.logoImageId).toBeNull()
    expect(createFromMarketingTemplate(template, null)).toEqual(result)
  })
  it('isolates each new creation from the catalog and other creations', () => {
    const first = createFromMarketingTemplate(template, brand)
    const second = createFromMarketingTemplate(template, brand)
    first.title = 'Outra campanha'
    first.primaryColor = '#ffffff'
    expect(second.title).toBe(template.defaults.title)
    expect(second.primaryColor).toBe('#07133D')
    expect(template.defaults.title).toBe('A segurança começa nos detalhes.')
  })
})
