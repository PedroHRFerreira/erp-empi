<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import MarketingTemplates from './MarketingTemplates.vue'
import MarketingLayers from './MarketingLayers.vue'
import type { MarketingBrand, MarketingCreationInput, MarketingTemplate } from '~/types/marketing'
import type { IPaginated, IStockItem } from '../../../server/contracts/types'
import { useMarketingStore } from '~/stores/useMarketingStore'
import { exportMarketingPng, renderMarketingCanvas } from '~/utils/marketingCanvas'

const props = defineProps<{ modelValue: MarketingCreationInput; brand?: MarketingBrand | null; startStep?: number; flowKey?: number }>()
const emit = defineEmits<{ 'update:modelValue': [MarketingCreationInput]; dirty: []; busy: [boolean]; selectTemplate: [MarketingTemplate]; step: [number] }>()
const step = ref(props.startStep || 1)
watch(() => props.flowKey, () => { step.value = props.startStep || 1 })
watch(step, value => emit('step', value), { immediate: true })
function goTo(value: number) { if (!exporting.value && !uploading.value && !layerUploading.value) { step.value = value; selectedLayerId.value = '' } }
const store = useMarketingStore()
const selectedTemplate = computed(() => store.templates.find(item => item.id === props.modelValue.template))
const simpleText = computed(() => {
  const layers = props.modelValue.layers || selectedTemplate.value?.layers || []
  return layers.find(layer => layer.id === selectedLayerId.value && layer.type === 'text') || layers.filter(layer => layer.type === 'text' && layer.visible).sort((a, b) => (b.fontSize || 0) - (a.fontSize || 0))[0]
})
function updateSimpleText(event: Event) {
  const layer = simpleText.value
  if (!layer) return
  selectedLayerId.value = layer.id
  patch({ layers: (props.modelValue.layers || selectedTemplate.value?.layers || []).map(item => item.id === layer.id ? { ...item, text: (event.target as HTMLTextAreaElement).value } : { ...item }) })
}
async function uploadSimplePhoto(event: Event) {
  if (!selectedTemplate.value) { await upload(event); return }
  const input = event.target as HTMLInputElement; const file = input.files?.[0]; input.value = ''
  if (!file || uploading.value) return
  error.value = ''
  if (!['image/png', 'image/jpeg'].includes(file.type) || file.size > 5 * 1024 * 1024) { error.value = 'Escolha uma imagem PNG ou JPEG de até 5 MB.'; return }
  const templateId = props.modelValue.template
  const layers = props.modelValue.layers || selectedTemplate.value.layers || []
  const photo = layers.find(item => item.id === 'photo-main') || layers.find(item => item.type === 'image' && item.id !== 'logo-brand' && !item.id.startsWith('decor-'))
  if (!photo) { error.value = 'Este modelo não possui uma área de foto. Use Mais opções para adicionar uma.'; return }
  uploading.value = true
  try {
    const result = await store.uploadImage(file)
    if (result && props.modelValue.template === templateId) patch({ layers: (props.modelValue.layers || layers).map(item => item.id === photo.id ? { ...item, imageId: result.id, visible: true } : { ...item }) })
    else if (!result) { error.value = store.error; store.error = '' }
  } catch { error.value = 'Não foi possível enviar a imagem. Sua criação foi mantida.' }
  finally { uploading.value = false }
}
const canvas = ref<HTMLCanvasElement | null>(null)
const layerRef = ref<InstanceType<typeof MarketingLayers> | null>(null)
const selectedLayerId = ref('')
const layerBounds = computed(() => {
  const template = selectedTemplate.value
  const layer = props.modelValue.layers?.find(item => item.id === selectedLayerId.value) || template?.layers?.find(item => item.id === selectedLayerId.value)
  if (!template || !layer || !layer.visible) return null
  const width = template.nativeWidth || 1080; const height = template.nativeHeight || 1080
  const formats = { square: [1080, 1080], feed: [1080, 1350], story: [1080, 1920], original: [width, height] }
  const [outputWidth, outputHeight] = formats[props.modelValue.format] as [number, number]
  const scale = Math.min(outputWidth / width, outputHeight / height)
  return { left: `${((outputWidth - width * scale) / 2 + layer.x * width * scale) / outputWidth * 100}%`, top: `${((outputHeight - height * scale) / 2 + layer.y * height * scale) / outputHeight * 100}%`, width: `${layer.width * width * scale / outputWidth * 100}%`, height: `${layer.height * height * scale / outputHeight * 100}%`, transform: `rotate(${layer.rotation || 0}deg)` }
})
function selectAt(event: MouseEvent) {
  if (step.value !== 2 || !canvas.value || !selectedTemplate.value || exporting.value || layerUploading.value) return
  const rect = canvas.value.getBoundingClientRect()
  const template = selectedTemplate.value; const width = template.nativeWidth || 1080; const height = template.nativeHeight || 1080
  const scale = Math.min(canvas.value.width / width, canvas.value.height / height)
  const x = ((event.clientX - rect.left) / rect.width * canvas.value.width - (canvas.value.width - width * scale) / 2) / scale
  const y = ((event.clientY - rect.top) / rect.height * canvas.value.height - (canvas.value.height - height * scale) / 2) / scale
  const layers = [...(props.modelValue.layers || template.layers || [])].reverse()
  const contains = (layer: typeof layers[number]) => {
    const centerX = (layer.x + layer.width / 2) * width; const centerY = (layer.y + layer.height / 2) * height
    const angle = -(layer.rotation || 0) * Math.PI / 180
    const localX = Math.cos(angle) * (x - centerX) - Math.sin(angle) * (y - centerY)
    const localY = Math.sin(angle) * (x - centerX) + Math.cos(angle) * (y - centerY)
    return layer.visible && Math.abs(localX) <= layer.width * width / 2 && Math.abs(localY) <= layer.height * height / 2
  }
  const found = layers.find(layer => layer.type === 'text' && contains(layer)) || layers.find(contains)
  if (found) layerRef.value?.select(found.id)
}
const products = ref<IStockItem[]>([])
const productSearch = ref('')
const loadingProducts = ref(false)
const refreshingProduct = ref(false)
const uploading = ref(false)
const exporting = ref(false)
const layerUploading = ref(false)
const rendering = ref(false)
watch([uploading, layerUploading, exporting, refreshingProduct], () => emit('busy', uploading.value || layerUploading.value || exporting.value || refreshingProduct.value), { flush: 'sync' })
const error = ref('')
const message = ref('')
const imageUrl = ref<string | null>(null)
const logoUrl = ref<string | null>(null)
const layerImageUrls = ref<Record<string, string>>({})
const pendingLayerImages = new Set<string>()
async function loadLayerImages() {
  const ids = new Set((props.modelValue.layers || []).map(layer => layer.imageId).filter((id): id is string => !!id))
  for (const [id, url] of Object.entries(layerImageUrls.value)) if (!ids.has(id)) { URL.revokeObjectURL(url); const next = { ...layerImageUrls.value }; delete next[id]; layerImageUrls.value = next }
  await Promise.all([...ids].filter(id => !layerImageUrls.value[id] && !pendingLayerImages.has(id)).map(async id => {
    pendingLayerImages.add(id)
    try {
      const url = await store.loadImage(id)
      if (!url) { error.value = store.error || 'Não foi possível carregar uma foto da arte.'; return }
      if (disposed || !(props.modelValue.layers || []).some(layer => layer.imageId === id)) { URL.revokeObjectURL(url); return }
      layerImageUrls.value = { ...layerImageUrls.value, [id]: url }
    } finally { pendingLayerImages.delete(id) }
  }))
}
let disposed = false
let renderSequence = 0
let imageSequence = 0
let logoSequence = 0
let timer: ReturnType<typeof setTimeout> | undefined
const matchingProducts = computed(() => products.value.filter(item => item.name.toLocaleLowerCase('pt-BR').includes(productSearch.value.toLocaleLowerCase('pt-BR'))))
const price = computed(() => props.modelValue.priceCents === null ? '' : (props.modelValue.priceCents / 100).toFixed(2).replace('.', ','))
const captions = computed(() => {
  const item = props.modelValue
  const title = item.title.trim() || 'Nossa novidade'
  const amount = item.priceCents === null ? '' : ` por ${new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(item.priceCents / 100)}`
  const contact = item.brandContact ? ` Fale conosco: ${item.brandContact}.` : ''
  const starts = { promotion: ['Aproveite esta oferta', 'Uma oportunidade para você', 'Oferta especial'], service: ['Cuide do seu veículo', 'Conte com nossa equipe', 'Qualidade em cada atendimento'], launch: ['Novidade por aqui', 'Acabou de chegar', 'Conheça nossa novidade'] }
  return starts[item.objective].map(start => `${start}: ${title}${amount}. ${item.description.trim()} ${item.cta.trim() || 'Entre em contato!'}${contact}`.replace(/\s+/g, ' ').trim())
})
function patch(values: Partial<MarketingCreationInput>) { emit('update:modelValue', { ...props.modelValue, ...values }); emit('dirty') }
function update<K extends keyof MarketingCreationInput>(field: K, event: Event) { patch({ [field]: (event.target as HTMLInputElement).value } as Pick<MarketingCreationInput, K>) }
function updateKind(event: Event) { const kind = (event.target as HTMLSelectElement).value as MarketingCreationInput['kind']; patch({ kind, productId: kind === 'product' ? props.modelValue.productId : null }) }
function updatePrice(event: Event) {
  const value = (event.target as HTMLInputElement).value.trim().replace(',', '.')
  if (!value) { error.value = ''; patch({ priceCents: null }); return }
  const number = Number(value)
  if (Number.isFinite(number) && number >= 0 && number <= 999999999.99) { error.value = ''; patch({ priceCents: Math.round(number * 100) }) }
  else error.value = 'Informe um preço válido entre R$ 0,00 e R$ 999.999.999,99. O preço anterior foi mantido.'
}
async function loadProducts() {
  loadingProducts.value = true
  let offset = 0
  const all: IStockItem[] = []
  try {
    while (!disposed) {
      const { data, status } = await useApiFetch<IPaginated<IStockItem>>('/stock', { query: { limit: 100, offset } })
      if (status.value === 'error' || !data.value) throw new Error('Não foi possível carregar os produtos. Tente novamente.')
      const rows = data.value.data || []
      all.push(...rows.filter(item => item.active))
      offset += rows.length
      if (!rows.length || offset >= data.value.total) break
    }
    if (!disposed) products.value = all
    return true
  } catch (cause) { error.value = (cause as Error).message; return false }
  finally { loadingProducts.value = false }
}
function selectProduct(event: Event) {
  const id = (event.target as HTMLSelectElement).value
  const item = products.value.find(product => product.id === id)
  if (item) patch({ kind: 'product', productId: id, title: item.name.slice(0, 140), description: item.description.slice(0, 500), priceCents: item.resalePriceCents })
  else patch({ productId: null })
}
async function refreshProduct() {
  const id = props.modelValue.productId
  if (!id || refreshingProduct.value) return
  refreshingProduct.value = true
  error.value = ''
  try {
    invalidateApiCache(['/stock'])
    if (!await loadProducts() || disposed || props.modelValue.productId !== id) return
    const item = products.value.find(product => product.id === id)
    if (item) patch({ title: item.name.slice(0, 140), description: item.description.slice(0, 500), priceCents: item.resalePriceCents })
    else error.value = 'Este produto não está disponível no estoque. Os dados salvos foram mantidos.'
  } finally { refreshingProduct.value = false }
}
function applyBrand() { if (props.brand) patch({ brandName: props.brand.name, brandContact: props.brand.contact, logoImageId: props.brand.logoImageId, primaryColor: props.brand.primaryColor, secondaryColor: props.brand.secondaryColor }) }
async function loadAsset(id: string | null, target: typeof imageUrl, sequence: number, kind: 'image' | 'logo') {
  const url = id ? await store.loadImage(id) : null
  if (disposed || sequence !== (kind === 'image' ? imageSequence : logoSequence)) { if (url) URL.revokeObjectURL(url); return }
  if (target.value) URL.revokeObjectURL(target.value)
  target.value = url
  if (id && !url) error.value = store.error
}
watch(() => (props.modelValue.layers || []).map(layer => layer.imageId).join(','), () => { void loadLayerImages() }, { immediate: true })
watch(() => props.modelValue.imageId, id => { void loadAsset(id, imageUrl, ++imageSequence, 'image') }, { immediate: true })
watch(() => props.modelValue.logoImageId, id => { void loadAsset(id, logoUrl, ++logoSequence, 'logo') }, { immediate: true })
async function upload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]; input.value = ''
  if (!file) return
  error.value = ''
  if (!['image/png', 'image/jpeg'].includes(file.type) || file.size > 5 * 1024 * 1024) { error.value = 'Escolha uma imagem PNG ou JPEG de até 5 MB.'; return }
  uploading.value = true
  try { const result = await store.uploadImage(file); if (result) patch({ imageId: result.id }); else error.value = store.error }
  catch { error.value = 'Não foi possível enviar a imagem. Sua criação foi mantida.' }
  finally { uploading.value = false }
}
function queueRender() {
  const sequence = ++renderSequence
  clearTimeout(timer)
  timer = setTimeout(async () => {
    if (!canvas.value || disposed) return
    rendering.value = true
    try {
      const buffer = document.createElement('canvas')
      await renderMarketingCanvas(buffer, { ...props.modelValue }, imageUrl.value, logoUrl.value, selectedTemplate.value, layerImageUrls.value)
      if (sequence !== renderSequence || disposed || !canvas.value) return
      canvas.value.width = buffer.width; canvas.value.height = buffer.height
      canvas.value.getContext('2d')?.drawImage(buffer, 0, 0)
    } catch (cause) { if (sequence === renderSequence) error.value = (cause as Error).message }
    finally { if (sequence === renderSequence) rendering.value = false }
  }, 120)
}
watch([() => props.modelValue, imageUrl, logoUrl, selectedTemplate, layerImageUrls], queueRender, { deep: true })
async function download() {
  if (exporting.value || uploading.value || layerUploading.value) return
  error.value = ''; message.value = ''; exporting.value = true
  try {
    if ((props.modelValue.layers || []).some(layer => layer.imageId && !layerImageUrls.value[layer.imageId])) throw new Error('Aguarde as fotos carregarem. Se houver uma falha, envie a foto novamente antes de exportar.')
    if ((props.modelValue.imageId && !imageUrl.value) || (props.modelValue.logoImageId && !logoUrl.value)) throw new Error('Aguarde as imagens carregarem antes de baixar a arte.')
    const blob = await exportMarketingPng({ ...props.modelValue }, imageUrl.value, logoUrl.value, selectedTemplate.value, layerImageUrls.value)
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a'); link.href = url; link.download = `marketing-${props.modelValue.format}.png`; link.click()
    setTimeout(() => URL.revokeObjectURL(url), 10000)
    message.value = 'Arte pronta para download.'
  } catch (cause) { error.value = (cause as Error).message || 'Não foi possível exportar. Sua criação foi mantida.' }
  finally { exporting.value = false }
}
async function copyCaption() {
  error.value = ''; message.value = ''
  try { await navigator.clipboard.writeText(props.modelValue.caption); message.value = 'Legenda copiada.' }
  catch { error.value = 'Não foi possível copiar automaticamente. Selecione a legenda e copie manualmente.' }
}
onMounted(() => { void loadProducts(); queueRender() })
onBeforeUnmount(() => { disposed = true; clearTimeout(timer); ++renderSequence; for (const url of Object.values(layerImageUrls.value)) URL.revokeObjectURL(url); if (imageUrl.value) URL.revokeObjectURL(imageUrl.value); if (logoUrl.value) URL.revokeObjectURL(logoUrl.value) })
</script>

