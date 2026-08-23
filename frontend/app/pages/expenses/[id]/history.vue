<script setup lang="ts">
import { ArrowLeft, CalendarDays, CircleDollarSign, FileClock, WalletCards } from '@lucide/vue'
import type { IExpense, IStockPurchase } from '../../../../server/contracts/types'
import { formatCurrency, formatDateTime } from '../../../utils/format'
import PageHeader from '../../../components/molecules/PageHeader/Index.vue'

const route = useRoute()
const expenseId = String(route.params.id || '')
const stockPurchaseId = expenseId.startsWith('stock-') ? expenseId.slice(6) : ''
const detailEndpoint = stockPurchaseId ? `/stock/purchases/${stockPurchaseId}` : `/expenses/${expenseId}`
const { data: detail, status, error } = await useApiFetch<IExpense | IStockPurchase>(detailEndpoint)

const expense = computed(() => stockPurchaseId ? null : detail.value as IExpense | null)
const stockPurchase = computed(() => stockPurchaseId ? detail.value as IStockPurchase | null : null)
const installments = computed(() => detail.value?.installments || [])
const paidInstallments = computed(() => installments.value.filter((row) => row.status === 'paid'))
const pendingInstallments = computed(() => installments.value.filter((row) => row.status === 'pending'))
const paidCents = computed(() => paidInstallments.value.reduce((total, row) => total + row.amountCents, 0))
const pendingCents = computed(() => pendingInstallments.value.reduce((total, row) => total + row.amountCents, 0))
const title = computed(() => expense.value?.description || stockPurchase.value?.supplierName || '')
const totalCents = computed(() => expense.value?.amountCents || stockPurchase.value?.totalCents || 0)
const category = computed(() => expense.value?.category || 'Estoque')
const occurredAt = computed(() => expense.value?.spentAt || stockPurchase.value?.purchasedAt || '')

function installmentStatus(value: string) {
  return value === 'paid' ? 'Paga' : value === 'cancelled' ? 'Cancelada' : 'Pendente'
}

function methodLabel(method?: string) {
  return ({ boleto: 'Boleto', pix: 'PIX', cash: 'Dinheiro', debit_card: 'Débito', credit_card: 'Crédito', legacy: 'Forma não registrada — legado' } as Record<string, string>)[method || ''] || '—'
}
</script>

<template>
  <section class="page expense-history-page">
    <PageHeader
      :title="detail ? `Histórico — ${title}` : 'Histórico do gasto'"
      subtitle="Vencimentos, pagamentos e situação das parcelas deste gasto."
    >
      <template #actions>
        <NuxtLink class="button button--secondary" to="/expenses"><ArrowLeft :size="18" /> Gastos</NuxtLink>
      </template>
    </PageHeader>

    <p v-if="status === 'error' || error" class="alert-box" role="alert">Não foi possível carregar o histórico deste gasto.</p>
    <section v-else-if="status === 'pending'" class="panel history-empty"><FileClock :size="24" /><strong>Carregando histórico…</strong></section>
    <section v-else-if="!detail" class="panel history-empty"><FileClock :size="24" /><strong>Gasto não encontrado</strong><span>Ele pode ter sido removido ou não está mais disponível.</span></section>
    <template v-else>
      <section class="history-summary" aria-label="Resumo do gasto">
        <article class="panel"><CircleDollarSign :size="20" /><div><span>Valor total</span><strong>{{ formatCurrency(totalCents) }}</strong><small>{{ category }}</small></div></article>
        <article class="panel"><WalletCards :size="20" /><div><span>Total pago</span><strong>{{ formatCurrency(paidCents) }}</strong><small>{{ paidInstallments.length }} parcelas pagas</small></div></article>
        <article class="panel"><CalendarDays :size="20" /><div><span>Saldo pendente</span><strong>{{ formatCurrency(pendingCents) }}</strong><small>{{ pendingInstallments.length }} parcelas pendentes</small></div></article>
      </section>

      <section class="panel expense-details">
        <dl class="entry-facts">
          <div><dt>Data do lançamento</dt><dd>{{ formatDateTime(occurredAt) }}</dd></div>
          <div><dt>Categoria</dt><dd>{{ category }}</dd></div>
          <div><dt>{{ stockPurchase ? 'Fornecedor' : 'Observações' }}</dt><dd>{{ stockPurchase?.supplierName || expense?.notes || '—' }}</dd></div>
        </dl>
      </section>

      <section v-if="!installments.length" class="panel history-empty"><FileClock :size="24" /><strong>Sem histórico de parcelas</strong><span>Este é um gasto antigo, criado antes do controle de contas a pagar.</span></section>
      <section v-else class="panel installments-table" role="region" aria-label="Parcelas do gasto" tabindex="0">
        <table>
          <thead><tr><th>Parcela</th><th>Vencimento</th><th>Valor</th><th>Forma prevista</th><th>Forma efetiva</th><th>Pagamento</th><th>Situação</th></tr></thead>
          <tbody>
            <tr v-for="row in installments" :key="row.id">
              <td>{{ row.number }}ª</td><td>{{ formatDateTime(row.dueDate) }}</td><td>{{ formatCurrency(row.amountCents) }}</td><td>{{ methodLabel(row.plannedMethod) }}</td><td>{{ methodLabel(row.paymentMethod) }}</td><td>{{ row.paidAt ? formatDateTime(row.paidAt) : '—' }}</td><td><span class="installment-status" :class="`installment-status--${row.status}`">{{ installmentStatus(row.status) }}</span></td>
            </tr>
          </tbody>
        </table>
      </section>
    </template>
  </section>
