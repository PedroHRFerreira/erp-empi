<script setup lang="ts">
import { onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useMarketingStore } from '~/stores/useMarketingStore'
import type { MarketingBrand } from '~/types/marketing'

const emit = defineEmits<{ dirty: [value: boolean]; saved: [brand: MarketingBrand]; busy: [value: boolean] }>()
const store = useMarketingStore()
const form = reactive<MarketingBrand>({ name: '', logoImageId: null, primaryColor: '#153a62', secondaryColor: '#ffffff', contact: '' })
const baseline = ref(JSON.stringify(form))
const loading = ref(true)
const saving = ref(false)
const uploading = ref(false)
const error = ref('')
const saved = ref(false)
const logoUrl = ref('')
let imageRequest = 0
let disposed = false
function replaceUrl(url: string) { if (logoUrl.value) URL.revokeObjectURL(logoUrl.value); logoUrl.value = url }
async function showLogo(id: string | null) {
  const request = ++imageRequest
  replaceUrl('')
  if (!id) return
  let url: string | null = null
  try { url = await store.loadImage(id) } catch { if (request === imageRequest && !disposed) error.value = 'Não foi possível carregar o logo. Tente novamente.' }
  if (request !== imageRequest || disposed) { if (url) URL.revokeObjectURL(url); return }
  if (url) replaceUrl(url)
  else error.value = store.error
}
function reset() {
  if (saving.value || uploading.value) return false
  if (store.brand) Object.assign(form, store.brand)
  baseline.value = JSON.stringify(form); error.value = ''; saved.value = false
  emit('dirty', false)
  void showLogo(form.logoImageId)
  return true
}
async function load() {
  if (!loading.value && (saving.value || uploading.value)) return
  loading.value = true; error.value = ''
  try { if (await store.loadBrand()) reset(); else error.value = store.error }
  catch { error.value = 'Não foi possível carregar a identidade visual.' }
  finally { loading.value = false }
}
watch(form, () => { saved.value = false; emit('dirty', JSON.stringify(form) !== baseline.value) }, { flush: 'sync' })
watch([saving, uploading], () => emit('busy', saving.value || uploading.value), { flush: 'sync' })
async function upload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]; input.value = ''
  if (!file || uploading.value || saving.value) return
  error.value = ''; saved.value = false
  if (!['image/png', 'image/jpeg'].includes(file.type) || file.size > 5 * 1024 * 1024) {
    error.value = 'Escolha uma imagem PNG ou JPEG de até 5 MB.'; return
  }
  uploading.value = true
  try {
    const image = await store.uploadImage(file)
    if (!image) { error.value = store.error; return }
    form.logoImageId = image.id
    ++imageRequest; replaceUrl(URL.createObjectURL(file))
  } catch { error.value = 'Não foi possível enviar a imagem. Tente novamente.' }
  finally { uploading.value = false }
}
function removeLogo() { if (saving.value || uploading.value) return; form.logoImageId = null; ++imageRequest; replaceUrl('') }
async function save() {
  if (saving.value || uploading.value) return
  saving.value = true; error.value = ''; saved.value = false
  try {
    if (!await store.saveBrand({ ...form })) { error.value = store.error; return }
    Object.assign(form, store.brand)
    baseline.value = JSON.stringify(form); emit('dirty', false); saved.value = true
    emit('saved', { ...form })
  } catch { error.value = 'Não foi possível salvar. Suas alterações continuam disponíveis.' }
  finally { saving.value = false }
}
defineExpose({ reset })
onMounted(load)
onBeforeUnmount(() => { disposed = true; ++imageRequest; replaceUrl('') })
</script>

<template>
  <section class="marketing-brand" aria-label="Identidade visual">
    <p v-if="loading" role="status">Carregando identidade visual...</p>
    <div v-else-if="!store.brand" class="form-error" role="alert">
      <p>{{ error }}</p><button class="button" type="button" @click="load">Tentar novamente</button>
    </div>
    <form v-else class="panel marketing-brand__form" :aria-busy="saving || uploading" @submit.prevent="save">
      <p>Defina a apresentação do seu negócio para preencher as próximas criações.</p>
      <label class="field"><span>Nome do negócio</span><input v-model="form.name" maxlength="140" :disabled="saving" autocomplete="organization" /></label>
      <label class="field"><span>Contato</span><input v-model="form.contact" maxlength="200" :disabled="saving" placeholder="Telefone, endereço ou perfil social" /></label>
      <div class="marketing-brand__colors">
        <label class="field"><span>Cor principal</span><input v-model="form.primaryColor" type="color" :disabled="saving" /></label>
        <label class="field"><span>Cor secundária</span><input v-model="form.secondaryColor" type="color" :disabled="saving" /></label>
      </div>
      <label class="field"><span>Logo · PNG ou JPEG de até 5 MB</span><input type="file" accept="image/png,image/jpeg" :disabled="saving || uploading" @change="upload" /></label>
      <p v-if="uploading" role="status">Enviando logo...</p>
      <div v-if="form.logoImageId" class="marketing-brand__logo">
        <img v-if="logoUrl" :src="logoUrl" alt="Logo do negócio" />
        <button class="button" type="button" :disabled="saving || uploading" @click="removeLogo">Remover logo</button>
      </div>
      <p class="field__hint">Os modelos guiados usam Inter. As artes do Instagram preservam a tipografia de cada modelo.</p>
      <p v-if="error" role="alert" class="form-error">{{ error }}</p>
      <p v-if="saved" role="status" class="form-success">Identidade visual salva.</p>
      <button class="button button--primary" type="submit" :disabled="saving || uploading">{{ saving ? 'Salvando...' : 'Salvar identidade visual' }}</button>
    </form>
  </section>
</template>

<style scoped>
.marketing-brand__form { display: grid; gap: 1rem; padding: 1.5rem; }
.marketing-brand__colors { display: flex; flex-wrap: wrap; gap: 1.5rem; }
.marketing-brand__logo { display: flex; flex-wrap: wrap; align-items: center; gap: 1rem; }
.marketing-brand__logo img { width: 100px; height: 100px; object-fit: contain; border: 1px solid #d7dde5; border-radius: 8px; background: #fff; padding: .25rem; }
.marketing-brand__form > .button { justify-self: start; }
</style>
