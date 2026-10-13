import '@fontsource/dancing-script/400.css'
import '@fontsource/dancing-script/700.css'
import '@fontsource/knewave/400.css'
import '@fontsource/kalam/400.css'
import '@fontsource/kalam/700.css'
import '@fontsource/roboto-condensed/400.css'
import '@fontsource/roboto-condensed/700.css'
import '@fontsource/roboto-condensed/900.css'
import '@fontsource/roboto-condensed/400-italic.css'
import '@fontsource/roboto-condensed/700-italic.css'
import '@fontsource/roboto-condensed/900-italic.css'
import '@fontsource/oswald/400.css'
import '@fontsource/oswald/700.css'
import '@fontsource/bebas-neue/400.css'
import '@fontsource/anton/400.css'
import type { MarketingCreationInput, MarketingTemplate, MarketingLayer } from '../types/marketing'

export const marketingDimensions = (format: string) => ({ width: 1080, height: format === 'story' ? 1920 : format === 'square' ? 1080 : 1350 })

let fontsReady: Promise<unknown> | undefined
function prepareMarketingFonts() {
  if (!fontsReady) fontsReady = Promise.all(
    ['Inter', 'Roboto Condensed', 'Oswald', 'Bebas Neue', 'Anton', 'Dancing Script', 'Knewave', 'Kalam'].flatMap(family =>
      [400, 700, 900].flatMap(weight => ['', 'italic '].map(style => document.fonts.load(`${style}${weight} 64px "${family}"`)))
    )
  ).catch(error => { fontsReady = undefined; throw error })
  return fontsReady
}

const imageCache = new Map<string, Promise<HTMLImageElement>>()
async function loadImage(url?: string | null) {
  if (!url) return null
  if (!imageCache.has(url)) {
    const pending = new Promise<HTMLImageElement>((resolve, reject) => {
      const image = new Image()
      image.onload = () => resolve(image)
      image.onerror = () => { imageCache.delete(url); reject(new Error('Não foi possível carregar a imagem da arte.')) }
      image.src = url
    })
    imageCache.set(url, pending)
    // Bound decoded resources while retaining photos and logos for repeated local edits.
    if (imageCache.size > 32) imageCache.delete(imageCache.keys().next().value!)
  }
  return imageCache.get(url)!
}

export function contrastingText(hex: string) {
  const rgb = hex.replace('#', '')
  const values = [0, 2, 4].map(start => parseInt(rgb.slice(start, start + 2), 16) / 255)
    .map(value => value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4)
  return 0.2126 * values[0]! + 0.7152 * values[1]! + 0.0722 * values[2]! > 0.179 ? '#101820' : '#ffffff'
}

// Split long words too, so even pasted URLs stay within the artwork.
export function wrapMarketingText(text: string, width: number, measure: (value: string) => number): string[] {
  const lines: string[] = []
  for (const paragraph of text.split('\n')) {
    let line = ''
    for (const word of paragraph.split(/\s+/).filter(Boolean)) {
      if (line && measure(`${line} ${word}`) <= width) { line += ` ${word}`; continue }
      if (line) { lines.push(line); line = '' }
      for (const character of word) {
        if (line && measure(line + character) > width) { lines.push(line); line = '' }
        line += character
      }
    }
    lines.push(line)
  }
  return lines
}

function textBlock(ctx: CanvasRenderingContext2D, text: string, x: number, y: number, width: number, height: number, size: number, weight = 500) {
  if (width <= 0 || height <= 0) return
  let lines: string[] = []
  let fontSize = size
  do {
    ctx.font = `${weight} ${fontSize}px Inter, sans-serif`
    lines = wrapMarketingText(text, width, value => ctx.measureText(value).width)
    if (lines.length * fontSize * 1.22 <= height) break
    fontSize -= 2
  } while (fontSize > 22)
  const maxLines = Math.max(1, Math.floor(height / (fontSize * 1.22)))
  const visible = lines.slice(0, maxLines)
  if (lines.length > maxLines) {
    let last = visible[maxLines - 1] || ''
    while (last && ctx.measureText(last + '…').width > width) last = last.slice(0, -1)
    visible[maxLines - 1] = last + '…'
  }
  ctx.textBaseline = 'top'
  ctx.save(); ctx.beginPath(); ctx.rect(x, y, width, height); ctx.clip()
  visible.forEach((line, index) => ctx.fillText(line, x, y + index * fontSize * 1.22))
  ctx.restore()
}

