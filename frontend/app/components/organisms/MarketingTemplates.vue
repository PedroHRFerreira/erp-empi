<script setup lang="ts">
import { computed, ref } from 'vue'
import { useMarketingStore } from '~/stores/useMarketingStore'
import type { MarketingTemplate } from '~/types/marketing'

const props = withDefaults(defineProps<{ loading?: boolean; error?: string }>(), { loading: false, error: '' })
const emit = defineEmits<{ select: [template: MarketingTemplate]; retry: [] }>()
const store = useMarketingStore()
const search = ref('')
const category = ref('')
const failedImages = ref(new Set<string>())
const busy = computed(() => props.loading)
const error = computed(() => props.error)
const categories = computed(() => [...new Set(store.templates.map(item => item.category))].sort((a, b) => a.localeCompare(b, 'pt-BR')))
const publicationCount = computed(() => new Set(store.templates.map(item => item.sourceUrl).filter(Boolean)).size)
const normalize = (value: string) => value.normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLocaleLowerCase('pt-BR')
const filtered = computed(() => {
  const query = normalize(search.value.trim())
  return store.templates.filter(item => (!category.value || item.category === category.value) && (!query || normalize(`${item.name} ${item.category}`).includes(query)))
})
function sourceLabel(item: MarketingTemplate) {
  if (item.sourceKind === 'reel') return 'Capa de Reel'
  if (item.sourceKind === 'carousel') return `Carrossel · página ${item.slide}`
  return 'Post'
}
function originalFilename(item: MarketingTemplate) {
  const extension = item.previewUrl?.match(/\.(jpe?g|png|webp)(?:[?#]|$)/i)?.[1]?.toLowerCase() || 'jpg'
  return `${item.id}-original.${extension}`
}
function failed(id: string) { failedImages.value = new Set([...failedImages.value, id]) }
function retryImage(id: string) { const next = new Set(failedImages.value); next.delete(id); failedImages.value = next }
function resetFilters() { search.value = ''; category.value = '' }
function retry() { if (!busy.value) emit('retry') }

</script>

<template>
  <section class="marketing-models" aria-labelledby="marketing-models-heading" :aria-busy="busy">
    <header class="marketing-models__intro">
      <h2 id="marketing-models-heading">Modelos da EMPI</h2>
      <p>Escolha uma arte original. Ela abre com a mesma composição da prévia. No editor, você pode substituir textos, fotos, preços e contatos e exportar sua versão em PNG.</p>
      <p v-if="store.templates.length">{{ store.templates.length }} modelos de {{ publicationCount }} publicações do perfil público da EMPI.</p>
    </header>
    <div v-if="busy" role="status">Carregando modelos...</div>
    <div v-else-if="error" role="alert" class="form-error">
      <p>{{ error }}</p>
      <button type="button" class="button" @click="retry">Tentar novamente</button>
    </div>
    <template v-else-if="store.templates.length">
      <div class="marketing-models__toolbar">
        <label for="marketing-model-search">Buscar modelos
          <input id="marketing-model-search" v-model="search" type="search" placeholder="Nome ou categoria" />
        </label>
        <label for="marketing-model-category">Categoria
          <select id="marketing-model-category" v-model="category">
            <option value="">Todas as categorias</option>
            <option v-for="item in categories" :key="item" :value="item">{{ item }}</option>
          </select>
        </label>
        <button v-if="search || category" type="button" class="button" @click="resetFilters">Limpar filtros</button>
      </div>
      <p role="status">{{ filtered.length }} {{ filtered.length === 1 ? 'modelo encontrado' : 'modelos encontrados' }}</p>
      <div v-if="filtered.length" class="marketing-models__grid">
        <article v-for="item in filtered" :key="item.id" class="panel marketing-models__card">
          <div class="marketing-models__preview">
            <img v-if="item.previewUrl && !failedImages.has(item.id)" :src="item.previewUrl" :alt="`Arte original: ${item.name}`" loading="lazy" decoding="async" @error="failed(item.id)" />
            <div v-else class="marketing-models__preview-error"><p>Prévia indisponível.</p><button type="button" class="button" @click="retryImage(item.id)">Carregar prévia</button></div>
          </div>
          <div class="marketing-models__details">
            <h3>{{ item.name }}</h3>
            <p>{{ item.category }} · {{ sourceLabel(item) }}</p>
            <small>Prévia da arte original</small>
            <div class="marketing-models__actions">
              <button type="button" class="button button--primary" :aria-label="`Usar modelo ${item.name}`" @click="emit('select', item)">Usar modelo</button>
              <a v-if="item.sourceUrl" :href="item.sourceUrl" target="_blank" rel="noopener noreferrer" :aria-label="`Ver original de ${item.name} no Instagram (abre em nova aba)`">Ver original</a>
              <a v-if="item.previewUrl && !failedImages.has(item.id)" :href="item.previewUrl" :download="originalFilename(item)" :aria-label="`Baixar arte original de ${item.name}`">Baixar arte original</a>
            </div>
          </div>
        </article>
      </div>
      <div v-else class="panel marketing-models__empty">
        <p>Nenhum modelo corresponde à busca.</p>
        <button type="button" class="button" @click="resetFilters">Limpar filtros</button>
      </div>
    </template>
    <div v-else class="panel marketing-models__empty">
      <p>Os modelos ainda não estão disponíveis.</p>
      <button type="button" class="button" @click="retry">Carregar modelos</button>
    </div>
  </section>
</template>
