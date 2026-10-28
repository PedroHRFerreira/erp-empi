<script setup lang="ts">
import { ArrowLeft, Check, Search, Trash2 } from '@lucide/vue'
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import type { IStockItem, PayableMethod } from '../../../../server/contracts/types'
import type { StockForm } from '../../../stores/useStockStore'
import { currencyMaskToCents, formatCentsAsCurrency, maskCurrency } from '../../../utils/masks'
import { formatCurrency } from '../../../utils/format'
import { divideCents } from '../../../utils/purchaseInstallments'
import PageHeader from '../../molecules/PageHeader/Index.vue'

type InstallmentRow = { amountInput: string; dueDate: string; plannedMethod: PayableMethod }
const props = defineProps<{ productId?: string | number }>()
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const stock = useStockStore()
const isEdit = computed(() => props.productId !== undefined && props.productId !== '')
const supplierMode = computed(() => isEdit.value && route.query.supplierMode === '1')
const costInput = ref('')
const entryCostInput = ref('')
const supplierName = ref(typeof route.query.supplierName === 'string' ? route.query.supplierName : '')
const quantity = ref(1)
const addQuantity = ref(!isEdit.value || route.query.replenish === '1')
const installmentCount = ref(1)
const firstDueDate = ref(today())
const installments = ref<InstallmentRow[]>([])
const selectedItem = ref<IStockItem | null>(null)
const loadingItem = ref(isEdit.value)
const loadError = ref('')
const formError = ref('')
const saving = ref(false)
const suggestions = ref<IStockItem[]>([])
const searching = ref(false)
const suggestionsOpen = ref(false)
const dismissedName = ref('')
const activeSuggestion = ref(-1)
let searchVersion = 0
let searchTimer: ReturnType<typeof setTimeout> | undefined
let itemVersion = 0
const initialMarkup = Number(auth.user?.markupPercent ?? 10)
const form = reactive<StockForm>({ name: '', description: '', costCents: 0, markupPercent: Number.isFinite(initialMarkup) ? initialMarkup : 10, quantity: 0 })
const existingSelection = computed(() => !isEdit.value && selectedItem.value !== null)
const metadataLocked = computed(() => supplierMode.value || existingSelection.value)
const hasEntry = computed(() => !isEdit.value || addQuantity.value || supplierMode.value)
const unitCostCents = computed(() => currencyMaskToCents(isEdit.value || existingSelection.value ? entryCostInput.value : costInput.value))
const totalCents = computed(() => unitCostCents.value * Math.max(0, Number(quantity.value || 0)))
const installmentTotal = computed(() => installments.value.reduce((sum, row) => sum + currencyMaskToCents(row.amountInput), 0))
const returnTo = computed(() => typeof route.query.returnTo === 'string' && /^\/stock\/suppliers(?:\/detail)?(?:\?|$)/.test(route.query.returnTo) ? route.query.returnTo : '/stock/products')
const title = computed(() => supplierMode.value ? 'Adicionar quantidade' : isEdit.value ? 'Editar produto' : 'Adicionar produto ao estoque')
const submitLabel = computed(() => supplierMode.value || existingSelection.value ? 'Confirmar entrada no estoque' : isEdit.value ? (hasEntry.value ? 'Salvar e confirmar entrada' : 'Salvar produto') : 'Cadastrar produto e confirmar entrada')