function fitImage(ctx: CanvasRenderingContext2D, image: HTMLImageElement | HTMLCanvasElement, x: number, y: number, width: number, height: number, contain = false) {
  const ratio = contain ? Math.min(width / image.width, height / image.height) : Math.max(width / image.width, height / image.height)
  const w = image.width * ratio, h = image.height * ratio
  ctx.save(); ctx.beginPath(); ctx.rect(x, y, width, height); ctx.clip()
  ctx.drawImage(image, x + (width - w) / 2, y + (height - h) / 2, w, h)
  ctx.restore()
}

export function marketingLayerChanged(layer: MarketingLayer | undefined, original: MarketingLayer) {
  if (!layer) return true
  return JSON.stringify({ ...original, ...layer }) !== JSON.stringify(original)
}

function layerFont(layer: MarketingLayer, size = layer.fontSize || 48) {
  return `${layer.italic ? 'italic ' : ''}${layer.fontWeight || 400} ${size}px "${layer.fontFamily || 'Inter'}"`
}

function drawLayerText(ctx: CanvasRenderingContext2D, layer: MarketingLayer, width: number, height: number) {
  const x = layer.x * width, y = layer.y * height, w = layer.width * width, h = layer.height * height
  let size = layer.fontSize || 48
  let lines: string[] = []
  do {
    ctx.font = layerFont(layer, size)
    lines = wrapMarketingText(layer.text || '', w, value => ctx.measureText(value).width)
    if (lines.length * size * 1.12 <= h || size <= 8) break
    size = Math.max(8, size - 1)
  } while (true)
  ctx.save(); ctx.beginPath(); ctx.rect(x, y, w, h); ctx.clip()
  if (layer.backgroundColor) { ctx.fillStyle = layer.backgroundColor; ctx.fillRect(x, y, w, h) }
  if (layer.rotation) {
    ctx.translate(x + w / 2, y + h / 2)
    ctx.rotate(layer.rotation * Math.PI / 180)
    ctx.translate(-x - w / 2, -y - h / 2)
  }
  ctx.fillStyle = layer.color || '#ffffff'; ctx.textBaseline = 'top'
  ctx.textAlign = layer.align || 'left'
  const tx = x + (layer.align === 'center' ? w / 2 : layer.align === 'right' ? w : 0)
  lines.forEach((line, i) => ctx.fillText(line, tx, y + i * size * 1.12))
  ctx.restore()
}

