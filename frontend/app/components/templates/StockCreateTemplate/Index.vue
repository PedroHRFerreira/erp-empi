<script setup lang="ts">
import { ArrowLeft, Trash2 } from '@lucide/vue'
import { computed, reactive, ref, watch } from 'vue'
import type { IStockItem, PayableMethod } from '../../../../server/contracts/types'
import type { StockForm } from '../../../stores/useStockStore'
import { currencyMaskToCents, maskCurrency } from '../../../utils/masks'
import { formatCurrency } from '../../../utils/format'
import { divideCents } from '../../../utils/purchaseInstallments'
import PageHeader from '../../molecules/PageHeader/Index.vue'

type InstallmentRow = { amountInput: string; dueDate: string; plannedMethod: PayableMethod }
const auth = useAuthStore()
const router = useRouter()
const stock = useStockStore()
const purchases = usePurchasesStore()
const costInput = ref('')
const supplierName = ref('')
const quantity = ref(1)
const installmentCount = ref(1)
const firstDueDate = ref(today())
const installments = ref<InstallmentRow[]>([])
const createdItem = ref<IStockItem | null>(null)
const saving = ref(false)
const form = reactive<StockForm>({ name: '', description: '', costCents: 0, markupPercent: Number(auth.user?.markupPercent ?? 10), quantity: 0 })
const unitCostCents = computed(() => currencyMaskToCents(costInput.value))
const totalCents = computed(() => unitCostCents.value * Math.max(0, Number(quantity.value || 0)))
const installmentTotal = computed(() => installments.value.reduce((sum, row) => sum + currencyMaskToCents(row.amountInput), 0))

function today() { const date = new Date(); return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}` }
function addMonths(value: string, months: number) { const [year, month, day] = value.split('-').map(Number); const date = new Date(year!, month! - 1 + months, day); return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}` }
function generateInstallments() {
  installments.value = divideCents(totalCents.value, installmentCount.value).map((amount, index) => ({
    amountInput: maskCurrency(String(amount)),
    dueDate: installments.value[index]?.dueDate || addMonths(firstDueDate.value, index),
    plannedMethod: installments.value[index]?.plannedMethod || 'boleto'
  }))
}
function removeInstallment(index: number) { if (installments.value.length <= 1) return; installments.value.splice(index, 1); installmentCount.value = installments.value.length; generateInstallments() }
async function save() {
  purchases.error = ''
  form.costCents = unitCostCents.value
  if (!supplierName.value.trim() || quantity.value < 1) { purchases.error = 'Informe o fornecedor e uma quantidade válida.'; return }
  if (!installments.value.length || installmentTotal.value !== totalCents.value) { purchases.error = 'A soma das parcelas precisa ser igual ao total da entrada.'; return }
  saving.value = true
  if (!createdItem.value) {
    const result = await stock.save({ ...form })
    if (result.status === 'error') { saving.value = false; return }
    createdItem.value = result.data as IStockItem
  }
  const saved = await purchases.createPurchase({ supplierName: supplierName.value.trim(), items: [{ stockItemId: createdItem.value.id, quantity: quantity.value, unitCostCents: unitCostCents.value }], installments: installments.value.map((row) => ({ amountCents: currencyMaskToCents(row.amountInput), dueDate: row.dueDate, plannedMethod: row.plannedMethod })) })
  saving.value = false
  if (saved) await router.push('/stock')
}
watch([totalCents, installmentCount, firstDueDate], generateInstallments, { immediate: true })
</script>

