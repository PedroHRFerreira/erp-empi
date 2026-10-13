<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useMarketingStore } from '~/stores/useMarketingStore'
import { renderMarketingCanvas } from '~/utils/marketingCanvas'
import type { MarketingCreation } from '~/types/marketing'

const props = defineProps<{ creation: MarketingCreation }>()
const store = useMarketingStore()
const container = ref<HTMLElement | null>(null)
const canvas = ref<HTMLCanvasElement | null>(null)
const loading = ref(true)
const error = ref('')
let observer: IntersectionObserver | undefined
let visible = false
let disposed = false
let generation = 0
async function render() {
  if (!visible || disposed) return
  const request = ++generation
  loading.value = true; error.value = ''
  const urls: string[] = []
  try {
    const creation = JSON.parse(JSON.stringify(props.creation)) as MarketingCreation
    const template = store.templates.find(item => item.id === creation.template)
    if (!template && !['bold', 'minimal'].includes(creation.template)) throw new Error('Modelo indisponível')
    const ids = [...new Set([...(template ? [] : [creation.imageId, creation.logoImageId]), ...(creation.layers || []).filter(layer => layer.visible).map(layer => layer.imageId)].filter((id): id is string => !!id))]
    const resources: Record<string, string> = {}
    const results = await Promise.allSettled(ids.map(async id => {
      const url = await store.loadImage(id)
      if (url) { urls.push(url); resources[id] = url }
    }))
    if (results.some(result => result.status === 'rejected') || ids.some(id => !resources[id])) throw new Error('Imagem indisponível')
    if (disposed || request !== generation) return
    const full = document.createElement('canvas')
    await renderMarketingCanvas(full, creation, resources[creation.imageId || ''], resources[creation.logoImageId || ''], template, resources)
    if (disposed || request !== generation) return
    await nextTick()
    const target = canvas.value
    if (!target) return
    target.width = Math.min(240, full.width)
    target.height = Math.round(full.height * target.width / full.width)
    const context = target.getContext('2d')
    if (!context) throw new Error('Prévia indisponível')
    context.drawImage(full, 0, 0, target.width, target.height)
  } catch {
    if (!disposed && request === generation) error.value = 'Não foi possível carregar a prévia.'
  } finally {
    urls.forEach(url => URL.revokeObjectURL(url))
    if (!disposed && request === generation) loading.value = false
  }
}
watch(() => [props.creation, store.templates], () => void render(), { deep: true })
onMounted(() => {
  if (!('IntersectionObserver' in window)) { visible = true; void render(); return }
  observer = new IntersectionObserver(entries => {
    if (!entries.some(entry => entry.isIntersecting)) return
    visible = true; observer?.disconnect(); void render()
  }, { rootMargin: '160px' })
  if (container.value) observer.observe(container.value)
})
onBeforeUnmount(() => { disposed = true; ++generation; observer?.disconnect() })
</script>

<template>
  <div ref="container" class="creation-preview" :aria-busy="loading">
    <canvas ref="canvas" v-show="!loading && !error" role="img" :aria-label="`Prévia da criação ${creation.title}`" />
    <p v-if="loading" role="status">Carregando prévia...</p>
    <div v-else-if="error" role="alert"><p>{{ error }}</p><button type="button" class="button" @click="render">Tentar novamente</button></div>
  </div>
</template>

<style scoped>
.creation-preview { min-height: 180px; display: grid; place-items: center; background: #edf0f4; border-radius: 8px; padding: .5rem; }
.creation-preview canvas { max-width: 100%; max-height: 320px; width: auto; height: auto; }
.creation-preview p { text-align: center; }
</style>
