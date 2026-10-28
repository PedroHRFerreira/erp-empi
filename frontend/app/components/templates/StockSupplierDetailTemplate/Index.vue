<script setup lang="ts">
import { ArrowLeft, PackagePlus, Truck } from '@lucide/vue'
import type { IStockItem, IStockPurchase } from '../../../../server/contracts/types'
import PageHeader from '../../molecules/PageHeader/Index.vue'
import { formatCurrency, formatDate, formatDateTime } from '../../../utils/format'

const props = defineProps<{ supplierName: string }>()
const stock = useStockStore()
const loaded = ref(false)
const failed = ref(false)

async function load() {
  loaded.value = false
  if (!props.supplierName.trim()) return
  failed.value = !await stock.loadSupplier(props.supplierName)
  loaded.value = !failed.value
}
await load()

const detail = computed(() => loaded.value ? stock.supplierDetail : null)
const supplierExists = computed(() => Boolean(detail.value && (detail.value.purchases.length || detail.value.products.length)))

function purchasedQuantity(product: IStockItem) {
  return (detail.value?.purchases || []).filter((purchase) => purchase.status === 'confirmed')
    .flatMap((purchase) => purchase.items).filter((item) => item.stockItemId === product.id)
    .reduce((total, item) => total + item.quantity, 0)
}
function restockLink(product: IStockItem) {
  const returnTo = `/stock/suppliers/detail?supplierName=${encodeURIComponent(props.supplierName)}`
  return { path: `/stock/${encodeURIComponent(product.id)}/edit`, query: { replenish: '1', supplierMode: '1', supplierName: props.supplierName, returnTo } }
}
function purchaseStatus(purchase: IStockPurchase) {
  if (purchase.status === 'cancelled') return 'Cancelada'
  if (purchase.installments.length && purchase.installments.every((part) => part.status === 'paid')) return 'Quitada'
  if (purchase.installments.some((part) => part.status === 'paid')) return 'Parcialmente paga'
  return 'Pagamento pendente'
}
function statusLabel(status: string) {
  return status === 'paid' ? 'Paga' : status === 'cancelled' ? 'Cancelada' : 'Pendente'
}
function methodLabel(method?: string) {
  const labels: Record<string, string> = { boleto: 'Boleto', pix: 'PIX', cash: 'Dinheiro', debit_card: 'Débito', credit_card: 'Crédito', legacy: 'Legado' }
  return labels[method || ''] || '—'
}
function dueDate(value: string) {
  return formatDate(/^\d{4}-\d{2}-\d{2}$/.test(value) ? `${value}T12:00:00Z` : value)
}
</script>

