<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { onBeforeRouteLeave } from 'vue-router'
import type { MarketingBrand, MarketingCreation, MarketingCreationInput, MarketingTemplate } from '~/types/marketing'
import { cloneMarketingInput, createFromMarketingTemplate, reopenMarketingCreation } from '~/utils/marketingTemplates'
import MarketingTemplates from '../../organisms/MarketingTemplates.vue'
import PageHeader from '../../molecules/PageHeader/Index.vue'
import MarketingEditor from '../../organisms/MarketingEditor.vue'
import MarketingCreations from '../../organisms/MarketingCreations.vue'
import MarketingBrandForm from '../../organisms/MarketingBrandForm.vue'

const marketing = useMarketingStore()
const ready = ref(false)
const templatesLoading = ref(false)
const templatesError = ref('')
const tab = ref<'models' | 'create' | 'creations' | 'brand'>('models')
const tabs = [{ id: 'models', label: 'Modelos' }, { id: 'create', label: 'Criar' }, { id: 'creations', label: 'Minhas criações' }, { id: 'brand', label: 'Identidade visual' }] as const
const flowKey = ref(0)
const startStep = ref(1)
const editorStep = ref(1)
const saving = ref(false)
const savedMessage = ref('')
const brandDirty = ref(false)
const brandBusy = ref(false)
const editorBusy = ref(false)
const brandForm = ref<{ reset: () => void } | null>(null)
const creations = ref<{ refresh: () => Promise<unknown> } | null>(null)
const creationId = ref<string>()
function emptyCreation(brand: MarketingBrand | null): MarketingCreationInput {
  return { kind: 'service', objective: 'promotion', template: 'bold', format: 'feed', title: '', description: '', priceCents: null, cta: 'Entre em contato', primaryColor: brand?.primaryColor || '#2563eb', secondaryColor: brand?.secondaryColor || '#ffffff', caption: '', imageId: null, productId: null, brandName: brand?.name || '', brandContact: brand?.contact || '', logoImageId: brand?.logoImageId || null }
}
function snapshot(input: MarketingCreationInput) {
  return JSON.stringify(Object.entries(cloneMarketingInput(input)).sort(([a], [b]) => a.localeCompare(b)))
}
const input = ref<MarketingCreationInput>(emptyCreation(null))
const baseline = ref(snapshot(input.value))
const editorDirty = computed(() => snapshot(input.value) !== baseline.value)
const dirty = computed(() => editorDirty.value || brandDirty.value)
function confirmDiscard() {
  return !dirty.value || window.confirm('Há alterações não salvas. Deseja descartá-las?')
}
function discard() {
  input.value = JSON.parse(JSON.stringify(Object.fromEntries(JSON.parse(baseline.value)))) as MarketingCreationInput
  brandForm.value?.reset()
  brandDirty.value = false
}
async function changeTab(next: typeof tab.value) {
  if (next === tab.value || saving.value || brandBusy.value || editorBusy.value) return
  if (!confirmDiscard()) return
  discard()
  tab.value = next
  if (next === 'creations') await creations.value?.refresh()
}
async function tabKeydown(event: KeyboardEvent) {
  const index = tabs.findIndex(item => item.id === tab.value)
  let next = index
  if (event.key === 'ArrowRight') next = (index + 1) % tabs.length
  else if (event.key === 'ArrowLeft') next = (index + tabs.length - 1) % tabs.length
  else if (event.key === 'Home') next = 0
  else if (event.key === 'End') next = tabs.length - 1
  else return
  event.preventDefault()
  await changeTab(tabs[next]!.id)
  await nextTick()
  document.getElementById(`marketing-tab-${tab.value}`)?.focus()
}
async function open(creation: MarketingCreation, duplicate = false) {
  if (saving.value || brandBusy.value || editorBusy.value || !confirmDiscard()) return
  if (creation.template !== 'bold' && creation.template !== 'minimal') {
    if (!marketing.templates.length) await loadTemplates()
    if (!marketing.templates.some(item => item.id === creation.template)) { marketing.error = 'O modelo desta criação não está disponível. Recarregue os modelos e tente novamente.'; return }
  }
  discard()
  creationId.value = duplicate ? undefined : creation.id
  input.value = reopenMarketingCreation(creation, marketing.templates.find(item => item.id === creation.template))
  if (duplicate) input.value.title = `${input.value.title.slice(0, 132)} (cópia)`
  baseline.value = duplicate ? snapshot(emptyCreation(marketing.brand)) : snapshot(input.value)
  savedMessage.value = ''
  startStep.value = 2; editorStep.value = 2; flowKey.value++
  tab.value = 'create'
}
function selectTemplate(template: MarketingTemplate) {
  if (saving.value || brandBusy.value || editorBusy.value || !confirmDiscard()) return
  brandForm.value?.reset()
  brandDirty.value = false
  creationId.value = undefined
  input.value = createFromMarketingTemplate(template, marketing.brand)
  baseline.value = snapshot(emptyCreation(marketing.brand))
  savedMessage.value = ''
  startStep.value = 2; editorStep.value = 2; flowKey.value++
  tab.value = 'create'
}
async function loadTemplates() {
  if (templatesLoading.value) return
  templatesLoading.value = true
  templatesError.value = ''
  try {
    if (!await marketing.loadTemplates(true)) templatesError.value = 'Não foi possível carregar os modelos. Você ainda pode criar com as composições básicas.'
  } catch { templatesError.value = 'Não foi possível carregar os modelos. Você ainda pode criar com as composições básicas.' }
  finally {
    if (marketing.error === 'Não foi possível carregar os modelos.') marketing.error = ''
    templatesLoading.value = false
  }
}
function newCreation() {
  if (saving.value || brandBusy.value || editorBusy.value || !confirmDiscard()) return
  brandForm.value?.reset()
  brandDirty.value = false
  creationId.value = undefined
  input.value = emptyCreation(marketing.brand)
  baseline.value = snapshot(input.value)
  savedMessage.value = ''
  startStep.value = 1; editorStep.value = 1; flowKey.value++
  tab.value = 'create'
}
function onDeleted(id: string) {
  if (creationId.value !== id) return
  creationId.value = undefined
  input.value = emptyCreation(marketing.brand)
  baseline.value = snapshot(input.value)
  savedMessage.value = ''
}
async function save() {
  if (saving.value || editorBusy.value || brandBusy.value) return
  saving.value = true
  savedMessage.value = ''
  const submitted = cloneMarketingInput(input.value)
  try {
    const saved = await marketing.saveCreation(submitted, creationId.value)
    if (saved) { creationId.value = saved.id; baseline.value = snapshot(submitted); savedMessage.value = 'Criação salva.' }
  } catch { marketing.error = 'Não foi possível salvar a criação. Seu conteúdo continua disponível.' }
  finally { saving.value = false }
}
function beforeUnload(event: BeforeUnloadEvent) {
  if (dirty.value || saving.value || brandBusy.value || editorBusy.value) { event.preventDefault(); event.returnValue = '' }
}
onBeforeRouteLeave(() => !saving.value && !brandBusy.value && !editorBusy.value && confirmDiscard())
async function loadBrand() {
  ready.value = false
  try { await marketing.loadBrand(true); input.value = emptyCreation(marketing.brand); baseline.value = snapshot(input.value) }
  catch { marketing.error = 'Não foi possível carregar a identidade visual.' }
  finally { ready.value = true }
}
onMounted(() => {
  window.addEventListener('beforeunload', beforeUnload)
  void loadBrand()
  void loadTemplates()
})
onBeforeUnmount(() => window.removeEventListener('beforeunload', beforeUnload))
</script>

