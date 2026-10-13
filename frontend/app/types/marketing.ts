export interface MarketingBrand {
  name: string
  logoImageId: string | null
  primaryColor: string
  secondaryColor: string
  contact: string
}

export interface MarketingLayer {
  id: string
  type: 'text' | 'image'
  x: number
  y: number
  width: number
  height: number
  text?: string
  fontFamily?: 'Inter' | 'Roboto Condensed' | 'Oswald' | 'Bebas Neue' | 'Anton' | 'Dancing Script' | 'Knewave' | 'Kalam'
  fontSize?: number
  color?: string
  fontWeight?: 400 | 500 | 600 | 700 | 800 | 900
  italic?: boolean
  rotation?: number
  align?: 'left' | 'center' | 'right'
  imageId?: string | null
  fit?: 'cover' | 'contain'
  visible: boolean
  backgroundColor?: string
}

export interface MarketingCreationInput {
  layers?: MarketingLayer[]
  kind: 'product' | 'service' | 'offer'
  objective: 'promotion' | 'service' | 'launch'
  template: string
  format: 'feed' | 'story' | 'square' | 'original'
  title: string
  description: string
  priceCents: number | null
  cta: string
  primaryColor: string
  secondaryColor: string
  caption: string
  imageId: string | null
  productId: string | null
  brandName: string
  brandContact: string
  logoImageId: string | null
}

export interface MarketingCreation extends MarketingCreationInput {
  id: string
  createdAt: string
  updatedAt: string
}

export interface MarketingImage { id: string; mimeType: string }

export type MarketingTemplateLayout = 'institutional' | 'photo-left' | 'photo-top' | 'checklist' | 'promotion' | 'technical' | 'brand'

export interface MarketingTemplate {
  cutouts?: Record<string, 'color' | 'logo'>
  photoTreatment?: { side: 'left' | 'bottom' | 'none'; color: string; opacity: number }
  nativeWidth?: number
  nativeHeight?: number
  layers?: MarketingLayer[]
  id: string
  name: string
  category: string
  layout: MarketingTemplateLayout
  sourceUrl: string
  previewUrl: string
  sourceKind: 'post' | 'carousel' | 'reel'
  slide: number
  backgroundUrl: string
  defaults: Pick<MarketingCreationInput, 'title' | 'description' | 'cta' | 'caption' | 'primaryColor' | 'secondaryColor' | 'kind' | 'objective'>
}

export type MarketingTemplateDescriptor = MarketingTemplate