<template>
  <section class="page supplier-detail-page">
    <PageHeader :title="supplierName || 'Fornecedor'" subtitle="Produtos fornecidos, entradas e histórico de pagamentos.">
      <template #actions><NuxtLink class="button button--secondary" to="/stock/suppliers"><ArrowLeft :size="18" /> Fornecedores</NuxtLink></template>
    </PageHeader>

    <section v-if="failed" class="panel supplier-state" role="alert">
      <strong>Não foi possível carregar o fornecedor</strong><span>{{ stock.error }}</span>
      <button class="button button--secondary" type="button" :disabled="stock.loading" @click="load">Tentar novamente</button>
    </section>
    <section v-else-if="stock.loading && !loaded" class="panel supplier-state" role="status"><strong>Carregando histórico...</strong></section>
    <section v-else-if="!supplierExists" class="panel supplier-state">
      <Truck :size="28" /><strong>Fornecedor não encontrado</strong>
      <span>Confira o fornecedor na lista de entradas cadastradas.</span>
      <NuxtLink class="button button--primary" to="/stock/suppliers">Ver fornecedores</NuxtLink>
    </section>
    <template v-else-if="detail">
      <section class="supplier-summary" aria-label="Resumo do fornecedor">
        <article class="panel"><span>Produtos / entradas</span><strong>{{ detail.productCount }} / {{ detail.purchaseCount }}</strong></article>
        <article class="panel"><span>Total comprado</span><strong>{{ formatCurrency(detail.totalCents) }}</strong></article>
        <article class="panel"><span>Total pago</span><strong>{{ formatCurrency(detail.paidCents) }}</strong></article>
        <article class="panel"><span>Total pendente</span><strong>{{ formatCurrency(detail.pendingCents) }}</strong></article>
      </section>

      <section class="panel supplier-products" aria-labelledby="supplier-products-heading">
        <header><h2 id="supplier-products-heading">Produtos deste fornecedor</h2><p>As quantidades abaixo são as unidades compradas deste fornecedor, sem entradas canceladas.</p></header>
        <p v-if="!detail.products.length" class="empty-products">Nenhum produto disponível para reposição.</p>
        <article v-for="product in detail.products" :key="product.id" class="supplier-product">
          <div><h3>{{ product.name }}</h3><span>{{ purchasedQuantity(product) }} unidades compradas deste fornecedor</span></div>
          <NuxtLink v-if="product.active" class="button button--secondary" :to="restockLink(product)"><PackagePlus :size="17" /> Adicionar quantidade<span class="visually-hidden"> a {{ product.name }}</span></NuxtLink>
          <span v-else class="purchase-status">Arquivado</span>
        </article>
      </section>

      <section class="supplier-history" aria-labelledby="supplier-history-heading">
        <h2 id="supplier-history-heading">Histórico de entradas e pagamentos</h2>
        <p v-if="!detail.purchases.length" class="panel supplier-state">Nenhuma entrada registrada.</p>
        <article v-for="purchase in detail.purchases" :key="purchase.id" class="panel supplier-purchase">
          <header>
            <div><span>{{ formatDateTime(purchase.purchasedAt) }}</span><h3>Entrada de {{ formatCurrency(purchase.totalCents) }}</h3></div>
            <span class="purchase-status" :class="{ 'purchase-status--cancelled': purchase.status === 'cancelled' }">{{ purchaseStatus(purchase) }}</span>
          </header>
          <p v-if="purchase.status === 'cancelled'" class="cancelled-note">Entrada cancelada{{ purchase.cancelledAt ? ` em ${formatDateTime(purchase.cancelledAt)}` : '' }}. Não compõe os totais confirmados.</p>
          <ul class="purchase-items">
            <li v-for="item in purchase.items" :key="item.id">
              <strong>{{ item.stockItem?.name || detail.products.find((product) => product.id === item.stockItemId)?.name || 'Produto indisponível' }}</strong>
              <span>{{ item.quantity }} un. × {{ formatCurrency(item.unitCostCents) }}</span><b>{{ formatCurrency(item.subtotalCents) }}</b>
            </li>
          </ul>
          <div class="supplier-installments" role="region" :aria-label="`Parcelas da entrada de ${formatDateTime(purchase.purchasedAt)}`" tabindex="0">
            <table>
              <thead><tr><th>Parcela</th><th>Vencimento</th><th>Valor</th><th>Forma prevista</th><th>Situação</th><th>Pagamento</th></tr></thead>
              <tbody>
                <tr v-for="part in purchase.installments" :key="part.id">
                  <td data-label="Parcela">{{ part.number }}ª</td><td data-label="Vencimento">{{ dueDate(part.dueDate) }}</td>
                  <td data-label="Valor">{{ formatCurrency(part.amountCents) }}</td><td data-label="Forma prevista">{{ methodLabel(part.plannedMethod) }}</td>
                  <td data-label="Situação"><span class="installment-status" :class="`installment-status--${part.status}`">{{ statusLabel(part.status) }}</span></td>
                  <td data-label="Pagamento"><template v-if="part.status === 'paid'">{{ methodLabel(part.paymentMethod) }}<small>{{ part.paidAt ? formatDateTime(part.paidAt) : 'Data não registrada' }}</small></template><template v-else>—</template></td>
                </tr>
              </tbody>
            </table>
          </div>
        </article>
      </section>
    </template>
  </section>
</template>

<style scoped lang="scss">
@use 'styles.module.scss';
</style>