<template>
  <section class="page marketing-template">
    <PageHeader title="Marketing" subtitle="Transforme produtos, serviços e ofertas em posts prontos para divulgar.">
      <template #actions><button class="button button--secondary" type="button" :disabled="saving || brandBusy || editorBusy || !ready" @click="newCreation">Nova criação</button></template>
    </PageHeader>
    <div class="marketing-template__tabs" role="tablist" aria-label="Marketing">
      <button v-for="item in tabs" :id="`marketing-tab-${item.id}`" :key="item.id" type="button" role="tab" :tabindex="tab === item.id ? 0 : -1" :aria-selected="tab === item.id" :aria-controls="`marketing-panel-${item.id}`" :disabled="saving || brandBusy || editorBusy || !ready" @click="changeTab(item.id)" @keydown="tabKeydown">{{ item.label }}</button>
    </div>
    <p v-if="marketing.error" class="marketing-template__error" role="alert">{{ marketing.error }}</p>
    <p v-if="!ready" role="status">Carregando identidade visual...</p>
    <div v-else-if="!marketing.brand" class="panel"><button class="button" type="button" @click="loadBrand">Tentar novamente</button></div>
    <template v-else>
      <div v-show="tab === 'models'" id="marketing-panel-models" role="tabpanel" aria-labelledby="marketing-tab-models"><MarketingTemplates :loading="templatesLoading" :error="templatesError" @select="selectTemplate" @retry="loadTemplates" /></div>
      <div v-show="tab === 'create'" id="marketing-panel-create" role="tabpanel" aria-labelledby="marketing-tab-create">
        <MarketingEditor v-model="input" :brand="marketing.brand" :start-step="startStep" :flow-key="flowKey" @step="editorStep = $event" @select-template="selectTemplate" @busy="editorBusy = $event" />
        <div v-show="editorStep === 3" class="marketing-template__save">
          <span role="status">{{ saving ? 'Salvando criação...' : editorDirty ? 'Alterações não salvas' : savedMessage }}</span>
          <button class="button" type="button" :disabled="saving || editorBusy || !input.title.trim()" @click="save">{{ saving ? 'Salvando...' : 'Salvar criação' }}</button>
        </div>
      </div>
      <div v-show="tab === 'creations'" id="marketing-panel-creations" role="tabpanel" aria-labelledby="marketing-tab-creations"><MarketingCreations ref="creations" @open="open($event)" @duplicate="open($event, true)" @deleted="onDeleted" /></div>
      <div v-show="tab === 'brand'" id="marketing-panel-brand" role="tabpanel" aria-labelledby="marketing-tab-brand"><MarketingBrandForm ref="brandForm" @dirty="brandDirty = $event" @busy="brandBusy = $event" /></div>
    </template>
  </section>
</template>
