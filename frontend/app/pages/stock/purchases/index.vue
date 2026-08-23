<script setup lang="ts">
import {
  AlertTriangle,
  ArrowLeft,
  CalendarDays,
  CircleDollarSign,
  PackageCheck,
  Plus,
  RotateCcw,
  Search,
  ShoppingCart,
  WalletCards,
  XCircle
} from '@lucide/vue'
import type { IStockPurchase } from '../../../../server/contracts/types'
import PageHeader from '../../../components/molecules/PageHeader/Index.vue'
import { formatCurrency, formatDate } from '../../../utils/format'

type PeriodFilter = 'month' | '30_days' | 'year' | 'all'
type StatusFilter = 'all' | 'open' | 'paid' | 'overdue' | 'cancelled'

const purchases = usePurchasesStore()
const supplierQuery = ref('')
const periodFilter = ref<PeriodFilter>('month')
const statusFilter = ref<StatusFilter>('all')
const cancellingPurchaseId = ref('')

await purchases.loadPurchases()

const today = computed(() => {
  const now = new Date()
  return new Date(now.getFullYear(), now.getMonth(), now.getDate())
})

function purchaseDate(value: string) {
  const datePart = value.slice(0, 10)
  const [year, month, day] = datePart.split('-').map(Number)
  return year && month && day ? new Date(year, month - 1, day) : new Date(value)
}

function isWithinPeriod(purchase: IStockPurchase) {
  if (periodFilter.value === 'all') return true

  const date = purchaseDate(purchase.purchasedAt)
  const now = today.value

  if (periodFilter.value === 'month') {
    return date.getFullYear() === now.getFullYear() && date.getMonth() === now.getMonth()
  }

  if (periodFilter.value === 'year') return date.getFullYear() === now.getFullYear()

  const start = new Date(now)
  start.setDate(start.getDate() - 29)
  return date >= start && date <= now
}

function isOverdue(dueDate: string) {
  return purchaseDate(dueDate) < today.value
}

function paidInstallments(purchase: IStockPurchase) {
  return purchase.installments.filter((installment) => installment.status === 'paid').length
}

function pendingInstallments(purchase: IStockPurchase) {
  return purchase.installments.filter((installment) => installment.status === 'pending')
}

function pendingCents(purchase: IStockPurchase) {
  return pendingInstallments(purchase).reduce((total, installment) => total + installment.amountCents, 0)
}

function purchaseState(purchase: IStockPurchase): Exclude<StatusFilter, 'all'> {
  if (purchase.status === 'cancelled') return 'cancelled'
  const pending = pendingInstallments(purchase)
  if (pending.some((installment) => isOverdue(installment.dueDate))) return 'overdue'
  return pending.length ? 'open' : 'paid'
}

function stateLabel(purchase: IStockPurchase) {
  const labels: Record<Exclude<StatusFilter, 'all'>, string> = {
    open: 'Em aberto',
    paid: 'Quitada',
    overdue: 'Atrasada',
    cancelled: 'Cancelada'
  }
  return labels[purchaseState(purchase)]
}

function itemSummary(purchase: IStockPurchase) {
  const units = purchase.items.reduce((sum, item) => sum + item.quantity, 0)
  const products = purchase.items.length
  return `${units} un. · ${products} ${products === 1 ? 'produto' : 'produtos'}`
}

const periodPurchases = computed(() => {
  const query = supplierQuery.value.trim().toLocaleLowerCase('pt-BR')
  return purchases.purchases.filter((purchase) => {
    const matchesSupplier = !query || purchase.supplierName.toLocaleLowerCase('pt-BR').includes(query)
    return matchesSupplier && isWithinPeriod(purchase)
  })
})

const visiblePurchases = computed(() => periodPurchases.value.filter((purchase) => {
  return statusFilter.value === 'all' || purchaseState(purchase) === statusFilter.value
}))

const confirmedPurchases = computed(() => periodPurchases.value.filter((purchase) => purchase.status === 'confirmed'))
const totalPurchasedCents = computed(() => confirmedPurchases.value.reduce((total, purchase) => total + purchase.totalCents, 0))
const totalPendingCents = computed(() => confirmedPurchases.value.reduce((total, purchase) => total + pendingCents(purchase), 0))
const overdueCount = computed(() => confirmedPurchases.value.reduce((total, purchase) => {
  return total + pendingInstallments(purchase).filter((installment) => isOverdue(installment.dueDate)).length
}, 0))

