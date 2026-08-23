<script setup lang="ts">
import { ArrowLeft, CalendarDays, CircleDollarSign, PackageCheck, WalletCards } from '@lucide/vue'
import type { IStockPurchase } from '../../../../server/contracts/types'
import { formatCurrency, formatDateTime } from '../../../utils/format'
import PageHeader from '../../../components/molecules/PageHeader/Index.vue'

const route = useRoute()
const stock = useStockStore()
const purchases = usePurchasesStore()
const productId = String(route.params.id || '')
await Promise.all([stock.load(0), purchases.loadPurchases()])

const entries = computed(() => purchases.purchases.filter((purchase) => purchase.items.some((item) => item.stockItemId === productId)))
const product = computed(() => stock.items.find((item) => item.id === productId) || entries.value.flatMap((purchase) => purchase.items).find((item) => item.stockItemId === productId)?.stockItem)
const confirmedEntries = computed(() => entries.value.filter((purchase) => purchase.status === 'confirmed'))
const totalPurchasedCents = computed(() => confirmedEntries.value.reduce((total, purchase) => total + (purchase.items.find((item) => item.stockItemId === productId)?.subtotalCents || 0), 0))
const pendingInstallments = computed(() => confirmedEntries.value.flatMap((purchase) => purchase.installments).filter((row) => row.status === 'pending'))
const pendingCents = computed(() => pendingInstallments.value.reduce((total, row) => total + row.amountCents, 0))

function purchaseItem(purchase: IStockPurchase) { return purchase.items.find((item) => item.stockItemId === productId) }
function purchaseStatus(purchase: IStockPurchase) {
  if (purchase.status === 'cancelled') return 'Entrada removida'
  if (purchase.installments.every((row) => row.status === 'paid')) return 'Quitada'
  if (purchase.installments.some((row) => row.status === 'paid')) return 'Parcialmente paga'
  return 'Pagamento pendente'
}
function installmentStatus(status: string) { return status === 'paid' ? 'Paga' : status === 'cancelled' ? 'Cancelada' : 'Pendente' }
function methodLabel(method?: string) { return ({ boleto: 'Boleto', pix: 'PIX', cash: 'Dinheiro', debit_card: 'Débito', credit_card: 'Crédito', legacy: 'Forma não registrada — legado' } as Record<string, string>)[method || ''] || '—' }
</script>

<template>
  <section class="page stock-history-page">
    <PageHeader :title="product ? `Histórico — ${product.name}` : 'Histórico do produto'" subtitle="Entradas, fornecedores e pagamentos relacionados a este produto.">
      <template #actions><NuxtLink class="button button--secondary" to="/stock"><ArrowLeft :size="18" /> Estoque</NuxtLink></template>
    </PageHeader>

    <p v-if="stock.error || purchases.error" class="alert-box" role="alert">{{ stock.error || purchases.error }}</p>
    <section v-else-if="!product" class="panel history-empty"><PackageCheck :size="24" /><strong>Produto não encontrado</strong><span>Ele pode ter sido removido ou não está mais disponível.</span></section>
    <template v-else>
      <section class="history-summary" aria-label="Resumo do produto">
        <article class="panel"><PackageCheck :size="20" /><div><span>Estoque atual</span><strong>{{ product.quantity }} unidades</strong><small>{{ product.usedQuantity }} usadas</small></div></article>
        <article class="panel"><CalendarDays :size="20" /><div><span>Entradas confirmadas</span><strong>{{ confirmedEntries.length }}</strong><small>{{ entries.length }} no histórico completo</small></div></article>
        <article class="panel"><CircleDollarSign :size="20" /><div><span>Total comprado</span><strong>{{ formatCurrency(totalPurchasedCents) }}</strong><small>Entradas não removidas</small></div></article>
        <article class="panel"><WalletCards :size="20" /><div><span>Saldo pendente</span><strong>{{ formatCurrency(pendingCents) }}</strong><small>{{ pendingInstallments.length }} parcelas</small></div></article>
      </section>

      <section v-if="!entries.length" class="panel history-empty"><PackageCheck :size="24" /><strong>Nenhuma entrada registrada</strong><span>As próximas reposições aparecerão aqui com seus pagamentos.</span></section>
      <section v-else class="history-list">
        <article v-for="purchase in entries" :key="purchase.id" class="panel history-entry">
          <header>
            <div><span class="eyebrow">{{ formatDateTime(purchase.purchasedAt) }}</span><h2>{{ purchase.supplierName }}</h2></div>
            <span class="history-status" :class="{ 'history-status--cancelled': purchase.status === 'cancelled' }">{{ purchaseStatus(purchase) }}</span>
          </header>
          <dl class="entry-facts">
            <div><dt>Quantidade</dt><dd>{{ purchaseItem(purchase)?.quantity || 0 }} unidades</dd></div>
            <div><dt>Custo unitário</dt><dd>{{ formatCurrency(purchaseItem(purchase)?.unitCostCents || 0) }}</dd></div>
            <div><dt>Total da entrada</dt><dd>{{ formatCurrency(purchaseItem(purchase)?.subtotalCents || 0) }}</dd></div>
          </dl>
          <div class="installments-table" role="region" aria-label="Parcelas da entrada" tabindex="0">
            <table><thead><tr><th>Parcela</th><th>Vencimento</th><th>Valor</th><th>Forma prevista</th><th>Forma efetiva</th><th>Situação</th></tr></thead>
              <tbody><tr v-for="row in purchase.installments" :key="row.id"><td>{{ row.number }}ª</td><td>{{ formatDateTime(row.dueDate) }}</td><td>{{ formatCurrency(row.amountCents) }}</td><td>{{ methodLabel(row.plannedMethod) }}</td><td>{{ methodLabel(row.paymentMethod) }}</td><td><span class="installment-status" :class="`installment-status--${row.status}`">{{ installmentStatus(row.status) }}</span></td></tr></tbody>
            </table>
          </div>
        </article>
      </section>
    </template>
  </section>