const textPatchCache = new Map<string, { canvas: HTMLCanvasElement; x: number; y: number }>()
function cleanTextPatch(original: HTMLImageElement, clean: HTMLImageElement, layer: MarketingLayer, defaults: MarketingLayer[], width: number, height: number, key: string) {
  if (textPatchCache.has(key)) return textPatchCache.get(key)!
  const padding = 10
  const x = Math.max(0, Math.floor(layer.x * width) - padding), y = Math.max(0, Math.floor(layer.y * height) - padding)
  const w = Math.min(width - x, Math.ceil((layer.x + layer.width) * width) + padding - x)
  const h = Math.min(height - y, Math.ceil((layer.y + layer.height) * height) + padding - y)
  const patch = document.createElement('canvas'); patch.width = w; patch.height = h
  const source = document.createElement('canvas'); source.width = w; source.height = h
  const ctx = patch.getContext('2d', { willReadFrequently: true })!, src = source.getContext('2d', { willReadFrequently: true })!
  src.drawImage(original, x, y, w, h, 0, 0, w, h)
  const cw = clean.naturalWidth || clean.width, ch = clean.naturalHeight || clean.height
  ctx.drawImage(clean, x / width * cw, y / height * ch, w / width * cw, h / height * ch, 0, 0, w, h)
  const before = src.getImageData(0, 0, w, h), after = ctx.getImageData(0, 0, w, h)
  const nearby = defaults.filter(item => item.id !== layer.id && item.type === 'text' && item.visible && item.x * width < x + w && (item.x + item.width) * width > x && item.y * height < y + h && (item.y + item.height) * height > y)
  const correction = [0, 0, 0]; let samples = 0
  // Match the source background at its boundaries, without deriving the tone from old lettering.
  for (let p = 0; p < w * h; p++) {
    const px = p % w, py = Math.floor(p / w), i = p * 4
    if (px > 2 && py > 2 && px < w - 3 && py < h - 3) continue
    const difference = [0, 1, 2].map(channel => before.data[i + channel]! - after.data[i + channel]!)
    if (difference.some(value => Math.abs(value) > 55)) continue
    samples++; difference.forEach((value, channel) => { correction[channel]! += value })
  }
  for (let p = 0; p < w * h; p++) {
    const i = p * 4, px = p % w, py = Math.floor(p / w)
    const r = before.data[i]!, g = before.data[i + 1]!, b = before.data[i + 2]!
    const cr = after.data[i]!, cg = after.data[i + 1]!, cb = after.data[i + 2]!
    const blue = b > 90 && b > r + 35 && b > g + 15
    const yellow = r > 150 && g > 90 && b < 135
    const retainedDecoration = (blue && cb > 90 && cb > cr + 35 && cb > cg + 15) || (yellow && cr > 150 && cg > 90 && cb < 135)
    const preserveText = nearby.some(other => {
      const nx = (x + px) / width, ny = (y + py) / height
      if (nx < other.x || nx > other.x + other.width || ny < other.y || ny > other.y + other.height) return false
      const color = (other.color || '#ffffff').slice(1)
      const rgb = [0, 2, 4].map(channel => parseInt(color.slice(channel, channel + 2), 16))
      if (rgb[2]! > rgb[0]! + 40 && rgb[2]! > rgb[1]! + 15) return blue
      if (rgb[0]! > 150 && rgb[1]! > 100 && rgb[2]! < 135) return yellow
      if (rgb.reduce((a, value) => a + value, 0) / 3 > 150) return Math.min(r, g, b) > 110 && Math.max(r, g, b) - Math.min(r, g, b) < 75
      return Math.max(r, g, b) < 100
    })
    after.data[i + 3] = Math.round(255 * Math.min(1, Math.max(0, Math.min(px, py, w - 1 - px, h - 1 - py)) / padding))
    for (let channel = 0; channel < 3; channel++) {
      after.data[i + channel] = retainedDecoration || preserveText ? before.data[i + channel]! : Math.max(0, Math.min(255, after.data[i + channel]! + correction[channel]! / Math.max(1, samples)))
    }
  }
  ctx.putImageData(after, 0, 0)
  const result = { canvas: patch, x, y }
  textPatchCache.set(key, result)
  if (textPatchCache.size > 60) textPatchCache.delete(textPatchCache.keys().next().value!)
  return result
}

const textStencilCache = new Map<string, HTMLCanvasElement>()
function sourceTextStencil(original: HTMLImageElement, clean: HTMLImageElement, base: MarketingLayer, defaults: MarketingLayer[], width: number, height: number, key: string) {
  if (textStencilCache.has(key)) return textStencilCache.get(key)!
  const background = cleanTextPatch(original, clean, base, defaults, width, height, key + ':background')
  const canvas = document.createElement('canvas')
  canvas.width = Math.max(1, Math.ceil(base.width * width)); canvas.height = Math.max(1, Math.ceil(base.height * height))
  const ctx = canvas.getContext('2d', { willReadFrequently: true })!
  ctx.drawImage(original, base.x * width, base.y * height, base.width * width, base.height * height, 0, 0, canvas.width, canvas.height)
  const foreground = ctx.getImageData(0, 0, canvas.width, canvas.height)
  const comparison = document.createElement('canvas'); comparison.width = canvas.width; comparison.height = canvas.height
  const bg = comparison.getContext('2d', { willReadFrequently: true })!
  bg.drawImage(background.canvas, base.x * width - background.x, base.y * height - background.y, base.width * width, base.height * height, 0, 0, canvas.width, canvas.height)
  const pixels = bg.getImageData(0, 0, canvas.width, canvas.height)
  const hex = (base.color || '#ffffff').slice(1), color = [0, 2, 4].map(i => parseInt(hex.slice(i, i + 2), 16))
  const blue = color[2]! > color[0]! + 40 && color[2]! > color[1]! + 15
  const yellow = color[0]! > 150 && color[1]! > 100 && color[2]! < 135
  const light = color.reduce((a, value) => a + value, 0) / 3 > 150
  for (let i = 0; i < foreground.data.length; i += 4) {
    const r = foreground.data[i]!, g = foreground.data[i + 1]!, b = foreground.data[i + 2]!
    const difference = Math.abs(r - pixels.data[i]!) + Math.abs(g - pixels.data[i + 1]!) + Math.abs(b - pixels.data[i + 2]!)
    const matches = blue ? b > r + 30 && b > g + 10 : yellow ? r > 135 && g > 75 && b < 145 : light ? (r + g + b) / 3 > 55 && Math.max(r, g, b) - Math.min(r, g, b) < 105 : (r + g + b) / 3 < 170
    foreground.data[i + 3] = matches && difference > 55 ? Math.min(255, Math.round(difference / 95 * 255)) : 0
  }
  ctx.putImageData(foreground, 0, 0)
  textStencilCache.set(key, canvas)
  if (textStencilCache.size > 60) textStencilCache.delete(textStencilCache.keys().next().value!)
  return canvas
}

