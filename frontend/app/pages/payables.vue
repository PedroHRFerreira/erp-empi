<script setup lang="ts">
import { AlertTriangle, CalendarClock, CheckCircle2, RotateCcw, Search, Sparkles, WalletCards, X } from '@lucide/vue'
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import type { IPayableInstallment, PaymentMethod } from '../../server/contracts/types'
import PageHeader from '../components/molecules/PageHeader/Index.vue'
import { formatCurrency, formatDate } from '../utils/format'

type StatusFilter = 'all' | 'overdue' | 'due_soon' | 'pending'
type PeriodFilter = 'all' | '7' | '30' | '90'

const purchases = usePurchasesStore()
const selectedMethods = reactive<Record<string, PaymentMethod>>({})
const paymentDate = ref('')
const loading = ref(true)
const loadError = ref('')
const search = ref('')
const statusFilter = ref<StatusFilter>('all')
const periodFilter = ref<PeriodFilter>('all')
const pendingPayment = ref<IPayableInstallment | null>(null)
const paying = ref(false)
const feedback = ref('')
const confirmDialog = ref<HTMLElement | null>(null)
const confirmCloseButton = ref<HTMLButtonElement | null>(null)
let paymentTrigger: HTMLElement | null = null

const today = computed(() => {
  const now = new Date()
  now.setHours(0, 0, 0, 0)
  return now
})

const openTotal = computed(() => purchases.payables.reduce((sum, row) => sum + row.amountCents, 0))
const overdueRows = computed(() => purchases.payables.filter((row) => state(row).className === 'overdue'))
const overdueTotal = computed(() => overdueRows.value.reduce((sum, row) => sum + row.amountCents, 0))
const nextSevenDays = computed(() => purchases.payables.filter((row) => {
  const days = daysUntil(row.dueDate)
  return days >= 0 && days <= 7
}))
const nextSevenDaysTotal = computed(() => nextSevenDays.value.reduce((sum, row) => sum + row.amountCents, 0))

const alertGroups = computed(() => {
  const labels = { overdue: 'Pagamento atrasado', due_today: 'Vence hoje', due_tomorrow: 'Vence amanhã', early_payment: 'Antecipação possível' }
  return Object.entries(purchases.alerts.reduce((groups, alert) => {
    const group = groups[alert.kind] || { count: 0, totalCents: 0 }
    group.count += 1
    group.totalCents += alert.amountCents
    groups[alert.kind] = group
    return groups
  }, {} as Record<string, { count: number; totalCents: number }>)).map(([kind, group]) => ({ kind, label: labels[kind as keyof typeof labels], ...group }))
})

const filteredPayables = computed(() => {
  const term = search.value.trim().toLocaleLowerCase('pt-BR')
  return purchases.payables.filter((row) => {
    const rowState = state(row).className
    const matchesSearch = !term || payableName(row).toLocaleLowerCase('pt-BR').includes(term)
    const matchesStatus = statusFilter.value === 'all'
      || (statusFilter.value === 'overdue' && rowState === 'overdue')
      || (statusFilter.value === 'due_soon' && ['today', 'tomorrow'].includes(rowState))
      || (statusFilter.value === 'pending' && ['pending', 'early'].includes(rowState))
    const days = daysUntil(row.dueDate)
    const matchesPeriod = periodFilter.value === 'all' || (days >= 0 && days <= Number(periodFilter.value))
    return matchesSearch && matchesStatus && matchesPeriod
  })
})

const hasFilters = computed(() => search.value.trim() !== '' || statusFilter.value !== 'all' || periodFilter.value !== 'all')

async function loadPage() {
  loading.value = true
  loadError.value = ''
  purchases.error = ''
  const [payablesLoaded] = await Promise.all([purchases.loadPayables(), purchases.loadAlerts(true)])
  if (!payablesLoaded) loadError.value = purchases.error || 'Não foi possível carregar as contas a pagar.'
  for (const row of purchases.payables) selectedMethods[row.id] = row.plannedMethod === 'boleto' ? 'pix' : row.plannedMethod
  loading.value = false
}

function clearFilters() {
  search.value = ''
  statusFilter.value = 'all'
  periodFilter.value = 'all'
}