const hasFilters = computed(() => supplierQuery.value.trim() !== '' || periodFilter.value !== 'month' || statusFilter.value !== 'all')

function clearFilters() {
  supplierQuery.value = ''
  periodFilter.value = 'month'
  statusFilter.value = 'all'
}

async function retry() {
  purchases.error = ''
  await purchases.loadPurchases()
}

async function cancel(purchase: IStockPurchase) {
  if (cancellingPurchaseId.value) return
  const message = `Cancelar a compra de ${formatCurrency(purchase.totalCents)} de ${purchase.supplierName} e reverter as unidades recebidas?`
  if (!window.confirm(message)) return
  purchases.error = ''
  cancellingPurchaseId.value = purchase.id
  await purchases.cancelPurchase(purchase.id)
  cancellingPurchaseId.value = ''
}
</script>

<template>
  <section class="page purchase-page">
    <PageHeader title="Compras de estoque" subtitle="Acompanhe entradas, fornecedores e pagamentos do estoque em um só lugar.">
      <template #actions>
        <NuxtLink class="button button--secondary" to="/stock"><ArrowLeft :size="18" /> Produtos</NuxtLink>
        <NuxtLink class="button button--primary" to="/stock/purchases/new"><Plus :size="18" /> Nova compra</NuxtLink>
      </template>
    </PageHeader>

    <section class="purchase-kpis" aria-label="Resumo das compras">
      <article class="panel purchase-kpi">
        <span class="purchase-kpi__icon"><PackageCheck :size="19" /></span>
        <div><span>Compras confirmadas</span><strong>{{ confirmedPurchases.length }}</strong><small>No período selecionado</small></div>
      </article>
      <article class="panel purchase-kpi">
        <span class="purchase-kpi__icon"><CircleDollarSign :size="19" /></span>
        <div><span>Total comprado</span><strong>{{ formatCurrency(totalPurchasedCents) }}</strong><small>Compras não canceladas</small></div>
      </article>
      <article class="panel purchase-kpi">
        <span class="purchase-kpi__icon"><WalletCards :size="19" /></span>
        <div><span>Saldo pendente</span><strong>{{ formatCurrency(totalPendingCents) }}</strong><small>Ainda não saiu do caixa</small></div>
      </article>
      <article class="panel purchase-kpi" :class="{ 'purchase-kpi--alert': overdueCount > 0 }">
        <span class="purchase-kpi__icon"><AlertTriangle :size="19" /></span>
        <div><span>Parcelas atrasadas</span><strong>{{ overdueCount }}</strong><small>{{ overdueCount ? 'Precisam de atenção' : 'Nenhum atraso no período' }}</small></div>
      </article>
    </section>

    <section class="panel purchase-filters" aria-label="Filtros do histórico">
      <label class="purchase-search">
        <span>Buscar fornecedor</span>
        <span class="purchase-search__control"><Search :size="17" /><input v-model="supplierQuery" type="search" placeholder="Nome do fornecedor"></span>
      </label>
      <label>
        <span>Período</span>
        <select v-model="periodFilter">
          <option value="month">Este mês</option>
          <option value="30_days">Últimos 30 dias</option>
          <option value="year">Este ano</option>
          <option value="all">Todo o histórico</option>
        </select>
      </label>
      <label>
        <span>Situação</span>
        <select v-model="statusFilter">
          <option value="all">Todas</option>
          <option value="open">Em aberto</option>
          <option value="overdue">Atrasadas</option>
          <option value="paid">Quitadas</option>
          <option value="cancelled">Canceladas</option>
        </select>
      </label>
      <button v-if="hasFilters" class="button button--secondary purchase-filters__clear" type="button" @click="clearFilters">
        <RotateCcw :size="16" /> Limpar filtros
      </button>
    </section>

    <section v-if="purchases.error" class="alert-box" role="alert">
      <div><strong>Não foi possível carregar as compras</strong><span>{{ purchases.error }}</span></div>
      <button class="button button--secondary" type="button" :disabled="purchases.loading" @click="retry">Tentar novamente</button>
    </section>

    <section v-if="purchases.loading && !purchases.purchases.length" class="panel purchase-state" aria-live="polite">
      <ShoppingCart :size="24" />
      <strong>Carregando compras...</strong>
      <span>Aguarde enquanto buscamos o histórico.</span>
    </section>

    <section v-else-if="!purchases.purchases.length && !purchases.error" class="panel purchase-state">
      <ShoppingCart :size="24" />
      <strong>Nenhuma compra registrada</strong>
      <span>Registre uma entrada para aumentar o estoque e gerar as contas a pagar.</span>
      <NuxtLink class="button button--primary" to="/stock/purchases/new"><Plus :size="17" /> Registrar primeira compra</NuxtLink>
    </section>

    <section v-else-if="purchases.error && !purchases.purchases.length" class="panel purchase-state purchase-state--error">
      <AlertTriangle :size="24" />
      <strong>Histórico indisponível</strong>
      <span>Tente carregar novamente para consultar suas compras.</span>
    </section>

    <section v-else-if="!visiblePurchases.length" class="panel purchase-state">
      <Search :size="24" />
      <strong>Nenhuma compra encontrada</strong>
      <span>Altere os filtros para consultar outros registros.</span>
      <button class="button button--secondary" type="button" @click="clearFilters">Limpar filtros</button>
    </section>

    <section v-else class="panel purchase-history" aria-label="Histórico de compras de estoque">
      <header class="purchase-history__header">
        <div><span class="eyebrow">Histórico</span><h2>Compras encontradas</h2></div>
        <span>{{ visiblePurchases.length }} {{ visiblePurchases.length === 1 ? 'registro' : 'registros' }}</span>
      </header>

      <div class="purchase-table table-wrap" tabindex="0">
        <table>
          <thead><tr><th>Fornecedor e data</th><th>Itens recebidos</th><th>Progresso</th><th>Total</th><th>Situação</th><th>Ação</th></tr></thead>
          <tbody>
            <tr v-for="purchase in visiblePurchases" :key="purchase.id">
              <td><strong>{{ purchase.supplierName }}</strong><small><CalendarDays :size="13" /> {{ formatDate(purchase.purchasedAt) }}</small></td>
              <td><strong>{{ itemSummary(purchase) }}</strong><small v-if="purchase.items[0]?.stockItem?.name">{{ purchase.items.map((item) => item.stockItem?.name).filter(Boolean).join(', ') }}</small></td>
              <td>
                <div class="installment-progress">
                  <span><strong>{{ paidInstallments(purchase) }}</strong> de {{ purchase.installments.length }} pagas</span>
                  <progress :value="paidInstallments(purchase)" :max="purchase.installments.length || 1">{{ paidInstallments(purchase) }}/{{ purchase.installments.length }}</progress>
                  <small v-if="pendingCents(purchase)">{{ formatCurrency(pendingCents(purchase)) }} pendentes</small>
                  <small v-else>Sem saldo pendente</small>
                </div>
              </td>
              <td class="numeric">{{ formatCurrency(purchase.totalCents) }}</td>
              <td><span class="status" :class="`status--${purchaseState(purchase)}`">{{ stateLabel(purchase) }}</span></td>
              <td>
                <button v-if="purchase.status === 'confirmed' && paidInstallments(purchase) === 0" class="cancel-action" type="button" :disabled="Boolean(cancellingPurchaseId)" @click="cancel(purchase)">
                  <XCircle :size="16" /> {{ cancellingPurchaseId === purchase.id ? 'Cancelando...' : 'Cancelar' }}
                </button>
                <span v-else class="no-action">Sem ações</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="purchase-cards">
        <article v-for="purchase in visiblePurchases" :key="purchase.id" class="purchase-card">
          <header>
            <div><strong>{{ purchase.supplierName }}</strong><span><CalendarDays :size="14" /> {{ formatDate(purchase.purchasedAt) }}</span></div>
            <span class="status" :class="`status--${purchaseState(purchase)}`">{{ stateLabel(purchase) }}</span>
          </header>
          <dl>
            <div><dt>Itens recebidos</dt><dd>{{ itemSummary(purchase) }}</dd></div>
            <div><dt>Total da compra</dt><dd class="numeric">{{ formatCurrency(purchase.totalCents) }}</dd></div>
          </dl>
          <div class="installment-progress">
            <span><strong>{{ paidInstallments(purchase) }}</strong> de {{ purchase.installments.length }} parcelas pagas</span>
            <progress :value="paidInstallments(purchase)" :max="purchase.installments.length || 1">{{ paidInstallments(purchase) }}/{{ purchase.installments.length }}</progress>
            <small>{{ pendingCents(purchase) ? `${formatCurrency(pendingCents(purchase))} pendentes` : 'Sem saldo pendente' }}</small>
          </div>
          <button v-if="purchase.status === 'confirmed' && paidInstallments(purchase) === 0" class="cancel-action" type="button" :disabled="Boolean(cancellingPurchaseId)" @click="cancel(purchase)">
            <XCircle :size="16" /> {{ cancellingPurchaseId === purchase.id ? 'Cancelando...' : 'Cancelar compra' }}
          </button>
        </article>
      </div>
    </section>
  </section>
