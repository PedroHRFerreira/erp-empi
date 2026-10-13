<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import MarketingCreationPreview from './MarketingCreationPreview.vue'
import { useMarketingStore } from '~/stores/useMarketingStore'
import type { MarketingCreation } from '~/types/marketing'

const emit = defineEmits<{ open: [creation: MarketingCreation]; duplicate: [creation: MarketingCreation]; deleted: [id: string] }>()
const store = useMarketingStore()
const page = ref(0)
const loading = ref(false)
const error = ref('')
const deleting = ref(false)
const deleteTarget = ref<MarketingCreation | null>(null)
const pages = computed(() => Math.max(1, Math.ceil(store.total / 24)))
const labels = { product: 'Produto', service: 'Serviço', offer: 'Oferta combinada' }
async function refresh() {
  if (loading.value) return
  loading.value = true; error.value = ''
  try {
    if (!await store.loadCreations(page.value * 24, true)) error.value = store.error
  } catch { error.value = 'Não foi possível carregar as criações. Tente novamente.' }
  finally { loading.value = false }
}
async function changePage(next: number) { if (loading.value || deleting.value) return; page.value = next; await refresh() }
async function remove() {
  if (!deleteTarget.value || deleting.value) return
  deleting.value = true; error.value = ''
  try {
    const id = deleteTarget.value.id
    if (!await store.deleteCreation(id)) { error.value = store.error; return }
    emit('deleted', id)
    deleteTarget.value = null
    page.value = Math.min(page.value, Math.max(0, Math.ceil(store.total / 24) - 1))
    await refresh()
  } catch { error.value = 'Não foi possível excluir a criação. Tente novamente.' }
  finally { deleting.value = false }
}
function date(value: string) { return new Date(value).toLocaleDateString('pt-BR') }
defineExpose({ refresh })
onMounted(refresh)
</script>

<template>
  <section class="marketing-creations" aria-label="Minhas criações" :aria-busy="loading">
    <p v-if="loading" role="status">Carregando criações...</p>
    <div v-else-if="error" role="alert" class="form-error">
      <p>{{ error }}</p>
      <button type="button" class="button" @click="refresh">Tentar novamente</button>
    </div>
    <div v-else-if="!store.creations.length" class="panel marketing-creations__empty">
      <h2>Suas divulgações ficam aqui</h2>
      <p>Crie e salve sua primeira arte na aba Criar para poder reutilizá-la depois.</p>
    </div>
    <template v-else>
      <p>{{ store.total }} {{ store.total === 1 ? 'criação salva' : 'criações salvas' }}</p>
      <div class="marketing-creations__grid">
        <article v-for="creation in store.creations" :key="creation.id" class="panel marketing-creations__card">
          <MarketingCreationPreview :creation="creation" />
          <h3>{{ creation.title }}</h3>
          <p>{{ labels[creation.kind] }} · {{ creation.format === 'feed' ? 'Feed' : creation.format === 'square' ? 'Quadrado' : creation.format === 'original' ? 'Original' : 'Story' }}</p>
          <small>Atualizada em {{ date(creation.updatedAt) }}</small>
          <div class="marketing-creations__actions">
            <button type="button" class="button button--primary" @click="emit('open', creation)">Abrir</button>
            <button type="button" class="button" @click="emit('duplicate', creation)">Duplicar</button>
            <button type="button" class="button" :aria-label="`Excluir ${creation.title}`" :disabled="deleting" @click="deleteTarget = creation">Excluir</button>
          </div>
        </article>
      </div>
      <nav v-if="pages > 1" aria-label="Páginas de criações" class="marketing-creations__pagination">
        <button type="button" class="button" :disabled="page === 0 || loading" @click="changePage(page - 1)">Anterior</button>
        <span>Página {{ page + 1 }} de {{ pages }}</span>
        <button type="button" class="button" :disabled="page + 1 >= pages || loading" @click="changePage(page + 1)">Próxima</button>
      </nav>
    </template>
    <div v-if="deleteTarget" class="panel marketing-creations__confirmation" role="group" aria-label="Confirmar exclusão">
      <p>Excluir “{{ deleteTarget.title }}”? Esta ação não pode ser desfeita.</p>
      <p v-if="error" role="alert" class="form-error">{{ error }}</p>
      <button type="button" class="button" :disabled="deleting" @click="deleteTarget = null">Cancelar</button>
      <button type="button" class="button button--primary" :disabled="deleting" @click="remove">{{ deleting ? 'Excluindo...' : 'Confirmar exclusão' }}</button>
    </div>
  </section>
</template>

<style scoped>
.marketing-creations__grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 260px), 1fr)); gap: 1rem; }
.marketing-creations__card { padding: 1rem; min-width: 0; }
.marketing-creations__card h3 { overflow-wrap: anywhere; }
.marketing-creations__actions, .marketing-creations__pagination { display: flex; flex-wrap: wrap; gap: .5rem; align-items: center; margin-top: 1rem; }
.marketing-creations__confirmation, .marketing-creations__empty { padding: 1.5rem; margin-top: 1rem; }
</style>