const cutoutCache = new Map<string, HTMLCanvasElement>()
function sourceCutout(original: HTMLImageElement, base: MarketingLayer, mode: 'color' | 'logo', width: number, height: number, key: string) {
  if (cutoutCache.has(key)) return cutoutCache.get(key)!
  const canvas = document.createElement('canvas')
  canvas.width = Math.max(1, Math.round(base.width * width)); canvas.height = Math.max(1, Math.round(base.height * height))
  const ctx = canvas.getContext('2d', { willReadFrequently: true })!
  ctx.drawImage(original, base.x * width, base.y * height, base.width * width, base.height * height, 0, 0, canvas.width, canvas.height)
  const pixels = ctx.getImageData(0, 0, canvas.width, canvas.height)
  for (let i = 0; i < pixels.data.length; i += 4) {
    const r = pixels.data[i]!, g = pixels.data[i + 1]!, b = pixels.data[i + 2]!
    const lum = (r + g + b) / 3
    const blue = b > r + (mode === 'logo' ? 5 : 35) && b > g + (mode === 'logo' ? 2 : 15)
    const yellow = r > 150 && g > 90 && b < 135
    const light = lum > (mode === 'logo' ? 75 : 125) && Math.max(r, g, b) - Math.min(r, g, b) < 75
    if (!blue && !yellow && !light) pixels.data[i + 3] = 0
  }
  ctx.putImageData(pixels, 0, 0)
  cutoutCache.set(key, canvas)
  if (cutoutCache.size > 32) cutoutCache.delete(cutoutCache.keys().next().value!)
  return canvas
}