stock.error = ''
stock.fieldErrors = {}
function today() { const date = new Date(); return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}` }
function addMonths(value: string, months: number) {
  const [year, month, day] = value.split('-').map(Number)
  if (!year || !month || !day) return ''
  const date = new Date(year, month - 1 + months, 1)
  date.setDate(Math.min(day, new Date(date.getFullYear(), date.getMonth() + 1, 0).getDate()))
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}
function generateInstallments() {
  const count = Number(installmentCount.value)
  if (!Number.isInteger(count) || count < 1 || count > 48 || !Number.isSafeInteger(totalCents.value) || totalCents.value < 0) { installments.value = []; return }
  installments.value = divideCents(totalCents.value, count).map((amount, index) => ({
    amountInput: maskCurrency(String(amount)),
    dueDate: installments.value[index]?.dueDate || addMonths(firstDueDate.value, index),
    plannedMethod: installments.value[index]?.plannedMethod || 'boleto'
  }))
}
function removeInstallment(index: number) {
  if (installments.value.length <= 1) return
  installments.value.splice(index, 1)
  installmentCount.value = installments.value.length
}
function fillItem(item: IStockItem) {
  selectedItem.value = item
  Object.assign(form, { name: item.name, description: item.description || '', costCents: item.costCents, markupPercent: item.markupPercent, quantity: item.quantity })
  costInput.value = formatCentsAsCurrency(item.costCents)
  entryCostInput.value = formatCentsAsCurrency(item.costCents)
}
async function loadProduct() {
  if (!isEdit.value) return
  const version = ++itemVersion
  loadingItem.value = true
  loadError.value = ''
  try {
    const item = await stock.loadItem(String(props.productId))
    if (version !== itemVersion) return
    if (item) fillItem(item)
    else loadError.value = 'Não foi possível carregar este produto. Tente novamente.'
  } catch {
    if (version === itemVersion) loadError.value = 'Não foi possível carregar este produto. Tente novamente.'
  } finally { if (version === itemVersion) loadingItem.value = false }
}
function selectExisting(item: IStockItem) {
  ++searchVersion
  fillItem(item)
  suggestionsOpen.value = false
  formError.value = ''
  stock.error = ''
  stock.fieldErrors = {}
}
function createAnother() {
  dismissedName.value = form.name
  suggestionsOpen.value = false
  suggestions.value = []
  activeSuggestion.value = -1
}
function clearExisting() {
  selectedItem.value = null
  dismissedName.value = ''
  form.name = ''
  form.description = ''
  form.quantity = 0
  costInput.value = ''
  entryCostInput.value = ''
  form.markupPercent = Number.isFinite(initialMarkup) ? initialMarkup : 10
}
function openSuggestions() {
  if (!metadataLocked.value && !isEdit.value && form.name.trim() && dismissedName.value !== form.name) suggestionsOpen.value = true
}
function onProductKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') { suggestionsOpen.value = false; return }
  if (!suggestionsOpen.value || !suggestions.value.length) return
  if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
    event.preventDefault()
    const step = event.key === 'ArrowDown' ? 1 : -1
    activeSuggestion.value = (activeSuggestion.value + step + suggestions.value.length) % suggestions.value.length
  } else if (event.key === 'Enter' && activeSuggestion.value >= 0) {
    event.preventDefault()
    const item = suggestions.value[activeSuggestion.value]
    if (item) selectExisting(item)
  } else if (event.key === 'Enter') event.preventDefault()
}
async function save() {
  if (saving.value || stock.saving || loadingItem.value || loadError.value) return
  formError.value = ''
  stock.error = ''
  stock.fieldErrors = {}
  if (hasEntry.value && (!Number.isInteger(installmentCount.value) || installmentCount.value < 1 || installmentCount.value > 48)) { formError.value = 'Informe de 1 a 48 parcelas.'; return }
  const payload: StockForm = {
    ...form,
    costCents: currencyMaskToCents(costInput.value),
    quantity: isEdit.value ? selectedItem.value?.quantity ?? 0 : 0,
    ...(isEdit.value && !supplierMode.value ? { id: String(props.productId) } : {}),
    ...(existingSelection.value || supplierMode.value ? { existingStockItemId: selectedItem.value!.id } : {}),
    ...(hasEntry.value ? { entry: {
      supplierName: supplierName.value.trim(), quantity: Number(quantity.value), unitCostCents: unitCostCents.value,
      installments: installments.value.map((row) => ({ amountCents: currencyMaskToCents(row.amountInput), dueDate: row.dueDate, plannedMethod: row.plannedMethod }))
    } } : {})
  }
  saving.value = true
  try {
    const result = await stock.save(payload)
    if (result.status === 'success') await router.push(returnTo.value)
  } catch { formError.value = 'Não foi possível concluir o salvamento. Confira os dados e tente novamente.' }
  finally { saving.value = false }
}
watch(() => props.productId, loadProduct, { immediate: true })
watch([totalCents, installmentCount], generateInstallments, { immediate: true })
watch(firstDueDate, () => { installments.value.forEach((row, index) => { row.dueDate = addMonths(firstDueDate.value, index) }) })
watch([() => form.name, supplierName], () => {
  if (searchTimer) clearTimeout(searchTimer)
  const version = ++searchVersion
  searching.value = false
  activeSuggestion.value = -1
  suggestions.value = []
  if (isEdit.value || existingSelection.value || !form.name.trim() || dismissedName.value === form.name) { suggestionsOpen.value = false; return }
  searching.value = true
  suggestionsOpen.value = import.meta.client && document.activeElement?.id === 'stock-product-name'
  searchTimer = setTimeout(async () => {
    try {
      const result = await stock.searchProducts(form.name.trim(), supplierName.value.trim())
      if (version === searchVersion) suggestions.value = result
    } catch {
      if (version === searchVersion) suggestions.value = []
    } finally { if (version === searchVersion) searching.value = false }
  }, 250)
})
onBeforeUnmount(() => { if (searchTimer) clearTimeout(searchTimer); ++searchVersion; ++itemVersion })
</script>

<template>
  <section class="page stock-entry-page">
    <PageHeader :title="title" :subtitle="supplierMode ? 'Registre uma nova compra deste produto com o fornecedor selecionado.' : isEdit ? 'Atualize os dados e, se necessário, registre uma nova entrada com suas próprias parcelas.' : 'Escolha um produto existente ou cadastre um novo e confirme a primeira entrada.'">
      <template #actions><NuxtLink class="button button--secondary" :to="returnTo"><ArrowLeft :size="18" /> {{ supplierMode ? 'Fornecedor' : 'Produtos' }}</NuxtLink></template>
    </PageHeader>
    <p v-if="loadingItem" class="panel loading-state" role="status">Carregando produto...</p>
    <div v-else-if="loadError" class="panel loading-state" role="alert"><p>{{ loadError }}</p><button class="button button--secondary" type="button" @click="loadProduct">Tentar novamente</button></div>
    <form v-else class="stock-entry" novalidate @submit.prevent="save">
      <fieldset class="stock-entry__fieldset" :disabled="saving || stock.saving">
        <section class="panel form-section">
          <header><span>1</span><div><h2>{{ isEdit ? 'Dados do produto' : 'Produto' }}</h2><p>{{ metadataLocked ? 'Os dados do cadastro serão preservados nesta entrada.' : 'Produtos existentes podem receber novas entradas sem duplicar o cadastro.' }}</p></div></header>
          <div v-if="selectedItem" class="stock-balance"><span>Saldo atual de {{ selectedItem.name }}</span><strong>{{ selectedItem.quantity }} {{ selectedItem.quantity === 1 ? 'unidade' : 'unidades' }}</strong></div>
          <div v-if="existingSelection" class="selection-notice" role="status"><Check :size="20" /><div><strong>Adicionar quantidade ao produto existente</strong><p>{{ selectedItem?.name }}</p></div><button class="text-action" type="button" @click="clearExisting">Escolher outro</button></div>
          <div class="form-grid">
            <div class="field product-search" @focusout="(event) => { if (!(event.currentTarget as HTMLElement).contains(event.relatedTarget as Node | null)) suggestionsOpen = false }">
              <label for="stock-product-name">Produto</label>
              <input id="stock-product-name" v-model="form.name" :disabled="metadataLocked" :role="!isEdit ? 'combobox' : undefined" :aria-expanded="!isEdit ? suggestionsOpen : undefined" :aria-controls="!isEdit ? 'stock-product-suggestions' : undefined" :aria-activedescendant="suggestionsOpen && activeSuggestion >= 0 ? `stock-suggestion-${activeSuggestion}` : undefined" :aria-autocomplete="!isEdit ? 'list' : undefined" :aria-invalid="!!stock.fieldErrors.name" autocomplete="off" required placeholder="Óleo, filtro, pastilha..." @focus="openSuggestions" @keydown="onProductKeydown" />
              <small v-if="stock.fieldErrors.name" class="field-error">{{ stock.fieldErrors.name }}</small>
              <div v-if="suggestionsOpen" class="suggestions">
                <p v-if="searching" role="status"><Search :size="15" /> Buscando produtos...</p>
                <p v-else-if="!suggestions.length">Nenhum produto encontrado. Continue para criar um novo.</p>
                <p v-else>Já existe um produto parecido? Escolha para adicionar quantidade.</p>
                <ul id="stock-product-suggestions" role="listbox" aria-label="Produtos existentes">
                  <li v-for="(item, index) in suggestions" :id="`stock-suggestion-${index}`" :key="item.id" role="option" :aria-selected="activeSuggestion === index">
                    <button class="suggestion-option" :class="{ 'suggestion-option--active': activeSuggestion === index }" type="button" @click="selectExisting(item)"><strong>{{ item.name }}</strong><span>Saldo: {{ item.quantity }} · {{ formatCurrency(item.costCents) }}</span><span class="suggestion-option__action">Adicionar quantidade ao produto existente</span></button>
                  </li>
                </ul>
                <button class="text-action suggestions__create" type="button" @click="createAnother">Criar outro produto</button>
              </div>
            </div>
            <label v-if="!isEdit" class="field"><span>Fornecedor</span><input v-model="supplierName" :aria-invalid="!!stock.fieldErrors.supplierName" required placeholder="Nome do fornecedor" /><small v-if="stock.fieldErrors.supplierName" class="field-error">{{ stock.fieldErrors.supplierName }}</small></label>
            <label class="field"><span>{{ isEdit || existingSelection ? 'Custo atual' : 'Custo unitário' }}</span><input :value="costInput" :disabled="metadataLocked || (isEdit && hasEntry)" inputmode="numeric" :aria-invalid="!!stock.fieldErrors.costCents" required placeholder="R$ 0,00" @input="costInput = maskCurrency(($event.target as HTMLInputElement).value)" /><small v-if="isEdit && hasEntry">O custo médio será recalculado a partir da nova entrada.</small><small v-if="stock.fieldErrors.costCents" class="field-error">{{ stock.fieldErrors.costCents }}</small></label>
            <label class="field"><span>Margem de revenda (%)</span><input v-model.number="form.markupPercent" :disabled="metadataLocked" :aria-invalid="!!stock.fieldErrors.markupPercent" type="number" min="0" step="0.1" required /><small v-if="stock.fieldErrors.markupPercent" class="field-error">{{ stock.fieldErrors.markupPercent }}</small></label>
            <label class="field field--wide"><span>Descrição</span><textarea v-model="form.description" :disabled="metadataLocked" placeholder="Detalhes do produto" /></label>
          </div>
          <label v-if="isEdit && !supplierMode" class="entry-toggle"><input v-model="addQuantity" type="checkbox" /><span><strong>Adicionar quantidade ao estoque</strong><small>Cria uma nova entrada e suas parcelas. Deixe desmarcado para editar somente o cadastro.</small></span></label>
        </section>
        <template v-if="hasEntry">
          <section class="panel form-section">
            <header><span>2</span><div><h2>{{ isEdit || existingSelection ? 'Nova entrada' : 'Primeira entrada' }}</h2><p>A quantidade informada será somada ao saldo atual.</p></div></header>
            <div class="form-grid">
              <label v-if="isEdit" class="field"><span>Fornecedor</span><input v-model="supplierName" :disabled="supplierMode" :aria-invalid="!!stock.fieldErrors.supplierName" required placeholder="Nome do fornecedor" /><small v-if="stock.fieldErrors.supplierName" class="field-error">{{ stock.fieldErrors.supplierName }}</small></label>
              <label class="field"><span>Quantidade a adicionar</span><input v-model.number="quantity" :aria-invalid="!!stock.fieldErrors.quantity" type="number" min="1" step="1" required /><small v-if="stock.fieldErrors.quantity" class="field-error">{{ stock.fieldErrors.quantity }}</small></label>
              <label v-if="isEdit || existingSelection" class="field"><span>Custo unitário da entrada</span><input :value="entryCostInput" :aria-invalid="!!stock.fieldErrors.unitCostCents" inputmode="numeric" required placeholder="R$ 0,00" @input="entryCostInput = maskCurrency(($event.target as HTMLInputElement).value)" /><small v-if="stock.fieldErrors.unitCostCents" class="field-error">{{ stock.fieldErrors.unitCostCents }}</small></label>
            </div>
            <div class="entry-total"><span>Total da entrada</span><strong>{{ formatCurrency(totalCents) }}</strong></div>
          </section>
          <section class="panel form-section">
            <header><span>3</span><div><h2>Pagamento da entrada</h2><p>As contas serão criadas junto com a entrada. Registre a quitação em Contas a pagar.</p></div></header>
            <div class="generator">
              <label class="field"><span>Número de parcelas</span><input v-model.number="installmentCount" type="number" min="1" max="48" step="1" /></label>
              <label class="field"><span>Primeiro vencimento</span><input v-model="firstDueDate" type="date" /></label>
            </div>
            <div class="installments">
              <div v-for="(row, index) in installments" :key="index" class="installment-row">
                <strong>{{ index + 1 }}ª</strong>
                <label class="field"><span>Valor</span><input :value="row.amountInput" inputmode="numeric" required @input="row.amountInput = maskCurrency(($event.target as HTMLInputElement).value)" /></label>
                <label class="field"><span>Vencimento</span><input v-model="row.dueDate" type="date" required /></label>
                <label class="field"><span>Forma prevista</span><select v-model="row.plannedMethod"><option value="boleto">Boleto</option><option value="pix">PIX</option><option value="cash">Dinheiro</option><option value="debit_card">Débito</option><option value="credit_card">Crédito</option></select></label>
                <button class="remove" type="button" :aria-label="`Remover ${index + 1}ª parcela`" :disabled="installments.length === 1" @click="removeInstallment(index)"><Trash2 :size="17" /></button>
              </div>
            </div>
            <div class="entry-total" :class="{ 'entry-total--invalid': installmentTotal !== totalCents }"><span>Soma das parcelas</span><strong>{{ formatCurrency(installmentTotal) }}</strong></div>
            <p v-if="stock.fieldErrors.installments" class="field-error">{{ stock.fieldErrors.installments }}</p>
          </section>
        </template>
        <p v-if="formError || stock.error" class="alert-box" role="alert">{{ formError || stock.error }}</p>
        <button class="button button--primary submit" type="submit" :disabled="saving || stock.saving">{{ saving || stock.saving ? 'Salvando...' : submitLabel }}</button>
      </fieldset>
    </form>
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