</template>

<style scoped lang="scss">
.purchase-page { min-width: 0; gap: 20px; }
.purchase-kpis { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 16px; }
.purchase-kpi { display: flex; min-width: 0; gap: 14px; padding: 20px; }
.purchase-kpi__icon { display: grid; width: 38px; height: 38px; flex: 0 0 auto; place-items: center; border: 1px solid color-mix(in srgb, var(--watt-data) 35%, var(--watt-border)); border-radius: 11px; color: var(--watt-data); background: color-mix(in srgb, var(--watt-data) 9%, transparent); }
.purchase-kpi > div { display: grid; min-width: 0; gap: 5px; }
.purchase-kpi > div > span, .purchase-kpi small { color: var(--watt-text-muted); font-size: 12px; }
.purchase-kpi strong { overflow: hidden; font: 700 23px/1.2 'Fira Code', monospace; text-overflow: ellipsis; white-space: nowrap; }
.purchase-kpi--alert { border-color: color-mix(in srgb, var(--watt-alert) 50%, var(--watt-border)); }
.purchase-kpi--alert .purchase-kpi__icon, .purchase-kpi--alert strong { color: var(--watt-alert); }
.purchase-kpi--alert .purchase-kpi__icon { border-color: color-mix(in srgb, var(--watt-alert) 45%, var(--watt-border)); background: var(--watt-alert-background); }
.purchase-filters { display: grid; grid-template-columns: minmax(260px, 1fr) minmax(160px, 210px) minmax(160px, 210px) auto; align-items: end; gap: 16px; padding: 20px; }
.purchase-filters label { display: grid; gap: 7px; color: var(--watt-text-muted); font-size: 12px; font-weight: 600; }
.purchase-filters input, .purchase-filters select { width: 100%; min-height: 42px; border: 1px solid var(--watt-border); border-radius: 10px; padding: 0 12px; color: var(--watt-text); background: var(--watt-surface); }
.purchase-search__control { display: flex; align-items: center; gap: 8px; min-height: 42px; border: 1px solid var(--watt-border); border-radius: 10px; padding-left: 12px; color: var(--watt-text-muted); background: var(--watt-surface); }
.purchase-search__control:focus-within { border-color: var(--watt-data); box-shadow: 0 0 0 2px color-mix(in srgb, var(--watt-data) 16%, transparent); }
.purchase-search__control input { min-width: 0; border: 0; padding-left: 0; outline: 0; background: transparent; }
.purchase-filters__clear { min-height: 42px; white-space: nowrap; }
.alert-box { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin: 0; padding: 16px 18px; border-left: 4px solid var(--watt-alert); color: var(--watt-alert); background: var(--watt-alert-background); }
.alert-box div { display: grid; gap: 3px; }.alert-box span { color: var(--watt-text-muted); }
.purchase-state { display: grid; min-height: 240px; place-items: center; align-content: center; gap: 9px; padding: 32px; color: var(--watt-text-muted); text-align: center; }
.purchase-state strong { color: var(--watt-text); }.purchase-state span { max-width: 460px; }.purchase-state .button { margin-top: 8px; }
.purchase-state--error > svg { color: var(--watt-alert); }
.purchase-history { min-width: 0; overflow: hidden; }
.purchase-history__header { display: flex; align-items: center; justify-content: space-between; gap: 16px; padding: 20px; border-bottom: 1px solid var(--watt-border); }
.purchase-history__header h2 { margin: 4px 0 0; font-size: 20px; }.purchase-history__header > span, .eyebrow { color: var(--watt-text-muted); font-size: 12px; }.eyebrow { letter-spacing: .08em; text-transform: uppercase; }
.purchase-table { border: 0; border-radius: 0; }
.purchase-table td:first-child, .purchase-table td:nth-child(2) { display: table-cell; }
.purchase-table td > strong, .purchase-table td > small { display: block; }.purchase-table td > small { max-width: 260px; margin-top: 5px; overflow: hidden; color: var(--watt-text-muted); text-overflow: ellipsis; white-space: nowrap; }.purchase-table td:first-child small { display: flex; align-items: center; gap: 5px; }
.numeric { font-family: 'Fira Code', monospace; font-weight: 700; white-space: nowrap; }
.installment-progress { display: grid; min-width: 150px; gap: 6px; }.installment-progress > span { font-size: 13px; }.installment-progress small { color: var(--watt-text-muted); font-size: 11px; }
progress { width: 100%; height: 5px; overflow: hidden; border: 0; border-radius: 999px; color: var(--watt-success); background: var(--watt-surface-raised); }progress::-webkit-progress-bar { border-radius: 999px; background: var(--watt-surface-raised); }progress::-webkit-progress-value { border-radius: 999px; background: var(--watt-success); }progress::-moz-progress-bar { border-radius: 999px; background: var(--watt-success); }
.status { display: inline-flex; padding: 6px 9px; border-radius: 8px; color: var(--watt-text-muted); background: var(--watt-surface-raised); white-space: nowrap; }.status--open { color: var(--watt-data); }.status--paid { color: var(--watt-success); background: color-mix(in srgb, var(--watt-success) 12%, transparent); }.status--overdue { color: var(--watt-alert); background: var(--watt-alert-background); }.status--cancelled { text-decoration: line-through; }
.cancel-action { display: inline-flex; align-items: center; gap: 6px; min-height: 36px; border: 1px solid color-mix(in srgb, var(--watt-alert) 45%, var(--watt-border)); border-radius: 9px; padding: 7px 10px; color: var(--watt-alert); background: transparent; font-weight: 600; white-space: nowrap; }.cancel-action:hover { background: var(--watt-alert-background); }.no-action { color: var(--watt-text-muted); font-size: 12px; white-space: nowrap; }
.cancel-action:disabled { cursor: wait; opacity: .6; }
.purchase-cards { display: none; }
@media (max-width: 1180px) { .purchase-kpis { grid-template-columns: repeat(2, minmax(0, 1fr)); }.purchase-filters { grid-template-columns: minmax(240px, 1fr) repeat(2, minmax(150px, 190px)); }.purchase-filters__clear { grid-column: 1 / -1; justify-self: start; } }
@media (max-width: 760px) { .purchase-kpis { grid-template-columns: 1fr; }.purchase-filters { grid-template-columns: 1fr; }.purchase-filters__clear { width: 100%; justify-content: center; }.alert-box { align-items: stretch; flex-direction: column; }.purchase-table { display: none; }.purchase-cards { display: grid; gap: 12px; padding: 12px; }.purchase-card { display: grid; gap: 18px; padding: 16px; border: 1px solid var(--watt-border); border-radius: 12px; background: var(--watt-surface); }.purchase-card header { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; }.purchase-card header > div { display: grid; gap: 5px; }.purchase-card header span:not(.status) { display: flex; align-items: center; gap: 5px; color: var(--watt-text-muted); font-size: 12px; }.purchase-card dl { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; margin: 0; }.purchase-card dl div { display: grid; gap: 5px; }.purchase-card dt { color: var(--watt-text-muted); font-size: 11px; }.purchase-card dd { margin: 0; }.purchase-card .cancel-action { justify-content: center; width: 100%; }.purchase-history__header { padding: 16px; } }
@media (max-width: 420px) { .purchase-card dl { grid-template-columns: 1fr; }.purchase-kpi strong { font-size: 21px; } }
</style>
