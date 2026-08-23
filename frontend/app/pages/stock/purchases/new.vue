<script setup lang="ts">
import { ArrowLeft, Trash2 } from '@lucide/vue'
import type { PayableMethod } from '../../../../server/contracts/types'
import { currencyMaskToCents, formatCentsAsCurrency, maskCurrency } from '../../../utils/masks'
import { formatCurrency } from '../../../utils/format'
import { divideCents } from '../../../utils/purchaseInstallments'
import PageHeader from '../../../components/molecules/PageHeader/Index.vue'

type ItemRow = { key: number; stockItemId: string; quantity: number; costInput: string }
type InstallmentRow = { key: number; amountInput: string; dueDate: string; plannedMethod: PayableMethod }

let nextRowKey = 1

const router = useRouter()
const stock = useStockStore()
const purchases = usePurchasesStore()
const requestedProductId = String(useRoute().query.productId || '')
if (!requestedProductId) await navigateTo('/stock/new', { replace: true })
await Promise.all([stock.load(0), purchases.loadPurchases()])
const requestedProduct = stock.items.find((product) => product.id === requestedProductId)
const lastPurchase = purchases.purchases.find((purchase) => purchase.status === 'confirmed' && purchase.items.some((item) => item.stockItemId === requestedProductId))
const supplierName = ref(lastPurchase?.supplierName || '')
const items = ref<ItemRow[]>([{
  key: nextRowKey++,
  stockItemId: requestedProduct?.id || '',
  quantity: 1,
  costInput: requestedProduct ? formatCentsAsCurrency(requestedProduct.costCents) : ''
}])
const installments = ref<InstallmentRow[]>([])
const installmentCount = ref(1)
const firstDueDate = ref(today())
const totalCents = computed(() => items.value.reduce((sum, item) => sum + currencyMaskToCents(item.costInput) * Number(item.quantity || 0), 0))
const installmentTotal = computed(() => installments.value.reduce((sum, row) => sum + currencyMaskToCents(row.amountInput), 0))

function today() {
  const date = new Date()
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}
function addMonths(value: string, months: number) {
  const [year, month, day] = value.split('-').map(Number)
  const date = new Date(year!, month! - 1 + months, day)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}
function fillProductCost(item: ItemRow) {
  const product = stock.items.find((candidate) => candidate.id === item.stockItemId)
  item.costInput = product ? formatCentsAsCurrency(product.costCents) : ''
}
function redistributeInstallments() {
  const count = installments.value.length
  if (!count) return
  divideCents(totalCents.value, count).forEach((amount, index) => {
    const row = installments.value[index]!
    row.amountInput = maskCurrency(String(amount))
  })
}
function removeInstallment(index: number) {
  if (installments.value.length <= 1) return
  installments.value.splice(index, 1)
  installmentCount.value = installments.value.length
  redistributeInstallments()
}
function generateInstallments() {
  const count = Math.max(1, Number(installmentCount.value || 1))
  installments.value = divideCents(totalCents.value, count).map((amount, index) => {
    return { key: nextRowKey++, amountInput: maskCurrency(String(amount)), dueDate: addMonths(firstDueDate.value, index), plannedMethod: 'boleto' }
  })
}
watch(totalCents, () => {
  if (installments.value.length) redistributeInstallments()
  else generateInstallments()
})

async function save() {
  purchases.error = ''
  if (!supplierName.value.trim() || items.value.some((row) => !row.stockItemId || row.quantity < 1 || currencyMaskToCents(row.costInput) <= 0)) {
    purchases.error = 'Informe o fornecedor e preencha todos os produtos.'
    return
  }
  if (!installments.value.length || installmentTotal.value !== totalCents.value) {
    purchases.error = 'A soma das parcelas precisa ser igual ao total da compra.'
    return
  }
  const saved = await purchases.createPurchase({
    supplierName: supplierName.value,
    items: items.value.map((row) => ({ stockItemId: row.stockItemId, quantity: row.quantity, unitCostCents: currencyMaskToCents(row.costInput) })),
    installments: installments.value.map((row) => ({ amountCents: currencyMaskToCents(row.amountInput), dueDate: row.dueDate, plannedMethod: row.plannedMethod }))
  })
  if (saved) await router.push('/stock')
}
</script>