<template>
  <section class="page stock-entry-page">
    <PageHeader title="Adicionar produto ao estoque" subtitle="Cadastre o produto e as informações da compra em um único fluxo.">
      <template #actions><NuxtLink class="button button--secondary" to="/stock"><ArrowLeft :size="18" /> Estoque</NuxtLink></template>
    </PageHeader>
    <p v-if="stock.error || purchases.error" class="alert-box" role="alert">{{ stock.error || purchases.error }}</p>
    <form class="stock-entry" @submit.prevent="save">
      <section class="panel form-section">
        <header><span>1</span><div><h2>Produto e entrada</h2><p>As informações financeiras registram a primeira entrada deste produto.</p></div></header>
        <div class="form-grid">
          <label class="field"><span>Produto</span><input v-model="form.name" :disabled="!!createdItem" required placeholder="Óleo, filtro, pastilha..." /></label>
          <label class="field"><span>Fornecedor</span><input v-model="supplierName" required placeholder="Nome do fornecedor" /></label>
          <label class="field"><span>Quantidade comprada</span><input v-model.number="quantity" type="number" min="1" required /></label>
          <label class="field"><span>Custo unitário</span><input :value="costInput" inputmode="numeric" required placeholder="R$ 0,00" @input="costInput = maskCurrency(($event.target as HTMLInputElement).value)" /></label>
          <label class="field"><span>Margem de revenda (%)</span><input v-model.number="form.markupPercent" :disabled="!!createdItem" type="number" min="0" step="0.1" required /></label>
          <label class="field field--wide"><span>Descrição</span><textarea v-model="form.description" :disabled="!!createdItem" placeholder="Detalhes do produto" /></label>
        </div>
        <div class="entry-total"><span>Total da entrada</span><strong>{{ formatCurrency(totalCents) }}</strong></div>
      </section>
      <section class="panel form-section">
        <header><span>2</span><div><h2>Pagamento</h2><p>À vista ou parcelado: as contas serão criadas automaticamente.</p></div></header>
        <div class="generator">
          <label class="field"><span>Número de parcelas</span><input v-model.number="installmentCount" type="number" min="1" max="48" /></label>
          <label class="field"><span>Primeiro vencimento</span><input v-model="firstDueDate" type="date" /></label>
        </div>
        <div class="installments">
          <div v-for="(row, index) in installments" :key="index" class="installment-row">
            <strong>{{ index + 1 }}ª</strong>
            <label class="field"><span>Valor</span><input :value="row.amountInput" inputmode="numeric" @input="row.amountInput = maskCurrency(($event.target as HTMLInputElement).value)" /></label>
            <label class="field"><span>Vencimento</span><input v-model="row.dueDate" type="date" /></label>
            <label class="field"><span>Forma prevista</span><select v-model="row.plannedMethod"><option value="boleto">Boleto</option><option value="pix">PIX</option><option value="cash">Dinheiro</option><option value="debit_card">Débito</option><option value="credit_card">Crédito</option></select></label>
            <button class="remove" type="button" :aria-label="`Remover ${index + 1}ª parcela`" :disabled="installments.length === 1" @click="removeInstallment(index)"><Trash2 :size="17" /></button>
          </div>
        </div>
        <div class="entry-total" :class="{ 'entry-total--invalid': installmentTotal !== totalCents }"><span>Soma das parcelas</span><strong>{{ formatCurrency(installmentTotal) }}</strong></div>
      </section>
      <button class="button button--primary submit" type="submit" :disabled="saving">{{ saving ? 'Salvando...' : 'Cadastrar produto e confirmar entrada' }}</button>
    </form>
  </section>
</template>

<style scoped lang="scss">
.stock-entry-page,.stock-entry,.form-section{display:grid;gap:20px}.form-section{padding:22px}.form-section>header{display:flex;gap:12px}.form-section>header>span{display:grid;width:32px;height:32px;flex:0 0 32px;place-items:center;border-radius:9px;color:var(--watt-background);background:var(--watt-data);font-weight:800}.form-section h2,.form-section p{margin:0}.form-section p{color:var(--watt-text-muted)}.form-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:16px}.field--wide{grid-column:1/-1}.generator{display:grid;grid-template-columns:180px 240px;gap:16px}.installments{display:grid;gap:10px}.installment-row{display:grid;grid-template-columns:48px 1fr 1fr 1fr 42px;gap:12px;align-items:end;padding:14px;border:1px solid var(--watt-border);border-radius:12px}.installment-row>strong{align-self:center}.remove{display:grid;width:38px;height:38px;place-items:center;border:1px solid var(--watt-border);border-radius:10px;color:var(--watt-alert);background:transparent}.remove:disabled{opacity:.4}.entry-total{display:flex;justify-content:space-between;gap:16px;padding:16px;border:1px solid var(--watt-border);border-radius:12px;background:var(--watt-surface-raised)}.entry-total strong{font:700 22px 'Fira Code',monospace}.entry-total--invalid{border-color:var(--watt-alert);color:var(--watt-alert)}.submit{justify-self:end}.alert-box{margin:0;padding:14px 16px;border-left:4px solid var(--watt-alert);color:var(--watt-alert);background:var(--watt-alert-background)}@media(max-width:760px){.form-grid,.generator,.installment-row{grid-template-columns:1fr}.field--wide{grid-column:auto}.installment-row>strong{grid-column:1/-1}.submit{width:100%}.form-section{padding:16px}.entry-total{align-items:flex-start;flex-direction:column}}
</style>