<template>
  <div class="marketing-flow">
    <nav class="marketing-flow__steps" aria-label="Etapas da criação"><button v-for="(label, index) in ['Modelo', 'Personalizar', 'Revisar']" :key="label" type="button" :aria-current="step === index + 1 ? 'step' : undefined" :disabled="exporting || uploading || layerUploading || (index === 2 && !modelValue.title.trim())" @click="goTo(index + 1)">{{ index + 1 }}. {{ label }}</button></nav>
    <section v-show="step === 1" aria-label="Escolher modelo">
      <h2>Escolha seu ponto de partida</h2><p>Use uma arte da sua galeria ou comece com uma composição simples.</p>
      <div class="marketing-flow__choices"><button class="button button--secondary" type="button" @click="patch({ template: 'bold', layers: undefined, format: 'feed' }); goTo(2)">Composição Destaque</button><button class="button button--secondary" type="button" @click="patch({ template: 'minimal', layers: undefined, format: 'feed' }); goTo(2)">Composição Essencial</button><button class="button" type="button" @click="goTo(2)">Continuar com o modelo atual</button></div>
      <MarketingTemplates @select="emit('selectTemplate', $event)" @retry="store.loadTemplates(true)" />
    </section>
    <div v-show="step !== 1" class="marketing-editor">
    <section class="marketing-editor__fields" aria-label="Editar criação">
      <h2>{{ step === 2 ? 'Personalizar' : 'Revisar' }}</h2>

      <div v-show="step === 2">
      <template v-if="!selectedTemplate">
      <label class="field"><span>Título</span><input :value="modelValue.title" maxlength="140" required @input="update('title', $event)" /></label>
      <label class="field"><span>Foto da criação</span><input type="file" accept="image/png,image/jpeg" :disabled="uploading" @change="upload" /><small>PNG ou JPEG de até 5 MB. {{ uploading ? 'Enviando imagem...' : '' }}</small></label>
      <details class="marketing-editor__more"><summary>Mais opções</summary><div class="marketing-editor__options">
      <div class="marketing-editor__row">
        <label class="field"><span>Tipo de conteúdo</span><select :value="modelValue.kind" @change="updateKind"><option value="product">Produto do estoque</option><option value="service">Serviço</option><option value="offer">Oferta combinada</option></select></label>
        <label class="field"><span>Objetivo</span><select :value="modelValue.objective" @change="update('objective', $event)"><option value="promotion">Promoção</option><option value="service">Divulgação de serviço</option><option value="launch">Lançamento</option></select></label>
      </div>
      <fieldset v-if="modelValue.kind === 'product'" class="marketing-editor__product">
        <legend>Selecionar produto</legend>
        <label class="field"><span>Buscar no estoque</span><input v-model="productSearch" type="search" placeholder="Nome do produto" /></label>
        <label class="field"><span>Produto</span><select :value="modelValue.productId || ''" :disabled="loadingProducts" @change="selectProduct"><option value="">{{ loadingProducts ? 'Carregando produtos...' : 'Selecione um produto' }}</option><option v-if="modelValue.productId && !matchingProducts.some(item => item.id === modelValue.productId)" :value="modelValue.productId">Produto da criação salva</option><option v-for="item in matchingProducts" :key="item.id" :value="item.id">{{ item.name }}</option></select></label>
        <button class="button button--secondary" type="button" :disabled="loadingProducts" @click="modelValue.productId ? refreshProduct() : loadProducts()">{{ modelValue.productId ? 'Atualizar dados pelo estoque' : 'Recarregar produtos' }}</button>
        <small>Nome e preço ficam salvos nesta arte. Atualize pelo estoque quando desejar.</small>
      </fieldset>
      <div class="marketing-editor__row">
        <label class="field"><span>Composição</span><select :value="modelValue.template" @change="update('template', $event)"><optgroup label="Composições básicas"><option value="bold">Destaque</option><option value="minimal">Essencial</option></optgroup><optgroup v-if="store.templates.length" label="Modelos EMPI"><option v-for="item in store.templates" :key="item.id" :value="item.id">{{ item.name }}</option></optgroup></select></label>
        <label class="field"><span>Formato</span><select :value="modelValue.format" @change="update('format', $event)"><option value="square">Quadrado · 1080 × 1080</option><option value="feed">Feed · 1080 × 1350</option><option value="story">Story · 1080 × 1920</option></select></label>
      </div>

      <label class="field"><span>Descrição</span><textarea :value="modelValue.description" maxlength="500" rows="3" @input="update('description', $event)" /></label>
      <div class="marketing-editor__row">
        <label class="field"><span>Preço em reais (opcional)</span><input :value="price" inputmode="decimal" placeholder="Sem preço" @change="updatePrice" /></label>
        <label class="field"><span>Chamada para ação</span><input :value="modelValue.cta" maxlength="100" placeholder="Entre em contato" @input="update('cta', $event)" /></label>
      </div>

      <button v-if="modelValue.imageId" type="button" class="button button--secondary" :disabled="uploading" @click="patch({ imageId: null })">Remover foto</button>
      <details><summary>Ajustes avançados: identidade e cores</summary><fieldset class="marketing-editor__brand">
        <legend>Identidade nesta criação</legend>
        <button class="button button--secondary" type="button" :disabled="!brand" @click="applyBrand">Aplicar identidade visual salva</button>
        <label class="field"><span>Nome do negócio</span><input :value="modelValue.brandName" maxlength="140" @input="update('brandName', $event)" /></label>
        <label class="field"><span>Contato</span><input :value="modelValue.brandContact" maxlength="200" @input="update('brandContact', $event)" /></label>
        <div class="marketing-editor__row"><label class="field"><span>Cor principal</span><input type="color" :value="modelValue.primaryColor" @input="update('primaryColor', $event)" /></label><label class="field"><span>Cor de destaque</span><input type="color" :value="modelValue.secondaryColor" @input="update('secondaryColor', $event)" /></label></div>
      </fieldset></details>
      </div></details>
      </template>
      <template v-if="selectedTemplate">
        <label v-if="simpleText" class="field"><span>Texto principal</span><textarea :value="simpleText.text" maxlength="500" rows="3" @input="updateSimpleText" /></label>
        <small v-if="simpleText">Toque na arte para editar outro texto.</small>
        <label class="field"><span>Foto da criação</span><input type="file" accept="image/png,image/jpeg" :disabled="uploading || layerUploading" @change="uploadSimplePhoto" /><small>PNG ou JPEG de até 5 MB.</small></label>
        <details class="marketing-editor__more"><summary>Mais opções</summary><div class="marketing-editor__options">
        <label class="field"><span>Nome da criação</span><input :value="modelValue.title" maxlength="140" @input="update('title', $event)" /></label>
        <label class="field"><span>Formato</span><select :value="modelValue.format" :disabled="exporting || layerUploading" @change="update('format', $event)"><option value="original">Original · {{ selectedTemplate.nativeWidth || 1080 }} × {{ selectedTemplate.nativeHeight || 1080 }}</option><option value="square">Quadrado · 1080 × 1080</option><option value="feed">Feed · 1080 × 1350</option><option value="story">Story · 1080 × 1920</option></select></label>

        <details><summary>Inserir dados de um produto do estoque (opcional)</summary><fieldset class="marketing-editor__product">
          <legend>Dados para inserir na arte</legend>
          <label class="field"><span>Buscar produto</span><input v-model="productSearch" type="search" placeholder="Nome no estoque" /></label>
          <label class="field"><span>Produto</span><select :value="modelValue.productId || ''" :disabled="loadingProducts || exporting || layerUploading" @change="selectProduct"><option value="">{{ loadingProducts ? 'Carregando produtos…' : 'Selecionar produto' }}</option><option v-if="modelValue.productId && !matchingProducts.some(item => item.id === modelValue.productId)" :value="modelValue.productId">Produto salvo nesta criação</option><option v-for="item in matchingProducts" :key="item.id" :value="item.id">{{ item.name }}</option></select></label>
          <label class="field"><span>Preço em reais (opcional)</span><input :value="price" inputmode="decimal" placeholder="Sem preço" @change="updatePrice" /></label>
          <button v-if="modelValue.productId" class="button button--secondary" type="button" :disabled="loadingProducts || refreshingProduct || exporting || layerUploading" @click="refreshProduct">Atualizar dados pelo estoque</button>
          <small>Use os botões do elemento para inserir nome e preço. O estoque não é alterado.</small>
        </fieldset></details>
        <MarketingLayers ref="layerRef" @selection="selectedLayerId = $event" :model-value="modelValue.layers || selectedTemplate.layers || []" :defaults="selectedTemplate.layers || []" :disabled="exporting || refreshingProduct" :title="modelValue.title" :price="modelValue.priceCents === null ? '' : new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(modelValue.priceCents / 100)" :contact="brand?.contact || modelValue.brandContact" @update:model-value="patch({ layers: $event })" @busy="layerUploading = $event" />
        </div></details>
      </template>
      </div>
      <fieldset v-show="step === 3" class="marketing-editor__captions">
        <legend>Legenda</legend>
        <div class="marketing-editor__suggestions"><button v-for="(caption, index) in captions" :key="index" type="button" class="button button--secondary" @click="patch({ caption })">Sugestão {{ index + 1 }}</button></div>
        <label class="field"><span>Texto para publicar</span><textarea :value="modelValue.caption" maxlength="2200" rows="5" @input="update('caption', $event)" /></label>
        <button class="button button--secondary" type="button" :disabled="!modelValue.caption" @click="copyCaption">Copiar legenda</button>
      </fieldset>
    </section>
    <section class="marketing-editor__preview" aria-label="Prévia da arte">
      <h2>Prévia</h2>

      <details v-if="selectedTemplate" class="marketing-editor__reference"><summary>Ver referência</summary><p><a :href="selectedTemplate.previewUrl" target="_blank" rel="noopener noreferrer">Ver arte original</a> · <a :href="selectedTemplate.sourceUrl" target="_blank" rel="noopener noreferrer">Publicação no Instagram</a></p></details>
      <div class="marketing-editor__canvas"><div class="marketing-editor__selection-frame"><canvas ref="canvas" @click="selectAt" role="img" :aria-label="`Arte: ${modelValue.title || 'Sua próxima divulgação'}`" /><span v-if="step === 2 && layerBounds" class="marketing-editor__selection" :style="layerBounds" aria-hidden="true" /></div></div>
      <p class="marketing-editor__status" aria-live="polite">{{ exporting ? 'Preparando imagem para download...' : rendering ? 'Atualizando prévia...' : message || 'Prévia atualizada' }}</p>
      <button v-show="step === 3" class="button button--primary" type="button" :disabled="exporting || uploading || layerUploading" @click="download">{{ exporting ? 'Gerando PNG...' : 'Baixar arte em PNG' }}</button>
      <small v-show="step === 3">Use a imagem e a legenda para publicar.</small>
    </section>
    <p v-if="error" class="form-error marketing-editor__error" role="alert">{{ error }}</p>
    </div>
    <div v-if="step !== 1" class="marketing-flow__navigation"><button class="button button--secondary" type="button" :disabled="exporting || uploading || layerUploading" @click="goTo(step - 1)">Voltar</button><button v-if="step === 2" class="button" type="button" :disabled="exporting || uploading || layerUploading || !modelValue.title.trim()" @click="goTo(3)">Revisar e finalizar</button></div>
  </div>