</template>

<style scoped lang="scss">
.stock-history-page,.history-list,.history-entry{display:grid;gap:20px}.history-summary{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:14px}.history-summary article{display:flex;gap:12px;padding:18px}.history-summary article>svg{flex:0 0 auto;color:var(--watt-data)}.history-summary div{display:grid;gap:4px}.history-summary span,.history-summary small,.eyebrow,dt{color:var(--watt-text-muted)}.history-summary strong{font-size:19px}.history-entry{padding:22px}.history-entry>header{display:flex;align-items:flex-start;justify-content:space-between;gap:16px}.history-entry h2{margin:4px 0 0}.eyebrow{font-size:12px;font-weight:750;text-transform:uppercase}.history-status,.installment-status{display:inline-flex;padding:5px 9px;border-radius:999px;color:var(--watt-data);background:var(--watt-data-background);font-size:12px;font-weight:750}.history-status--cancelled,.installment-status--cancelled{color:var(--watt-alert);background:var(--watt-alert-background)}.installment-status--paid{color:var(--watt-success);background:var(--watt-success-background)}.entry-facts{display:grid;grid-template-columns:repeat(3,1fr);gap:12px;margin:0}.entry-facts div{display:grid;gap:5px;padding:14px;border:1px solid var(--watt-border);border-radius:10px}.entry-facts dd{margin:0;font-weight:750}.installments-table{overflow-x:auto}.installments-table table{width:100%;min-width:760px}.installments-table th,.installments-table td{text-align:left;white-space:nowrap}.history-empty{display:grid;justify-items:center;gap:8px;padding:36px;text-align:center}.history-empty span{color:var(--watt-text-muted)}.alert-box{margin:0;padding:14px 16px;border-left:4px solid var(--watt-alert);color:var(--watt-alert);background:var(--watt-alert-background)}@media(max-width:1000px){.history-summary{grid-template-columns:1fr 1fr}}@media(max-width:640px){.history-summary,.entry-facts{grid-template-columns:1fr}.history-entry{padding:16px}.history-entry>header{align-items:flex-start;flex-direction:column}.installments-table table,.installments-table thead,.installments-table tbody,.installments-table tr,.installments-table td{display:block;min-width:0}.installments-table thead{display:none}.installments-table tr{display:grid;grid-template-columns:1fr 1fr;gap:10px;padding:14px 0;border-top:1px solid var(--watt-border)}.installments-table td::before{display:block;margin-bottom:3px;color:var(--watt-text-muted);font-size:10px;font-weight:750;text-transform:uppercase}.installments-table td:nth-child(1)::before{content:'Parcela'}.installments-table td:nth-child(2)::before{content:'Vencimento'}.installments-table td:nth-child(3)::before{content:'Valor'}.installments-table td:nth-child(4)::before{content:'Forma prevista'}.installments-table td:nth-child(5)::before{content:'Forma efetiva'}.installments-table td:nth-child(6)::before{content:'Situação'}}
</style>