onMounted(async () => {
  paymentDate.value = localDateInputValue(new Date())
  await loadPage()
})

function localDateInputValue(value: Date) {
  const year = value.getFullYear()
  const month = String(value.getMonth() + 1).padStart(2, '0')
  const day = String(value.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function parseDate(value: string) {
  return new Date(`${value.slice(0, 10)}T12:00:00`)
}

function daysUntil(value: string) {
  return Math.ceil((parseDate(value).getTime() - today.value.getTime()) / 86_400_000)
}

function methodLabel(method: string) {
  return ({ boleto: 'Boleto', cash: 'Dinheiro', pix: 'PIX', debit_card: 'Débito', credit_card: 'Crédito' } as Record<string, string>)[method] || method
}

function alertFor(id: string) {
  return purchases.alerts.find((alert) => alert.installmentId === id)
}

function payableName(row: IPayableInstallment) {
  return row.expense?.description || row.stockPurchase?.supplierName || 'Origem não informada'
}

function payableKind(row: IPayableInstallment) {
  return row.expense ? 'Gasto operacional' : 'Compra de estoque'
}

function installmentTotal(row: IPayableInstallment) {
  return row.expense?.installments?.length || row.stockPurchase?.installments?.length || '?'
}

function state(row: IPayableInstallment) {
  const alert = alertFor(row.id)
  if (!alert) return { label: 'Pendente', className: 'pending' }
  if (alert.kind === 'overdue') return { label: 'Atrasada', className: 'overdue' }
  if (alert.kind === 'due_today') return { label: 'Vence hoje', className: 'today' }
  if (alert.kind === 'due_tomorrow') return { label: 'Vence amanhã', className: 'tomorrow' }
  return { label: 'Pode antecipar', className: 'early' }
}

function requestPayment(row: IPayableInstallment) {
  feedback.value = ''
  paymentTrigger = document.activeElement instanceof HTMLElement ? document.activeElement : null
  pendingPayment.value = row
  paymentDate.value = localDateInputValue(new Date())
  nextTick(() => confirmCloseButton.value?.focus())
}

function closePayment() {
  if (paying.value) return
  pendingPayment.value = null
  nextTick(() => paymentTrigger?.focus())
}

function onPaymentDialogKeydown(event: KeyboardEvent) {
  if (!pendingPayment.value) return
  if (event.key === 'Escape') {
    event.preventDefault()
    closePayment()
    return
  }
  if (event.key !== 'Tab' || !confirmDialog.value) return
  const focusable = Array.from(confirmDialog.value.querySelectorAll<HTMLElement>('button:not([disabled]), select:not([disabled]), input:not([disabled]), [href], [tabindex]:not([tabindex="-1"])'))
  if (!focusable.length) return
  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault()
    last?.focus()
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault()
    first?.focus()
  }
}

async function confirmPayment() {
  const row = pendingPayment.value
  if (!row || paying.value || !paymentDate.value) return
  const method = selectedMethods[row.id] || (row.plannedMethod === 'boleto' ? 'pix' : row.plannedMethod)
  purchases.error = ''
  paying.value = true
  const success = await purchases.payInstallment(row.id, method as PaymentMethod, paymentDate.value)
  paying.value = false
  if (!success) return
  feedback.value = `Parcela de ${formatCurrency(row.amountCents)} paga via ${methodLabel(method)} em ${formatDate(paymentDate.value)}.`
  closePayment()
}

if (import.meta.client) window.addEventListener('keydown', onPaymentDialogKeydown)
onBeforeUnmount(() => {
  if (import.meta.client) window.removeEventListener('keydown', onPaymentDialogKeydown)
})
</script>

<template>
  <section class="page payables-page">
    <PageHeader title="Contas a pagar" subtitle="Acompanhe parcelas de gastos e compras de estoque; o caixa só registra a saída quando você confirma o pagamento." />

    <div v-if="loadError || purchases.error" class="alert-box" role="alert">
      <span>{{ loadError || purchases.error }}</span>
      <button v-if="loadError" class="button button--secondary" type="button" :disabled="loading" @click="loadPage">Tentar novamente</button>
    </div>
    <p v-if="feedback" class="success-box" role="status"><CheckCircle2 :size="18" />{{ feedback }}</p>

    <section class="kpi-grid" aria-label="Resumo de contas a pagar">
      <article class="kpi-card"><span>Total em aberto</span><strong>{{ formatCurrency(openTotal) }}</strong><small>{{ purchases.payables.length }} parcela(s)</small></article>
      <article class="kpi-card kpi-card--danger"><span>Total atrasado</span><strong>{{ formatCurrency(overdueTotal) }}</strong><small>{{ overdueRows.length }} vencida(s)</small></article>
      <article class="kpi-card"><span>Próximos 7 dias</span><strong>{{ formatCurrency(nextSevenDaysTotal) }}</strong><small>{{ nextSevenDays.length }} compromisso(s)</small></article>
    </section>

    <section v-if="!loading && !loadError && purchases.alerts.length" class="alerts-grid" aria-label="Alertas financeiros">
      <article v-for="group in alertGroups" :key="group.kind" :class="`payable-alert--${group.kind}`" class="payable-alert">
        <AlertTriangle v-if="group.kind === 'overdue' || group.kind === 'due_today'" :size="20" />
        <CalendarClock v-else-if="group.kind === 'due_tomorrow'" :size="20" />
        <Sparkles v-else :size="20" />
        <div><strong>{{ group.label }}</strong><span>{{ group.count }} parcela(s) · {{ formatCurrency(group.totalCents) }}</span></div>
      </article>
    </section>

    <section class="filters panel" aria-label="Filtros de contas a pagar">
      <label class="search-field"><span>Origem</span><div><Search :size="17" /><input v-model="search" type="search" placeholder="Buscar gasto ou fornecedor" /></div></label>
      <label><span>Situação</span><select v-model="statusFilter"><option value="all">Todas</option><option value="overdue">Atrasadas</option><option value="due_soon">Vencendo em breve</option><option value="pending">Pendentes</option></select></label>
      <label><span>Período de vencimento</span><select v-model="periodFilter"><option value="all">Todo o período</option><option value="7">Próximos 7 dias</option><option value="30">Próximos 30 dias</option><option value="90">Próximos 90 dias</option></select></label>
      <span class="filter-result">{{ filteredPayables.length }} resultado(s)</span>
      <button v-if="hasFilters" class="button button--secondary filters__clear" type="button" @click="clearFilters"><RotateCcw :size="16" /> Limpar filtros</button>
    </section>

    <section class="panel payables-panel">
      <header><div><span class="eyebrow">Parcelas abertas</span><h2>Compromissos pendentes</h2></div><span class="count">{{ filteredPayables.length }}</span></header>
      <div v-if="loading" class="empty-state" aria-live="polite"><WalletCards :size="30" /><strong>Carregando contas a pagar...</strong><span>Aguarde enquanto buscamos seus compromissos.</span></div>
      <div v-else-if="loadError && !purchases.payables.length" class="empty-state"><AlertTriangle :size="30" /><strong>Contas indisponíveis</strong><span>Tente carregar novamente para consultar seus compromissos.</span></div>
      <div v-else-if="!purchases.payables.length" class="empty-state"><CheckCircle2 :size="30" /><strong>Nenhuma conta pendente</strong><span>As parcelas criadas em gastos e compras de estoque aparecerão aqui.</span></div>
      <div v-else-if="!filteredPayables.length" class="empty-state"><Search :size="28" /><strong>Nenhuma conta encontrada</strong><span>Ajuste os filtros para consultar outros compromissos.</span></div>

      <div v-else class="desktop-table">
        <table>
          <thead><tr><th>Origem / parcela</th><th>Vencimento</th><th>Previsto</th><th>Valor</th><th>Situação</th><th>Quitar com</th><th>Ação</th></tr></thead>
          <tbody>
            <tr v-for="row in filteredPayables" :key="row.id">
              <td><strong>{{ payableName(row) }}</strong><small>{{ payableKind(row) }} · Parcela {{ row.number }} de {{ installmentTotal(row) }}</small></td>
              <td>{{ formatDate(row.dueDate) }}</td><td>{{ methodLabel(row.plannedMethod) }}</td><td class="amount">{{ formatCurrency(row.amountCents) }}</td>
              <td><span :class="`status status--${state(row).className}`">{{ state(row).label }}</span></td>
              <td><select v-model="selectedMethods[row.id]" :aria-label="`Meio de pagamento da parcela ${row.number}`"><option value="cash">Dinheiro</option><option value="pix">PIX</option><option value="debit_card">Débito</option><option value="credit_card">Crédito</option></select></td>
              <td><button class="button button--primary pay" type="button" @click="requestPayment(row)">Registrar pagamento</button></td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="!loading && filteredPayables.length" class="mobile-cards">
        <article v-for="row in filteredPayables" :key="row.id" class="payable-card">
          <header><div><strong>{{ payableName(row) }}</strong><span>{{ payableKind(row) }} · Parcela {{ row.number }} de {{ installmentTotal(row) }}</span></div><span :class="`status status--${state(row).className}`">{{ state(row).label }}</span></header>
          <dl><div><dt>Vencimento</dt><dd>{{ formatDate(row.dueDate) }}</dd></div><div><dt>Forma prevista</dt><dd>{{ methodLabel(row.plannedMethod) }}</dd></div><div class="payable-card__total"><dt>Valor</dt><dd>{{ formatCurrency(row.amountCents) }}</dd></div></dl>
          <label><span>Quitar com</span><select v-model="selectedMethods[row.id]" :aria-label="`Meio de pagamento da parcela ${row.number}`"><option value="cash">Dinheiro</option><option value="pix">PIX</option><option value="debit_card">Débito</option><option value="credit_card">Crédito</option></select></label>
          <button class="button button--primary" type="button" @click="requestPayment(row)">Registrar pagamento</button>
        </article>
      </div>
    </section>

    <Teleport to="body">
      <div v-if="pendingPayment" class="confirm-payment" role="presentation">
        <button class="confirm-payment__backdrop" type="button" aria-label="Cancelar pagamento" @click="closePayment" />
        <section ref="confirmDialog" class="confirm-payment__dialog" role="dialog" aria-modal="true" aria-labelledby="confirm-payment-title" aria-describedby="confirm-payment-description">
          <button ref="confirmCloseButton" class="confirm-payment__close" type="button" aria-label="Fechar confirmação" @click="closePayment"><X :size="19" /></button>
          <div class="confirm-payment__icon"><WalletCards :size="24" /></div>
          <span class="eyebrow">Confirmar quitação</span>
          <h2 id="confirm-payment-title">Registrar pagamento?</h2>
          <p id="confirm-payment-description">Esta ação lançará a saída no caixa e {{ pendingPayment.expense ? 'em Gastos operacionais' : 'como compra de estoque realizada' }} na data informada.</p>
          <dl><div><dt>Origem</dt><dd>{{ payableName(pendingPayment) }}</dd></div><div><dt>Valor</dt><dd>{{ formatCurrency(pendingPayment.amountCents) }}</dd></div><div><dt>Meio</dt><dd>{{ methodLabel(selectedMethods[pendingPayment.id] ?? (pendingPayment.plannedMethod === 'boleto' ? 'pix' : pendingPayment.plannedMethod)) }}</dd></div></dl>
          <label class="confirm-payment__date"><span>Data do pagamento</span><input v-model="paymentDate" type="date" required :disabled="paying" /></label>
          <div class="confirm-payment__actions"><button class="button button--ghost" type="button" :disabled="paying" @click="closePayment">Cancelar</button><button class="button button--primary" type="button" :disabled="paying || !paymentDate" @click="confirmPayment">{{ paying ? 'Registrando...' : 'Confirmar pagamento' }}</button></div>
        </section>
      </div>
    </Teleport>
  </section>
</template>

<style scoped lang="scss">
.payables-page{min-width:0;gap:20px}.kpi-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:16px}.kpi-card{display:grid;gap:7px;padding:20px;border:1px solid var(--watt-border);border-radius:16px;background:var(--watt-surface)}.kpi-card>span{color:var(--watt-text-muted);font-size:12px;font-weight:700;letter-spacing:.06em;text-transform:uppercase}.kpi-card strong{font:700 clamp(22px,2vw,30px) 'Fira Code',monospace}.kpi-card small{color:var(--watt-text-muted)}.kpi-card--danger{border-left:4px solid var(--watt-alert)}.kpi-card--danger strong{color:var(--watt-alert)}.alerts-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:12px}.payable-alert{display:flex;gap:12px;padding:16px;border:1px solid var(--watt-border);border-left:4px solid var(--watt-data);border-radius:12px;background:var(--watt-surface)}.payable-alert>div,.payables-panel td:first-child{display:grid;gap:3px}.payable-alert span,td small{color:var(--watt-text-muted)}.payable-alert--overdue,.payable-alert--due_today{border-left-color:var(--watt-alert);color:var(--watt-alert);background:var(--watt-alert-background)}.payable-alert--due_tomorrow{border-left-color:var(--status-warning);color:var(--status-warning)}.payable-alert--early_payment{border-left-color:var(--watt-success);color:var(--watt-success)}.filters{display:grid;grid-template-columns:minmax(240px,1fr) 210px 220px auto;align-items:end;gap:14px;padding:16px 20px}.filters label{display:grid;gap:7px}.filters label>span{color:var(--watt-text-muted);font-size:11px;font-weight:700;text-transform:uppercase}.filters input,.filters select,.payables-panel select{width:100%;min-height:42px;border:1px solid var(--watt-border);border-radius:10px;padding:8px 11px;color:var(--watt-text);background:var(--watt-surface-raised)}.search-field>div{display:flex;align-items:center;gap:8px;border:1px solid var(--watt-border);border-radius:10px;padding-left:11px;background:var(--watt-surface-raised)}.search-field>div:focus-within{border-color:var(--watt-data)}.search-field input{border:0;background:transparent}.filter-result{align-self:center;color:var(--watt-text-muted);font-size:12px;white-space:nowrap}.payables-panel{min-width:0;padding:20px}.payables-panel>header{display:flex;align-items:center;justify-content:space-between;margin-bottom:16px}.payables-panel h2{margin:3px 0 0}.eyebrow{color:var(--watt-text-muted);font-size:11px;font-weight:700;letter-spacing:.06em;text-transform:uppercase}.count{display:grid;min-width:34px;height:34px;place-items:center;border-radius:9px;background:var(--watt-surface-raised)}.desktop-table{overflow-x:auto}table{width:100%;border-collapse:collapse}th{padding:12px 16px;color:var(--watt-text-muted);font-size:11px;letter-spacing:.08em;text-align:left;text-transform:uppercase}td{padding:14px 16px;border-top:1px solid var(--watt-border)}tbody tr:hover{background:var(--watt-surface-raised)}.amount{font:700 14px 'Fira Code',monospace;white-space:nowrap}.status{display:inline-flex;padding:6px 9px;border-radius:8px;background:var(--watt-surface-raised);font-size:12px;white-space:nowrap}.status--overdue,.status--today{color:var(--watt-alert);background:var(--watt-alert-background)}.status--tomorrow{color:var(--status-warning)}.status--early{color:var(--watt-success)}.pay{min-width:155px;padding-inline:12px}.empty-state{display:grid;min-height:220px;place-items:center;align-content:center;gap:8px;color:var(--watt-text-muted);text-align:center}.alert-box,.success-box{display:flex;align-items:center;gap:9px;margin:0;padding:14px 16px;border-left:4px solid var(--watt-alert);color:var(--watt-alert);background:var(--watt-alert-background)}.success-box{border-left-color:var(--watt-success);color:var(--watt-success);background:color-mix(in srgb,var(--watt-success) 10%,var(--watt-surface))}.mobile-cards{display:none}.confirm-payment{position:fixed;inset:0;z-index:130;display:grid;place-items:center;padding:20px}.confirm-payment__backdrop{position:absolute;inset:0;border:0;background:rgba(0,0,0,.68);backdrop-filter:blur(2px)}.confirm-payment__dialog{position:relative;display:grid;width:min(480px,100%);gap:10px;padding:24px;border:1px solid var(--watt-border);border-radius:16px;background:var(--watt-surface);box-shadow:var(--shadow-default)}.confirm-payment__close{position:absolute;top:14px;right:14px;display:grid;width:38px;height:38px;place-items:center;border:1px solid var(--watt-border);border-radius:10px;color:var(--watt-text);background:var(--watt-surface-raised)}.confirm-payment__icon{display:grid;width:48px;height:48px;place-items:center;border-radius:12px;color:var(--watt-data);background:var(--braip-theme-surface-muted)}.confirm-payment h2{margin:0}.confirm-payment p{margin:0 0 6px;color:var(--watt-text-muted);line-height:1.5}.confirm-payment dl{display:grid;grid-template-columns:repeat(3,1fr);gap:12px;margin:4px 0;padding:14px;border:1px solid var(--watt-border);border-radius:12px;background:var(--watt-surface-raised)}.confirm-payment dl div{display:grid;gap:3px}.confirm-payment dt{color:var(--watt-text-muted);font-size:10px;text-transform:uppercase}.confirm-payment dd{margin:0;font-weight:700}.confirm-payment__actions{display:flex;justify-content:flex-end;gap:10px;margin-top:6px}.confirm-payment__actions .button{min-width:150px}@media(max-width:1100px){.filters{grid-template-columns:1fr 1fr}.filter-result{display:none}.desktop-table{display:none}.mobile-cards{display:grid;gap:12px}.payable-card{display:grid;gap:16px;padding:16px;border:1px solid var(--watt-border);border-radius:14px;background:var(--watt-surface-raised)}.payable-card>header{display:flex;align-items:flex-start;justify-content:space-between;gap:12px}.payable-card>header div{display:grid;gap:4px}.payable-card>header div span{color:var(--watt-text-muted);font-size:12px}.payable-card dl{display:grid;grid-template-columns:1fr 1fr;gap:14px;margin:0}.payable-card dl div{display:grid;gap:4px}.payable-card dt,.payable-card label>span{color:var(--watt-text-muted);font-size:10px;font-weight:700;text-transform:uppercase}.payable-card dd{margin:0}.payable-card__total{grid-column:1/-1;padding-top:12px;border-top:1px solid var(--watt-border)}.payable-card__total dd{font:700 20px 'Fira Code',monospace}.payable-card label{display:grid;gap:6px}.payable-card>.button{width:100%}}@media(max-width:700px){.kpi-grid{grid-template-columns:1fr}.filters{grid-template-columns:1fr}.payables-panel{padding:16px}.alerts-grid{grid-template-columns:1fr}.confirm-payment__dialog{padding:20px}.confirm-payment dl{grid-template-columns:1fr}.confirm-payment__actions{display:grid}.confirm-payment__actions .button{width:100%}}@media(max-width:420px){.payable-card>header{display:grid}.status{justify-self:start}}
.filters__clear { grid-column: 1 / -1; justify-self: start; }
.alert-box { justify-content: space-between; }
.desktop-table table tbody td:first-child { display: table-cell !important; }
.desktop-table td:first-child > strong,
.desktop-table td:first-child > small { display: block; }
.desktop-table td:first-child > small { margin-top: 3px; }
@media (max-width: 700px) { .alert-box { align-items: stretch; flex-direction: column; } .filters__clear { width: 100%; justify-content: center; } }

/* Teleported dialog must remain usable on short mobile and zoomed viewports. */
.confirm-payment {
  padding: max(16px, env(safe-area-inset-top)) max(16px, env(safe-area-inset-right)) max(16px, env(safe-area-inset-bottom)) max(16px, env(safe-area-inset-left));
}
.confirm-payment__backdrop { width: 100%; height: 100%; }
.confirm-payment__dialog {
  max-height: calc(100dvh - 32px);
  overflow-y: auto;
  overscroll-behavior: contain;
}
.confirm-payment dd { min-width: 0; overflow-wrap: anywhere; }
.confirm-payment__date { display: grid; gap: 7px; }
.confirm-payment__date span { color: var(--watt-text-muted); font-size: 11px; font-weight: 700; text-transform: uppercase; }
.confirm-payment__date input { width: 100%; min-height: 42px; border: 1px solid var(--watt-border); border-radius: 10px; padding: 8px 11px; color: var(--watt-text); background: var(--watt-surface-raised); }
@media(max-width:700px) {
  .confirm-payment__dialog { max-height: calc(100dvh - 24px); }
  .payable-card { min-width: 0; }
  .payable-card > header strong { overflow-wrap: anywhere; }
}
</style>
