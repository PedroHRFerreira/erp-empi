<script setup lang="ts">
import { AlertTriangle, CalendarClock, CheckCircle2, Sparkles, X } from '@lucide/vue'
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'
import type { IPayableAlert } from '../../../../server/contracts/types'
import { formatCurrency, formatDate } from '../../../utils/format'

const props = defineProps<{
  open: boolean
  alerts: IPayableAlert[]
  loading?: boolean
}>()

const emit = defineEmits<{ close: [] }>()
const closeButton = ref<HTMLButtonElement | null>(null)
let returnFocus: HTMLElement | null = null
let previousOverflow = ''

function label(kind: IPayableAlert['kind']) {
  return ({
    overdue: 'Pagamento atrasado',
    due_today: 'Vence hoje',
    due_tomorrow: 'Vence amanhã',
    early_payment: 'Pode antecipar'
  })[kind]
}

function plannedMethod(method: IPayableAlert['plannedMethod']) {
  return ({ boleto: 'Boleto', cash: 'Dinheiro', pix: 'PIX', debit_card: 'Débito', credit_card: 'Crédito' })[method]
}

function onKeydown(event: KeyboardEvent) {
  if (!props.open) return
  if (event.key === 'Escape') {
    event.preventDefault()
    emit('close')
    return
  }
  if (event.key !== 'Tab') return

  const panel = closeButton.value?.closest<HTMLElement>('.financial-drawer__panel')
  const focusable = panel
    ? Array.from(panel.querySelectorAll<HTMLElement>('button:not([disabled]), a[href], input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])'))
    : []
  if (!focusable.length) return

  const first = focusable[0]!
  const last = focusable[focusable.length - 1]!
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault()
    last.focus()
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault()
    first.focus()
  }
}

watch(() => props.open, async (open) => {
  if (open) {
    returnFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null
    previousOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    await nextTick()
    closeButton.value?.focus()
  } else {
    document.body.style.overflow = previousOverflow
    returnFocus?.focus()
    returnFocus = null
  }
})

if (import.meta.client) window.addEventListener('keydown', onKeydown)
onBeforeUnmount(() => {
  if (!import.meta.client) return
  window.removeEventListener('keydown', onKeydown)
  document.body.style.overflow = previousOverflow
})
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="financial-drawer" role="presentation">
      <button class="financial-drawer__backdrop" type="button" aria-label="Fechar notificações" @click="emit('close')" />
      <aside id="financial-notifications" class="financial-drawer__panel" role="dialog" aria-modal="true" aria-labelledby="financial-drawer-title" aria-describedby="financial-drawer-description">
        <header class="financial-drawer__header">
          <div>
            <span class="financial-drawer__eyebrow">Central financeira</span>
            <h2 id="financial-drawer-title">Notificações</h2>
            <p id="financial-drawer-description">Acompanhe vencimentos e oportunidades de antecipação.</p>
          </div>
          <button ref="closeButton" class="financial-drawer__close" type="button" aria-label="Fechar notificações" @click="emit('close')">
            <X :size="20" />
          </button>
        </header>

        <div class="financial-drawer__summary">
          <strong>{{ alerts.length }}</strong>
          <span>{{ alerts.length === 1 ? 'compromisso requer atenção' : 'compromissos requerem atenção' }}</span>
        </div>

        <div class="financial-drawer__content">
          <p v-if="loading" class="financial-drawer__empty" role="status" aria-live="polite">Atualizando notificações...</p>
          <div v-else-if="!alerts.length" class="financial-drawer__empty">
            <CheckCircle2 :size="32" />
            <strong>Tudo em dia</strong>
            <span>Nenhum vencimento urgente ou pagamento antecipável.</span>
          </div>
          <article v-for="alert in alerts" v-else :key="alert.installmentId" :class="`financial-alert--${alert.kind}`" class="financial-alert">
            <div class="financial-alert__icon" aria-hidden="true">
              <AlertTriangle v-if="alert.kind === 'overdue' || alert.kind === 'due_today'" :size="18" />
              <CalendarClock v-else-if="alert.kind === 'due_tomorrow'" :size="18" />
              <Sparkles v-else :size="18" />
            </div>
            <div class="financial-alert__body">
              <span class="financial-alert__kind">{{ label(alert.kind) }}</span>
              <strong>{{ alert.supplierName }}</strong>
              <dl>
                <div><dt>Parcela</dt><dd>{{ alert.number }}</dd></div>
                <div><dt>Vencimento</dt><dd>{{ formatDate(alert.dueDate) }}</dd></div>
                <div><dt>Previsto</dt><dd>{{ plannedMethod(alert.plannedMethod) }}</dd></div>
              </dl>
              <span class="financial-alert__amount">{{ formatCurrency(alert.amountCents) }}</span>
            </div>
          </article>
        </div>

        <footer class="financial-drawer__footer">
          <NuxtLink class="button button--primary" to="/payables" @click="emit('close')">Ver contas a pagar</NuxtLink>
        </footer>
      </aside>
    </div>
  </Teleport>
</template>