</template>

<style scoped lang="scss">
.expense-history-page{display:grid;min-width:0;gap:20px}.history-summary{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:14px}.history-summary article{display:flex;min-width:0;gap:12px;padding:18px}.history-summary article>svg{flex:0 0 auto;color:var(--watt-data)}.history-summary div{display:grid;min-width:0;gap:4px}.history-summary span,.history-summary small,dt{color:var(--watt-text-muted)}.history-summary strong{font-size:19px;overflow-wrap:anywhere}.expense-details{padding:20px}.entry-facts{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:12px;margin:0}.entry-facts div{display:grid;min-width:0;gap:5px;padding:14px;border:1px solid var(--watt-border);border-radius:10px}.entry-facts dd{margin:0;font-weight:750;overflow-wrap:anywhere}.installments-table{min-width:0;overflow-x:auto;padding:20px}.installments-table table{width:100%;min-width:880px}.installments-table th,.installments-table td{text-align:left;white-space:nowrap}.installment-status{display:inline-flex;padding:5px 9px;border-radius:999px;color:var(--watt-data);background:var(--watt-data-background);font-size:12px;font-weight:750}.installment-status--paid{color:var(--watt-success);background:var(--watt-success-background)}.installment-status--cancelled{color:var(--watt-alert);background:var(--watt-alert-background)}.history-empty{display:grid;justify-items:center;gap:8px;padding:36px;text-align:center}.history-empty span{color:var(--watt-text-muted)}.alert-box{margin:0;padding:14px 16px;border-left:4px solid var(--watt-alert);color:var(--watt-alert);background:var(--watt-alert-background)}@media(max-width:800px){.history-summary,.entry-facts{grid-template-columns:1fr}.installments-table{padding:4px 16px}.installments-table table,.installments-table thead,.installments-table tbody,.installments-table tr,.installments-table td{display:block;min-width:0}.installments-table thead{display:none}.installments-table tr{display:grid;grid-template-columns:1fr 1fr;gap:10px;padding:14px 0;border-top:1px solid var(--watt-border)}.installments-table tr:first-child{border-top:0}.installments-table td{white-space:normal;overflow-wrap:anywhere}.installments-table td::before{display:block;margin-bottom:3px;color:var(--watt-text-muted);font-size:10px;font-weight:750;text-transform:uppercase}.installments-table td:nth-child(1)::before{content:'Parcela'}.installments-table td:nth-child(2)::before{content:'Vencimento'}.installments-table td:nth-child(3)::before{content:'Valor'}.installments-table td:nth-child(4)::before{content:'Forma prevista'}.installments-table td:nth-child(5)::before{content:'Forma efetiva'}.installments-table td:nth-child(6)::before{content:'Pagamento'}.installments-table td:nth-child(7)::before{content:'Situação'}}@media(max-width:480px){.installments-table tr{grid-template-columns:1fr}.history-empty{padding:28px 18px}}
</style>
