<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { MarketingLayer } from '~/types/marketing'
import { useMarketingStore } from '~/stores/useMarketingStore'
const props = defineProps<{ modelValue: MarketingLayer[]; defaults: MarketingLayer[]; disabled?: boolean; title?: string; price?: string; contact?: string }>()
const emit = defineEmits<{ 'update:modelValue': [MarketingLayer[]]; busy: [boolean]; selection: [string] }>()
const store = useMarketingStore()
const selectedId = ref(props.modelValue[0]?.id || '')
const selected = computed(() => props.modelValue.find(item => item.id === selectedId.value))
watch(selectedId, id => emit('selection', id), { immediate: true })
function select(id: string) { if (props.modelValue.some(layer => layer.id === id)) selectedId.value = id }
const uploading = ref(false)
const error = ref('')
const fonts = ['Inter', 'Roboto Condensed', 'Oswald', 'Bebas Neue', 'Anton', 'Dancing Script', 'Knewave', 'Kalam'] as const
watch(() => props.modelValue, layers => { if (!layers.some(item => item.id === selectedId.value)) selectedId.value = layers[0]?.id || '' })
function replace(layers: MarketingLayer[]) { emit('update:modelValue', layers.map(layer => ({ ...layer }))) }
function patch(values: Partial<MarketingLayer>) { if (!selected.value || props.disabled || uploading.value) return; replace(props.modelValue.map(layer => layer.id === selectedId.value ? { ...layer, ...values } : layer)) }
function input(field: keyof MarketingLayer, event: Event) { patch({ [field]: (event.target as HTMLInputElement).value }) }
function position(field: 'x' | 'y' | 'width' | 'height', event: Event) {
  const value = Number((event.target as HTMLInputElement).value) / 100
  if (!Number.isFinite(value) || !selected.value) return
  const layer = selected.value
  const next = { ...layer, [field]: Math.max(field === 'width' || field === 'height' ? .01 : 0, Math.min(1, value)) }
  next.width = Math.min(next.width, 1 - next.x); next.height = Math.min(next.height, 1 - next.y)
  if (next.width < .01) { next.x = .99; next.width = .01 }
  if (next.height < .01) { next.y = .99; next.height = .01 }
  patch(next)
}
function add(text = 'Novo texto') {
  if (props.disabled || uploading.value) return
  if (props.modelValue.length >= 60) { error.value = 'Esta arte atingiu o limite de 60 elementos.'; return }
  const id = `custom-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 7)}`
  replace([...props.modelValue, { id, type: 'text', x: .08, y: .45, width: .84, height: .12, text, fontFamily: 'Inter', fontSize: 48, color: '#ffffff', fontWeight: 700, align: 'left', visible: true }]); selectedId.value = id
}
function addPhoto() {
  if (props.disabled || uploading.value) return
  if (props.modelValue.length >= 60 || props.modelValue.filter(layer => layer.type === 'image').length >= 12) { error.value = 'Você pode usar até 60 elementos e 12 fotos nesta arte.'; return }
  const id = `custom-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 7)}`
  replace([...props.modelValue, { id, type: 'image', x: .1, y: .25, width: .8, height: .5, visible: true, imageId: null, fit: 'cover' }]); selectedId.value = id
}
function restore() { const original = props.defaults.find(item => item.id === selectedId.value); if (original) patch({ ...original }) }
function reset() { if (JSON.stringify(props.modelValue) !== JSON.stringify(props.defaults) && !window.confirm('Restaurar todos os elementos? As alterações desta arte serão descartadas.')) return; replace(props.defaults) }
function remove() { if (!selected.value) return; if (props.defaults.some(item => item.id === selectedId.value)) patch({ visible: false }); else replace(props.modelValue.filter(item => item.id !== selectedId.value)) }
async function upload(event: Event) {
  const input = event.target as HTMLInputElement; const file = input.files?.[0]; input.value = ''
  if (!file || !selected.value || uploading.value || props.disabled) return
  error.value = ''
  if (!['image/png', 'image/jpeg'].includes(file.type) || file.size > 5 * 1024 * 1024) { error.value = 'Escolha PNG ou JPEG de até 5 MB.'; return }
  const id = selectedId.value; uploading.value = true; emit('busy', true)
  try { const result = await store.uploadImage(file); if (result) replace(props.modelValue.map(layer => layer.id === id ? { ...layer, imageId: result.id, visible: true } : layer)); else { error.value = store.error || 'Não foi possível enviar a foto. A anterior foi mantida.'; store.error = '' } }
  catch { error.value = 'Não foi possível enviar a foto. Sua criação foi mantida.' }
  finally { uploading.value = false; emit('busy', false) }
}
defineExpose({ add, select })
</script>
<template>
  <fieldset class="marketing-layers" :disabled="disabled || uploading">
    <legend>Elementos da arte</legend>
    <label class="field"><span>Selecionar elemento</span><select v-model="selectedId"><option v-for="(layer, index) in modelValue" :key="layer.id" :value="layer.id">{{ index + 1 }} · {{ layer.type === 'image' ? (layer.id === 'logo-brand' ? 'Logo' : layer.id.startsWith('decor-') ? 'Elemento gráfico' : 'Foto') : (layer.text || 'Texto vazio').slice(0, 70) }}{{ !layer.visible ? ' (oculto)' : '' }}</option></select></label>
    <template v-if="selected">
      <label class="marketing-layers__check"><input type="checkbox" :checked="selected.visible" @change="patch({ visible: ($event.target as HTMLInputElement).checked })" /> Mostrar elemento</label>
      <template v-if="selected.type === 'text'">
        <label class="field"><span>Texto</span><textarea :value="selected.text" rows="3" maxlength="500" @input="input('text', $event)" /></label>
        <details><summary>Ajustes avançados do texto</summary><div class="marketing-editor__row"><label class="field"><span>Fonte</span><select :value="selected.fontFamily || 'Inter'" @change="input('fontFamily', $event)"><option v-for="font in fonts" :key="font">{{ font }}</option></select></label><label class="field"><span>Cor do texto</span><input type="color" :value="selected.color || '#ffffff'" @input="input('color', $event)" /></label></div>
        <div class="marketing-editor__row"><label class="field"><span>Tamanho do texto (px)</span><input type="number" min="8" max="400" step="1" :value="selected.fontSize || 48" @input="patch({ fontSize: Math.max(8, Math.min(400, Number(($event.target as HTMLInputElement).value))) })" /></label><label class="field"><span>Alinhamento</span><select :value="selected.align || 'left'" @change="input('align', $event)"><option value="left">Esquerda</option><option value="center">Centro</option><option value="right">Direita</option></select></label></div>
        <div class="marketing-layers__actions"><label><input type="checkbox" :checked="(selected.fontWeight || 400) >= 700" @change="patch({ fontWeight: ($event.target as HTMLInputElement).checked ? 700 : 400 })" /> Negrito</label><label><input type="checkbox" :checked="selected.italic" @change="patch({ italic: ($event.target as HTMLInputElement).checked })" /> Itálico</label></div>
        </details><div class="marketing-layers__actions"><button v-if="title" type="button" class="button button--secondary" @click="patch({ text: title })">Usar nome do produto</button><button v-if="price" type="button" class="button button--secondary" @click="patch({ text: price })">Usar preço</button><button v-if="contact" type="button" class="button button--secondary" @click="patch({ text: contact })">Usar contato salvo</button></div>
      </template>
      <template v-else><label class="field"><span>Substituir foto</span><input type="file" accept="image/png,image/jpeg" @change="upload" /><small>PNG ou JPEG de até 5 MB.</small></label><label class="field"><span>Enquadramento</span><select :value="selected.fit || 'cover'" @change="input('fit', $event)"><option value="cover">Preencher área</option><option value="contain">Mostrar foto inteira</option></select></label></template>
      <details><summary>Ajustes avançados: posição e tamanho</summary><div class="marketing-editor__row"><label v-for="field in (['x', 'y', 'width', 'height'] as const)" :key="field" class="field"><span>{{ { x: 'Esquerda', y: 'Topo', width: 'Largura', height: 'Altura' }[field] }} (%)</span><input type="number" min="0" max="100" step="0.1" :value="Math.round(selected[field] * 1000) / 10" @change="position(field, $event)" /></label></div><label class="field"><span>Rotação (graus)</span><input type="number" min="-30" max="30" step="1" :value="selected.rotation || 0" @change="patch({ rotation: Math.max(-30, Math.min(30, Number(($event.target as HTMLInputElement).value))) })" /></label></details>
      <div class="marketing-layers__actions"><button v-if="defaults.some(item => item.id === selectedId)" class="button button--secondary" type="button" @click="restore">Restaurar elemento</button><button class="button button--secondary" type="button" @click="remove">{{ defaults.some(item => item.id === selectedId) ? 'Ocultar elemento' : 'Excluir elemento' }}</button></div>
    </template>
    <div class="marketing-layers__actions"><button type="button" class="button button--secondary" @click="add()">Adicionar texto</button><button type="button" class="button button--secondary" @click="addPhoto">Adicionar foto</button><button v-if="price" type="button" class="button button--secondary" @click="add(price)">Adicionar preço</button><button type="button" class="button button--secondary" @click="reset">Restaurar arte original</button></div>
  </fieldset>
  <p v-if="uploading" role="status">Enviando foto…</p><p v-if="error" class="form-error" role="alert">{{ error }}</p>
</template>
<style scoped>
.marketing-layers { display: grid; gap: 1rem; min-width: 0; }
.marketing-layers__actions { display: flex; flex-wrap: wrap; gap: .75rem; align-items: center; }
.marketing-layers__check { display: flex; gap: .5rem; align-items: center; }
summary { cursor: pointer; padding: .5rem 0; }
</style>
