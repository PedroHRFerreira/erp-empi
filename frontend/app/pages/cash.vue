<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { currencyMaskToCents, maskCurrency } from '../utils/masks'
import { formatCurrency, formatDate, formatTime } from '../utils/format'
import PageHeader from '../components/molecules/PageHeader/Index.vue'
import { calculateCashSummary } from '../utils/cashSummary'

const cash = useCashStore()
await cash.load()
const openingInput = ref('')
const closingInput = ref('')
const closingNotes = ref('')
const adjustmentValue = ref('')
const adjustment = reactive({ direction: 'out' as 'in' | 'out', paymentMethod: 'cash' as 'cash' | 'pix' | 'debit_card' | 'credit_card', description: '', reason: '' })
const summary = computed(() => calculateCashSummary(
  cash.dailyEntries,
  cash.session?.openingCashCents || 0,
  cash.balances,
))
const totals = computed(() => summary.value.totals)
const dailyResultLabel = computed(() => {
  if (summary.value.dailyResultCents < 0) return 'prejuízo'
  if (summary.value.dailyResultCents > 0) return 'lucro'
  return 'sem lucro nem prejuízo'
})

function money(event: Event, target: 'opening' | 'closing' | 'adjustment') {
  const value = maskCurrency((event.target as HTMLInputElement).value)
  if (target === 'opening') openingInput.value = value
  else if (target === 'closing') closingInput.value = value
  else adjustmentValue.value = value
}
async function open() { await cash.open(currencyMaskToCents(openingInput.value)) }
async function close() {
  if (!closingInput.value.trim()) { cash.error = 'Informe o dinheiro contado.'; return }
  await cash.close(currencyMaskToCents(closingInput.value), closingNotes.value)
}
async function saveAdjustment() {
  await cash.addAdjustment({ ...adjustment, amountCents: currencyMaskToCents(adjustmentValue.value) })
  adjustmentValue.value = ''
}
</script>