<template>
  <section class="page purchase-form-page">
    <PageHeader :title="`Adicionar estoque${requestedProduct ? ` — ${requestedProduct.name}` : ''}`" subtitle="Confirme a reposição e programe o pagamento.">
      <template #actions><NuxtLink class="button button--secondary" to="/stock"><ArrowLeft :size="18" /> Estoque</NuxtLink></template>
    </PageHeader>
    <p v-if="purchases.error" class="alert-box" role="alert">{{ purchases.error }}</p>

    <form class="purchase-form" @submit.prevent="save">
      <section class="panel form-section">
        <header><span>1</span><div><h2>Fornecedor e produto</h2><p>Fornecedor e custo vêm sugeridos pela última entrada e podem ser alterados.</p></div></header>
        <label class="field"><span>Fornecedor</span><input v-model="supplierName" required placeholder="Nome do fornecedor" /></label>
        <div class="items-table" tabindex="0" aria-label="Itens da compra">
          <div class="items-head"><span>Produto</span><span>Quantidade</span><span>Custo unitário</span><span>Subtotal</span><span></span></div>
          <div v-for="(item, index) in items" :key="item.key" class="items-row">
            <select v-model="item.stockItemId" aria-label="Produto" disabled @change="fillProductCost(item)"><option value="">Selecione</option><option v-for="product in stock.items" :key="product.id" :value="product.id">{{ product.name }}</option></select>
            <input v-model.number="item.quantity" aria-label="Quantidade" type="number" min="1" />
            <input :value="item.costInput" aria-label="Custo unitário" inputmode="numeric" placeholder="R$ 0,00" @input="item.costInput = maskCurrency(($event.target as HTMLInputElement).value)" />
            <strong>{{ formatCurrency(currencyMaskToCents(item.costInput) * Number(item.quantity || 0)) }}</strong>
            <span></span>
          </div>
        </div>
        <div class="purchase-total"><span>Total da entrada</span><strong>{{ formatCurrency(totalCents) }}</strong></div>
      </section>

      <section class="panel form-section">
        <header><span>2</span><div><h2>Parcelas</h2><p>Gere a divisão e ajuste valores, vencimentos ou formas previstas.</p></div></header>
        <div class="generator"><label class="field"><span>Número de parcelas</span><input v-model.number="installmentCount" type="number" min="1" max="48" /></label><label class="field"><span>Primeiro vencimento</span><input v-model="firstDueDate" type="date" /></label><button class="button button--secondary" type="button" @click="generateInstallments">Gerar parcelas</button></div>
        <div class="installments">
          <div v-for="(row, index) in installments" :key="row.key" class="installment-row">
            <strong>{{ index + 1 }}ª</strong>
            <label class="field"><span>Valor</span><input :value="row.amountInput" inputmode="numeric" @input="row.amountInput = maskCurrency(($event.target as HTMLInputElement).value)" /></label>
            <label class="field"><span>Vencimento</span><input v-model="row.dueDate" type="date" /></label>
            <label class="field"><span>Forma prevista</span><select v-model="row.plannedMethod"><option value="boleto">Boleto</option><option value="pix">PIX</option><option value="cash">Dinheiro</option><option value="debit_card">Débito</option><option value="credit_card">Crédito</option></select></label>
            <button class="remove installment-remove" type="button" :aria-label="`Remover ${index + 1}ª parcela`" :disabled="installments.length === 1" @click="removeInstallment(index)"><Trash2 :size="17" /></button>
          </div>
        </div>
        <div :class="{ 'sum--invalid': installmentTotal !== totalCents }" class="installment-sum"><span>Soma das parcelas</span><strong>{{ formatCurrency(installmentTotal) }}</strong></div>
      </section>
      <button class="button button--primary submit" type="submit">Confirmar entrada e gerar contas</button>
    </form>
  </section>
</template>

<style scoped lang="scss">
.purchase-form-page,.purchase-form,.form-section{display:grid;gap:20px}.form-section{padding:22px}.form-section>header{display:flex;gap:12px}.form-section>header>span{display:grid;width:32px;height:32px;place-items:center;border-radius:9px;color:var(--watt-background);background:var(--watt-data);font-weight:800}.form-section h2,.form-section p{margin:0}.form-section p{color:var(--watt-text-muted)}.items-head,.items-row{display:grid;grid-template-columns:minmax(190px,1.4fr) 110px 160px 140px 42px;gap:10px;align-items:center}.items-head{padding:8px 0;color:var(--watt-text-muted);font-size:11px;text-transform:uppercase}.items-row{padding:10px 0;border-top:1px solid var(--watt-border)}.items-row input,.items-row select{min-height:40px;border:1px solid var(--watt-border);border-radius:10px;padding:8px;color:var(--watt-text);background:var(--watt-surface)}.remove{display:grid;width:38px;height:38px;place-items:center;border:1px solid var(--watt-border);border-radius:10px;color:var(--watt-alert);background:transparent}.remove:disabled{cursor:not-allowed;opacity:.4}.add-row{width:max-content}.purchase-total,.installment-sum{display:flex;justify-content:space-between;padding:16px;border:1px solid var(--watt-border);border-radius:12px;background:var(--watt-surface-raised)}.purchase-total strong,.installment-sum strong{font:700 24px 'Fira Code',monospace}.generator{display:grid;grid-template-columns:160px 220px auto;gap:12px;align-items:end}.installments{display:grid;gap:10px}.installment-row{display:grid;grid-template-columns:48px minmax(130px,1fr) minmax(150px,1fr) minmax(150px,1fr) 42px;gap:12px;align-items:end;padding:14px;border:1px solid var(--watt-border);border-radius:12px}.installment-row>strong{align-self:center}.installment-remove{margin-bottom:1px}.sum--invalid{border-color:var(--watt-alert);color:var(--watt-alert)}.submit{justify-self:end}.alert-box{margin:0;padding:14px 16px;border-left:4px solid var(--watt-alert);color:var(--watt-alert);background:var(--watt-alert-background)}@media(max-width:900px){.items-table{overflow-x:auto}.items-head,.items-row{min-width:760px}.generator,.installment-row{grid-template-columns:1fr 1fr}.installment-row>strong{grid-column:1/-1}.installment-remove{justify-self:end}}@media(max-width:640px){.generator,.installment-row{grid-template-columns:1fr}.installment-remove{justify-self:start}.submit{width:100%}}
.items-table { max-width: 100%; overscroll-behavior-x: contain; }
.items-table:focus-visible { outline-offset: 3px; }
.purchase-total,.installment-sum { gap: 16px; }
.purchase-total strong,.installment-sum strong { min-width: 0; overflow-wrap: anywhere; text-align: right; }
@media(max-width:640px){
  .form-section{padding:16px}
  .form-section>header>span{flex:0 0 32px}
  .purchase-total,.installment-sum{align-items:flex-start;flex-direction:column}
  .purchase-total strong,.installment-sum strong{text-align:left}
}
</style>