<style scoped lang="scss">
.financial-drawer{position:fixed;inset:0;z-index:120}.financial-drawer__backdrop{position:absolute;inset:0;width:100%;border:0;background:rgba(0,0,0,.64);backdrop-filter:blur(2px)}.financial-drawer__panel{position:absolute;top:0;right:0;display:flex;width:min(440px,100vw);height:100%;flex-direction:column;border-left:1px solid var(--watt-border);background:var(--watt-background);box-shadow:-24px 0 64px rgba(0,0,0,.42);animation:drawer-in 180ms ease-out}.financial-drawer__header{display:flex;align-items:flex-start;justify-content:space-between;gap:20px;padding:24px;border-bottom:1px solid var(--watt-border)}.financial-drawer__eyebrow{color:var(--watt-data);font-size:11px;font-weight:700;letter-spacing:.1em;text-transform:uppercase}.financial-drawer__header h2{margin:5px 0 4px;font-size:24px}.financial-drawer__header p{margin:0;color:var(--watt-text-muted);font-size:13px;line-height:1.5}.financial-drawer__close{display:grid;min-width:42px;height:42px;place-items:center;border:1px solid var(--watt-border);border-radius:12px;color:var(--watt-text);background:var(--watt-surface)}.financial-drawer__close:hover{border-color:var(--watt-data);color:var(--watt-data)}.financial-drawer__summary{display:flex;align-items:baseline;gap:9px;margin:20px 24px 4px;padding:14px 16px;border:1px solid var(--watt-border);border-radius:12px;background:var(--watt-surface)}.financial-drawer__summary strong{color:var(--watt-data);font:700 24px 'Fira Code',monospace}.financial-drawer__summary span{color:var(--watt-text-muted);font-size:13px}.financial-drawer__content{display:grid;align-content:start;flex:1;gap:12px;overflow-y:auto;padding:16px 24px 24px}.financial-alert{display:flex;gap:12px;padding:16px;border:1px solid var(--watt-border);border-left:4px solid var(--watt-data);border-radius:14px;background:var(--watt-surface)}.financial-alert__icon{display:grid;width:34px;height:34px;flex:0 0 34px;place-items:center;border-radius:9px;color:var(--watt-data);background:var(--watt-surface-raised)}.financial-alert__body{display:grid;min-width:0;flex:1;gap:5px}.financial-alert__kind{color:var(--watt-data);font-size:11px;font-weight:800;letter-spacing:.06em;text-transform:uppercase}.financial-alert__body>strong{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.financial-alert dl{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:8px;margin:7px 0}.financial-alert dl div{display:grid;gap:2px}.financial-alert dt{color:var(--watt-text-muted);font-size:10px;text-transform:uppercase}.financial-alert dd{margin:0;font-size:12px}.financial-alert__amount{font:700 18px 'Fira Code',monospace}.financial-alert--overdue,.financial-alert--due_today{border-left-color:var(--watt-alert);background:var(--watt-alert-background)}.financial-alert--overdue .financial-alert__icon,.financial-alert--due_today .financial-alert__icon,.financial-alert--overdue .financial-alert__kind,.financial-alert--due_today .financial-alert__kind{color:var(--watt-alert)}.financial-alert--due_tomorrow{border-left-color:var(--status-warning)}.financial-alert--due_tomorrow .financial-alert__icon,.financial-alert--due_tomorrow .financial-alert__kind{color:var(--status-warning)}.financial-alert--early_payment{border-left-color:var(--watt-success)}.financial-alert--early_payment .financial-alert__icon,.financial-alert--early_payment .financial-alert__kind{color:var(--watt-success)}.financial-drawer__empty{display:grid;min-height:220px;place-items:center;align-content:center;gap:8px;margin:0;color:var(--watt-text-muted);text-align:center}.financial-drawer__empty svg{color:var(--watt-success)}.financial-drawer__footer{padding:16px 24px 24px;border-top:1px solid var(--watt-border)}.financial-drawer__footer .button{width:100%}@keyframes drawer-in{from{transform:translateX(100%)}to{transform:translateX(0)}}@media(max-width:520px){.financial-drawer__panel{width:100%}.financial-drawer__header{padding:20px 16px}.financial-drawer__summary{margin:16px 16px 0}.financial-drawer__content{padding:16px}.financial-drawer__footer{padding:14px 16px 20px}.financial-alert dl{grid-template-columns:1fr 1fr}.financial-alert dl div:last-child{grid-column:1/-1}}@media(prefers-reduced-motion:reduce){.financial-drawer__panel{animation:none}}
.financial-drawer__backdrop { height: 100%; }
.financial-drawer__panel { min-width: 0; height: 100dvh; }
.financial-drawer__header > div { min-width: 0; }
.financial-drawer__content { min-height: 0; overscroll-behavior: contain; }
.financial-alert dd { min-width: 0; overflow-wrap: anywhere; }
.financial-drawer__footer { padding-bottom: max(24px, env(safe-area-inset-bottom)); }
@media(max-width:520px){
  .financial-drawer__header{padding-top:max(20px,env(safe-area-inset-top));padding-right:max(16px,env(safe-area-inset-right));padding-left:max(16px,env(safe-area-inset-left))}
  .financial-drawer__content,.financial-drawer__footer{padding-right:max(16px,env(safe-area-inset-right));padding-left:max(16px,env(safe-area-inset-left))}
  .financial-alert__body>strong{white-space:normal;overflow-wrap:anywhere}
}
</style>