async function renderCatalog(canvas: HTMLCanvasElement, creation: MarketingCreationInput, template: MarketingTemplate, layerImageUrls: Record<string, string>) {
  const original = await loadImage(template.previewUrl)
  if (!original) throw new Error('Não foi possível carregar a arte original.')
  const width = original.naturalWidth || original.width, height = original.naturalHeight || original.height
  const defaults = template.layers || []
  const layers = creation.layers ?? defaults
  const byId = new Map(layers.map(layer => [layer.id, layer]))
  const changedDefaults = defaults.filter(layer => marketingLayerChanged(byId.get(layer.id), layer))
  const changedText = changedDefaults.filter(layer => layer.type === 'text')
  const changedLayers = layers.filter(layer => !defaults.some(base => base.id === layer.id && !marketingLayerChanged(layer, base)))
  const photoChanged = changedDefaults.some(layer => layer.id === 'photo-main')
  const photoBase = defaults.find(layer => layer.id === 'photo-main')
  const photoLayer = byId.get('photo-main')
  const affectedText = photoChanged && photoBase ? layers.filter(layer => layer.type === 'text' && layer.visible &&
    layer.x < photoBase.x + photoBase.width && layer.x + layer.width > photoBase.x && layer.y < photoBase.y + photoBase.height && layer.y + layer.height > photoBase.y) : []
  const visibleText = [...new Map([...changedLayers.filter(layer => layer.visible && layer.type === 'text'), ...affectedText].map(layer => [layer.id, layer])).values()]
  const images = await Promise.all(changedLayers.filter(layer => layer.visible && layer.type === 'image' && layer.imageId).map(async layer => {
    const url = layerImageUrls[layer.imageId!]
    if (!url) throw new Error('Aguarde o carregamento da foto antes de exportar.')
    const image = await loadImage(url)
    if (!image) throw new Error('Não foi possível carregar a foto deste elemento.')
    return [layer.id, image] as const
  }))
  const clean = defaults.some(layer => layer.type === 'text') ? await loadImage(template.backgroundUrl) : null
  await prepareMarketingFonts()
  if (changedText.length && !clean) throw new Error('Não foi possível carregar as camadas deste modelo.')
  if (visibleText.length) await Promise.all([...new Set(visibleText.map(layer => layerFont(layer)))].map(font => document.fonts.load(font)))
  const dimensions = creation.format === 'original' ? { width, height } : marketingDimensions(creation.format)
  canvas.width = dimensions.width; canvas.height = dimensions.height
  const ctx = canvas.getContext('2d')
  if (!ctx) throw new Error('Seu navegador não conseguiu preparar a arte.')
  ctx.fillStyle = creation.primaryColor || '#050d2c'; ctx.fillRect(0, 0, canvas.width, canvas.height)
  const scale = Math.min(canvas.width / width, canvas.height / height)
  ctx.save(); ctx.translate((canvas.width - width * scale) / 2, (canvas.height - height * scale) / 2); ctx.scale(scale, scale)
  ctx.drawImage(original, 0, 0, width, height)
  // Only edited regions are reconstructed. Untouched artwork always uses its original pixels.
  const eraseDefaults = [...new Map([...changedDefaults, ...defaults.filter(base => affectedText.some(layer => layer.id === base.id))].map(layer => [layer.id, layer])).values()]
  for (const base of eraseDefaults) {
    const x = base.x * width, y = base.y * height, w = base.width * width, h = base.height * height
    if (base.type === 'text' && clean) {
      const patch = cleanTextPatch(original, clean, base, defaults, width, height, template.id + ':' + base.id + ':' + template.backgroundUrl)
      ctx.drawImage(patch.canvas, patch.x, patch.y)
    } else if (base.type === 'image') {
      ctx.fillStyle = base.backgroundColor || creation.primaryColor || '#050d2c'; ctx.fillRect(x, y, w, h)
    }
  }
  const loadedImages = new Map(images)
  const drawEditedImage = (layer: MarketingLayer) => {
    const photo = loadedImages.get(layer.id)
    if (photo) fitImage(ctx, photo, layer.x * width, layer.y * height, layer.width * width, layer.height * height, layer.fit === 'contain')
    else {
      const base = defaults.find(item => item.id === layer.id)
      if (base && template.cutouts?.[base.id]) {
        const cutout = sourceCutout(original, base, template.cutouts[base.id]!, width, height, template.id + ':' + base.id)
        fitImage(ctx, cutout, layer.x * width, layer.y * height, layer.width * width, layer.height * height, layer.fit === 'contain')
      } else if (base) ctx.drawImage(base.id === 'photo-main' && clean ? clean : original, base.x * width, base.y * height, base.width * width, base.height * height, layer.x * width, layer.y * height, layer.width * width, layer.height * height)
    }
  }
  if (photoChanged && photoLayer?.visible) drawEditedImage(photoLayer)
  if (photoChanged) {
    const treatment = template.photoTreatment
    if (treatment && photoLayer?.visible) {
      ctx.save()
      const gradient = treatment.side === 'bottom' ? ctx.createLinearGradient(0, height, 0, 0) : ctx.createLinearGradient(0, 0, width, 0)
      const rgba = (opacity: number) => `rgba(0,0,0,${opacity})`
      gradient.addColorStop(0, rgba(treatment.opacity))
      gradient.addColorStop(treatment.side === 'none' ? .999 : .42, rgba(treatment.side === 'none' ? treatment.opacity : treatment.opacity * .85))
      gradient.addColorStop(1, rgba(treatment.side === 'none' ? treatment.opacity : 0))
      ctx.fillStyle = gradient; ctx.fillRect(photoLayer.x * width, photoLayer.y * height, photoLayer.width * width, photoLayer.height * height)
      ctx.restore()
    }
    for (const base of defaults.filter(item => item.type === 'image' && item.id !== 'photo-main' && !marketingLayerChanged(byId.get(item.id), item))) {
      const layer = byId.get(base.id) || base
      if (!layer.visible) continue
      if (template.cutouts?.[base.id]) {
        const cutout = sourceCutout(original, base, template.cutouts[base.id]!, width, height, template.id + ':' + base.id)
        fitImage(ctx, cutout, layer.x * width, layer.y * height, layer.width * width, layer.height * height, layer.fit === 'contain')
      } else ctx.drawImage(original, base.x * width, base.y * height, base.width * width, base.height * height, layer.x * width, layer.y * height, layer.width * width, layer.height * height)
    }
  }
  changedLayers.filter(layer => layer.visible && layer.type === 'image' && layer.id !== 'photo-main').forEach(drawEditedImage)
  visibleText.forEach(layer => {
    const base = defaults.find(item => item.id === layer.id)
    if (photoChanged && base && clean && !marketingLayerChanged(layer, base)) {
      const stencil = sourceTextStencil(original, clean, base, defaults, width, height, template.id + ':' + base.id + ':' + template.backgroundUrl)
      ctx.drawImage(stencil, layer.x * width, layer.y * height, layer.width * width, layer.height * height)
    } else drawLayerText(ctx, layer, width, height)
  })
  ctx.restore()
}