</template>

<style scoped>
.marketing-flow__steps, .marketing-flow__choices, .marketing-flow__navigation { display: flex; flex-direction: row; flex-wrap: wrap; gap: .75rem; margin: 1rem 0; }
.marketing-flow__steps button { flex: 1; padding: .8rem; border: 1px solid var(--watt-border); border-radius: .5rem; background: var(--watt-surface); color: var(--watt-text); cursor: pointer; }
.marketing-flow__steps button[aria-current="step"] { background: var(--watt-data); color: var(--watt-on-accent); border-color: var(--watt-data); }
.marketing-editor__fields > div { display: grid; gap: 18px; }
.marketing-editor__fields > h2 { margin: 0; }
.marketing-flow__steps button:focus-visible { outline: 2px solid var(--watt-data); outline-offset: 3px; }
.marketing-flow__navigation { justify-content: space-between; }
.marketing-editor__options { display: grid; gap: 18px; padding-top: 12px; }
summary { cursor: pointer; padding: .75rem 0; }

@media (max-width: 1199px) {
  .marketing-editor__preview { position: sticky; top: 96px; z-index: 3; padding: 10px; gap: 6px; }
  .marketing-editor__preview h2 { font-size: 14px; }
  .marketing-editor__canvas { padding: 4px; }
  .marketing-editor__selection-frame { width: fit-content; max-width: 100%; margin: 0 auto; }
  .marketing-editor__selection-frame canvas { width: auto; max-width: 100%; height: auto; max-height: min(220px, 28vh); }
  .marketing-editor__reference { display: none; }
  .marketing-editor__status { min-height: 16px; }
}
.marketing-editor__selection-frame { position: relative; line-height: 0; }
.marketing-editor__selection-frame canvas { display: block; cursor: pointer; }
.marketing-editor__selection { position: absolute; pointer-events: none; border: 2px solid #60a5fa; box-shadow: 0 0 0 1px #ffffff; transform-origin: center; }
</style>