<template>
  <section class="page cash-page">
    <PageHeader title="Caixa" subtitle="Movimento financeiro diário e conferência separada do dinheiro físico." />
    <p v-if="cash.error" class="cash-alert" role="alert">{{ cash.error }}</p>

    <section class="panel cash-summary">
      <header class="cash-section__header"><div><span class="eyebrow">Movimento financeiro</span><h2>Resumo do dia</h2></div></header>
      <div class="cash-results">
        <article class="cash-result"><span>Saldo total disponível</span><strong>{{ formatCurrency(summary.availableBalanceCents) }}</strong><small>Dinheiro esperado na gaveta + saldos acumulados de PIX e cartões.</small></article>
        <article class="cash-result"><span>Resultado do dia</span><strong :class="summary.dailyResultCents < 0 ? 'out' : 'in'">{{ formatCurrency(summary.dailyResultCents) }}</strong><small>Entradas do dia − saídas do dia: {{ dailyResultLabel }}. O fundo inicial não entra neste resultado.</small></article>
      </div>
      <div class="cash-methods">
        <article v-for="method in ['cash','pix','debit_card','credit_card']" :key="method" class="cash-method"><span>{{ method === 'cash' ? 'Dinheiro · hoje' : method === 'pix' ? 'PIX · acumulado' : method === 'debit_card' ? 'Débito · acumulado' : 'Crédito · acumulado' }}</span><strong>{{ formatCurrency(totals[method] || 0) }}</strong></article>
      </div>
      <p class="description">Dinheiro mostra o movimento de hoje. PIX e cartões preservam o saldo acumulado, mesmo após fechar e reabrir a gaveta.</p>
    </section>

    <section v-if="!cash.session" class="panel cash-section cash-opening">
      <span class="eyebrow">Conferência física</span><h2>Abrir gaveta de dinheiro</h2><p class="description">Informe somente o fundo de troco físico. Os outros meios não dependem desta abertura.</p>
      <label class="field"><span>Dinheiro inicial</span><input :value="openingInput" inputmode="numeric" placeholder="R$ 0,00" @input="money($event, 'opening')" /></label>
      <button class="button button--primary action" type="button" @click="open">Abrir caixa físico</button>
    </section>
    <section v-else class="panel cash-section cash-physical">
      <header class="cash-section__header"><div><span class="eyebrow">Conferência física</span><h2>Dinheiro esperado na gaveta</h2></div><span class="status">Aberto</span></header>
      <strong class="expected">{{ formatCurrency(cash.session.expectedCashCents) }}</strong><p class="description">Fundo inicial: {{ formatCurrency(cash.session.openingCashCents) }}. PIX e cartões não alteram este valor.</p>
      <div class="close-grid"><label class="field"><span>Dinheiro contado</span><input :value="closingInput" inputmode="numeric" @input="money($event, 'closing')" /></label><label class="field"><span>Justificativa da diferença</span><input v-model="closingNotes" /></label><button class="button button--primary" type="button" @click="close">Fechar caixa físico</button></div>
    </section>

    <section class="panel cash-section">
      <header class="cash-section__header"><div><span class="eyebrow">Extrato</span><h2>Movimentos financeiros de hoje</h2></div><span class="count">{{ cash.dailyEntries.length }}</span></header>
      <p v-if="!cash.dailyEntries.length" class="description">Nenhum movimento registrado hoje.</p>
      <div v-else class="table-wrap" tabindex="0" aria-label="Extrato financeiro de hoje"><table><thead><tr><th>Horário e origem</th><th>Meio</th><th>Valor</th></tr></thead><tbody><tr v-for="entry in cash.dailyEntries" :key="entry.id"><td><NuxtLink v-if="entry.referenceType === 'payable_installment' && entry.referenceId" class="entry-link" :to="`/payment-history/${entry.referenceId}`" :aria-label="`Ver histórico do pagamento de ${entry.description}`"><strong>{{ entry.description }}</strong><small>{{ formatTime(entry.occurredAt) }}</small></NuxtLink><template v-else><strong>{{ entry.description }}</strong><small>{{ formatTime(entry.occurredAt) }}</small></template></td><td>{{ entry.paymentMethod === 'cash' ? 'Dinheiro' : entry.paymentMethod === 'pix' ? 'PIX' : entry.paymentMethod === 'debit_card' ? 'Débito' : 'Crédito' }}</td><td :class="entry.amountCents < 0 ? 'out' : 'in'">{{ formatCurrency(entry.amountCents) }}</td></tr></tbody></table></div>
    </section>

    <section class="panel cash-section">
      <header class="cash-section__header"><div><span class="eyebrow">Correção auditável</span><h2>Registrar ajuste</h2></div></header><p class="description">Dinheiro exige gaveta aberta; ajustes em PIX ou cartões podem ser registrados sem ela.</p>
      <div class="adjustment-grid"><label class="field"><span>Tipo</span><select v-model="adjustment.direction"><option value="in">Entrada</option><option value="out">Saída</option></select></label><label class="field"><span>Meio</span><select v-model="adjustment.paymentMethod"><option value="cash">Dinheiro</option><option value="pix">PIX</option><option value="debit_card">Débito</option><option value="credit_card">Crédito</option></select></label><label class="field"><span>Valor</span><input :value="adjustmentValue" inputmode="numeric" @input="money($event, 'adjustment')" /></label><label class="field"><span>Descrição</span><input v-model="adjustment.description" /></label><label class="field"><span>Motivo</span><input v-model="adjustment.reason" /></label><button class="button button--secondary" type="button" @click="saveAdjustment">Registrar ajuste</button></div>
    </section>

    <section class="panel cash-section"><header class="cash-section__header"><div><span class="eyebrow">Histórico</span><h2>Últimos fechamentos físicos</h2></div></header><p v-if="!cash.sessions.length" class="description">Nenhum fechamento registrado.</p><div v-else class="table-wrap" tabindex="0" aria-label="Histórico de fechamentos físicos"><table><thead><tr><th>Data</th><th>Situação</th><th>Contagem</th></tr></thead><tbody><tr v-for="day in cash.sessions" :key="day.id"><td>{{ formatDate(day.businessDate) }}</td><td>{{ day.status === 'open' ? 'Aberto' : 'Fechado' }}</td><td>{{ formatCurrency(day.status === 'closed' ? (day.closingCashCents || 0) : day.expectedCashCents) }}</td></tr></tbody></table></div></section>
  </section>