export async function renderMarketingCanvas(canvas: HTMLCanvasElement, creation: MarketingCreationInput, imageUrl?: string | null, logoUrl?: string | null, template?: MarketingTemplate, layerImageUrls: Record<string, string> = {}) {
  if (template) return renderCatalog(canvas, creation, template, layerImageUrls)
  await Promise.all([document.fonts.load('700 64px Inter'), document.fonts.load('500 40px Inter'), document.fonts.ready])
  const [image, logo] = await Promise.all([loadImage(imageUrl), loadImage(logoUrl)])
  const { width, height } = marketingDimensions(creation.format)
  canvas.width = width; canvas.height = height
  const ctx = canvas.getContext('2d')
  if (!ctx) throw new Error('Seu navegador não conseguiu preparar a arte.')
  const bold = creation.template === 'bold'
  const primary = creation.primaryColor || '#164e63', secondary = creation.secondaryColor || '#f59e0b'
  ctx.fillStyle = bold ? primary : '#ffffff'; ctx.fillRect(0, 0, width, height)
  ctx.fillStyle = secondary; ctx.fillRect(0, 0, bold ? 24 : width, bold ? height : 20)
  const foreground = bold ? contrastingText(primary) : '#101820'
  ctx.fillStyle = foreground
  const left = 72, contentWidth = width - 144
  if (logo) fitImage(ctx, logo, left, 55, 120, 100, true)
  textBlock(ctx, creation.brandName || 'Seu negócio', logo ? 218 : left, 76, logo ? contentWidth - 146 : contentWidth, 70, 38, 700)
  const labels = { promotion: 'OFERTA ESPECIAL', service: 'CUIDE DO SEU VEÍCULO', launch: 'NOVIDADE POR AQUI' }
  ctx.fillStyle = bold ? contrastingText(primary) : primary
  textBlock(ctx, labels[creation.objective], left, 195, contentWidth, 60, 28, 700)
  textBlock(ctx, creation.title || 'Sua próxima divulgação', left, 270, contentWidth, height === 1080 ? 120 : 210, 76, 700)
  const photoY = height === 1080 ? 400 : 510, photoHeight = height === 1920 ? 670 : height === 1080 ? 200 : 360
  if (image) fitImage(ctx, image, left, photoY, contentWidth, photoHeight)
  else {
    ctx.save(); ctx.globalAlpha = 0.12; ctx.fillStyle = bold ? '#ffffff' : primary
    ctx.fillRect(left, photoY, contentWidth, photoHeight); ctx.restore()
    ctx.fillStyle = foreground
    textBlock(ctx, creation.kind === 'service' ? 'Qualidade e confiança para o seu veículo' : 'Uma oferta preparada para você', left + 50, photoY + photoHeight / 2 - 60, contentWidth - 100, 160, 46, 600)
  }
  ctx.fillStyle = foreground
  textBlock(ctx, creation.description, left, photoY + photoHeight + 30, contentWidth, 140, 36)
  if (creation.priceCents !== null && creation.priceCents !== undefined) {
    textBlock(ctx, new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(creation.priceCents / 100), left, height - 270, contentWidth, 80, 66, 700)
  }
  ctx.fillStyle = secondary; ctx.fillRect(left, height - 170, contentWidth, 80)
  ctx.fillStyle = contrastingText(secondary)
  textBlock(ctx, creation.cta || 'Entre em contato', left + 24, height - 150, contentWidth - 48, 50, 34, 700)
  ctx.fillStyle = foreground
  textBlock(ctx, creation.brandContact || '', left, height - 64, contentWidth, 46, 26)
}

export async function exportMarketingPng(creation: MarketingCreationInput, imageUrl?: string | null, logoUrl?: string | null, template?: MarketingTemplate, layerImageUrls: Record<string, string> = {}) {
  const canvas = document.createElement('canvas')
  await renderMarketingCanvas(canvas, creation, imageUrl, logoUrl, template, layerImageUrls)
  return new Promise<Blob>((resolve, reject) => canvas.toBlob(blob => blob ? resolve(blob) : reject(new Error('Não foi possível exportar a imagem.')), 'image/png'))
}