</template>

<style scoped lang="scss">
.cash-page,.cash-summary,.cash-section{display:grid;min-width:0;gap:18px}.cash-summary,.cash-section{padding:22px}.cash-section__header{display:flex;align-items:center;justify-content:space-between;gap:16px}.cash-section h2,.cash-summary h2{margin:3px 0 0}.eyebrow{color:var(--watt-text-muted);font-size:11px;font-weight:700;letter-spacing:.06em;text-transform:uppercase}.description{margin:0;color:var(--watt-text-muted)}.expected{font:700 30px 'Fira Code',monospace;color:var(--watt-data);white-space:nowrap}.cash-results{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:12px}.cash-result{display:grid;gap:6px;padding:16px;border:1px solid var(--watt-border);border-radius:12px;background:var(--watt-surface-raised)}.cash-result span{font-weight:700}.cash-result strong{font:700 26px 'Fira Code',monospace;color:var(--watt-data)}.cash-result strong.out{color:var(--watt-alert)}.cash-result strong.in{color:var(--watt-success)}.cash-result small{color:var(--watt-text-muted)}.cash-methods{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:12px}.cash-method{display:grid;gap:6px;padding:16px;border:1px solid var(--watt-border);border-radius:12px;background:var(--watt-surface-raised)}.cash-method span{color:var(--watt-text-muted);font-size:11px;text-transform:uppercase}.cash-method strong{font:700 19px 'Fira Code',monospace}.status,.count{display:grid;min-width:32px;height:32px;place-items:center;border-radius:9px;padding:0 10px;color:var(--watt-success);background:color-mix(in srgb,var(--watt-success) 14%,transparent)}.action{width:max-content}.close-grid,.adjustment-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:12px;align-items:end}.adjustment-grid{grid-template-columns:repeat(5,minmax(0,1fr)) auto}.cash-alert{margin:0;padding:14px 16px;border-left:4px solid var(--watt-alert);color:var(--watt-alert);background:var(--watt-alert-background)}td:first-child{display:grid;gap:3px}td small{color:var(--watt-text-muted)}td:last-child{font-family:'Fira Code',monospace;font-weight:700}.in{color:var(--watt-success)}.out{color:var(--watt-alert)}@media(max-width:1000px){.cash-methods{grid-template-columns:repeat(2,1fr)}.adjustment-grid{grid-template-columns:repeat(2,1fr)}.close-grid{grid-template-columns:1fr 1fr}}@media(max-width:640px){.cash-results,.cash-methods,.adjustment-grid,.close-grid{grid-template-columns:1fr}.cash-section__header{display:grid;grid-template-columns:minmax(0,1fr) auto;align-items:start}.cash-result strong{font-size:20px}}
.cash-section__header > * { min-width: 0; }
.cash-method { min-width: 0; }
.cash-method strong { overflow: hidden; text-overflow: ellipsis; }
.table-wrap:focus-visible { outline-offset: 2px; }
.table-wrap table { min-width: 560px; }
.table-wrap table tbody td:first-child { display: table-cell !important; }
.table-wrap td:first-child > strong,
.table-wrap td:first-child > small { display: block; }
.table-wrap td:first-child > small { margin-top: 3px; }
.entry-link { display: block; width: fit-content; color: inherit; text-decoration: none; }
.entry-link > strong, .entry-link > small { display: block; }
.entry-link > small { margin-top: 3px; }
.entry-link:hover > strong { text-decoration: underline; }
.entry-link:focus-visible { border-radius: 3px; outline: 2px solid var(--watt-focus, currentColor); outline-offset: 3px; }
@media(max-width:640px){
  .cash-summary,.cash-section{padding:16px}
  .cash-section__header{grid-template-columns:minmax(0,1fr)}
  .expected{max-width:100%;overflow-wrap:anywhere;white-space:normal}
  .status,.count{justify-self:start}
}
</style>
