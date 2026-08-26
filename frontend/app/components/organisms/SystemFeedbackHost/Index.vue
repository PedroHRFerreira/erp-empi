<script setup lang="ts">
import { AlertCircle, AlertTriangle, CheckCircle2, Info, X } from '@lucide/vue'
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'

const feedback = useSystemFeedback()
const dialog = ref<HTMLElement | null>(null)
const cancelButton = ref<HTMLButtonElement | null>(null)
const timers = new Map<string, { remaining: number; started: number; timer: ReturnType<typeof setTimeout> }>()
let trigger: HTMLElement | null = null

function icon(tone: string) {
  return tone === 'success' ? CheckCircle2 : tone === 'danger' ? AlertCircle : tone === 'warning' ? AlertTriangle : Info
}

function startToast(id: string, duration: number) {
  clearToastTimer(id)
  timers.set(id, { remaining: duration, started: Date.now(), timer: setTimeout(() => feedback.dismiss(id), duration) })
}

function pauseToast(id: string) {
  const state = timers.get(id)
  if (!state) return
  clearTimeout(state.timer)
  state.remaining = Math.max(0, state.remaining - (Date.now() - state.started))
}

function resumeToast(id: string) {
  const state = timers.get(id)
  if (!state) return
  state.started = Date.now()
  state.timer = setTimeout(() => feedback.dismiss(id), state.remaining)
}

function clearToastTimer(id: string) {
  const state = timers.get(id)
  if (state) clearTimeout(state.timer)
  timers.delete(id)
}

function answer(value: boolean) {
  feedback.answer(value)
}

function onKeydown(event: KeyboardEvent) {
  if (!feedback.confirmation.value) return
  if (event.key === 'Escape') { event.preventDefault(); answer(false); return }
  if (event.key !== 'Tab' || !dialog.value) return
  const focusable = Array.from(dialog.value.querySelectorAll<HTMLElement>('button:not([disabled]), [href], [tabindex]:not([tabindex="-1"])'))
  const first = focusable[0]
  const last = focusable.at(-1)
  if (!first || !last) return
  if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last.focus() }
  else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first.focus() }
}

watch(() => feedback.toasts.value.map(item => item.id), (ids) => {
  if (!import.meta.client) return
  for (const item of feedback.toasts.value) if (!timers.has(item.id)) startToast(item.id, item.duration)
  for (const id of timers.keys()) if (!ids.includes(id)) clearToastTimer(id)
}, { deep: true, immediate: true })

watch(() => feedback.confirmation.value, async (current, previous) => {
  if (!import.meta.client) return
  if (current && !previous) {
    trigger = document.activeElement instanceof HTMLElement ? document.activeElement : null
    window.addEventListener('keydown', onKeydown)
    await nextTick()
    cancelButton.value?.focus()
  } else if (!current && previous) {
    window.removeEventListener('keydown', onKeydown)
    await nextTick()
    trigger?.focus()
    trigger = null
  }
})

onBeforeUnmount(() => {
  if (import.meta.client) window.removeEventListener('keydown', onKeydown)
  for (const id of timers.keys()) clearToastTimer(id)
})
</script>

<template>
  <Teleport to="body">
    <div class="system-toasts" aria-live="polite" aria-atomic="false">
      <article v-for="toast in feedback.toasts.value" :key="toast.id" class="system-toast" :class="`system-toast--${toast.tone}`" role="status" @mouseenter="pauseToast(toast.id)" @mouseleave="resumeToast(toast.id)">
        <component :is="icon(toast.tone)" :size="20" aria-hidden="true" />
        <div><strong>{{ toast.title }}</strong><p v-if="toast.message">{{ toast.message }}</p></div>
        <button type="button" :aria-label="`Fechar aviso: ${toast.title}`" @click="feedback.dismiss(toast.id)"><X :size="17" /></button>
        <span class="system-toast__timer" :style="{ animationDuration: `${toast.duration}ms` }" />
      </article>
    </div>

    <div v-if="feedback.confirmation.value" class="system-dialog" role="presentation" @mousedown.self="answer(false)">
      <section ref="dialog" class="system-dialog__panel" role="alertdialog" aria-modal="true" aria-labelledby="system-dialog-title" aria-describedby="system-dialog-message" tabindex="-1">
        <div class="system-dialog__icon" :class="`system-dialog__icon--${feedback.confirmation.value.tone}`"><component :is="icon(feedback.confirmation.value.tone)" :size="25" /></div>
        <span class="system-dialog__eyebrow">{{ feedback.confirmation.value.tone === 'danger' ? 'Ação de risco' : 'Confirmação' }}</span>
        <h2 id="system-dialog-title">{{ feedback.confirmation.value.title }}</h2>
        <p id="system-dialog-message">{{ feedback.confirmation.value.message }}</p>
        <dl v-if="feedback.confirmation.value.details?.length" class="system-dialog__details"><div v-for="detail in feedback.confirmation.value.details" :key="detail.label"><dt>{{ detail.label }}</dt><dd>{{ detail.value }}</dd></div></dl>
        <label v-if="feedback.confirmation.value.input" class="system-dialog__input"><span>{{ feedback.confirmation.value.input.label }}</span><input v-model="feedback.confirmation.value.input.value" :type="feedback.confirmation.value.input.type" /></label>
        <p v-if="feedback.confirmation.value.warning" class="system-dialog__warning"><strong>{{ feedback.confirmation.value.warning }}</strong></p>
        <div class="system-dialog__actions">
          <button ref="cancelButton" class="button button--ghost" type="button" @click="answer(false)">{{ feedback.confirmation.value.cancelLabel }}</button>
          <button class="button system-dialog__confirm" :class="`system-dialog__confirm--${feedback.confirmation.value.tone}`" type="button" @click="answer(true)">{{ feedback.confirmation.value.confirmLabel }}</button>
        </div>
      </section>
    </div>
  </Teleport>
</template>

<style scoped lang="scss">
.system-toasts{position:fixed;z-index:180;top:max(18px,env(safe-area-inset-top));right:max(18px,env(safe-area-inset-right));display:grid;width:min(390px,calc(100vw - 32px));gap:10px;pointer-events:none}.system-toast{--tone:var(--watt-data);position:relative;display:grid;grid-template-columns:auto minmax(0,1fr) auto;gap:11px;overflow:hidden;padding:15px;border:1px solid color-mix(in srgb,var(--tone) 38%,var(--watt-border));border-radius:13px;color:var(--watt-text);background:color-mix(in srgb,var(--tone) 8%,var(--watt-surface));box-shadow:var(--shadow-default);pointer-events:auto}.system-toast--success{--tone:var(--watt-success)}.system-toast--warning,.system-toast--danger{--tone:var(--watt-alert)}.system-toast>svg{color:var(--tone)}.system-toast div{display:grid;gap:3px}.system-toast p{margin:0;color:var(--watt-text-muted);line-height:1.4}.system-toast button{display:grid;width:30px;height:30px;place-items:center;border:0;border-radius:8px;color:var(--watt-text-muted);background:transparent}.system-toast button:hover{background:var(--watt-surface-raised)}.system-toast__timer{position:absolute;right:0;bottom:0;left:0;height:3px;transform-origin:left;background:var(--tone);animation:toast-countdown linear forwards}@keyframes toast-countdown{to{transform:scaleX(0)}}.system-toast:hover .system-toast__timer{animation-play-state:paused}.system-dialog{position:fixed;inset:0;z-index:170;display:grid;place-items:center;padding:16px;background:rgba(0,0,0,.68);backdrop-filter:blur(2px)}.system-dialog__panel{display:grid;width:min(500px,100%);max-height:calc(100dvh - 32px);gap:11px;overflow-y:auto;padding:24px;border:1px solid var(--watt-border);border-radius:16px;color:var(--watt-text);background:var(--watt-surface);box-shadow:var(--shadow-default)}.system-dialog__icon{display:grid;width:48px;height:48px;place-items:center;border-radius:12px;color:var(--watt-data);background:var(--watt-data-background)}.system-dialog__icon--success{color:var(--watt-success);background:var(--watt-success-background)}.system-dialog__icon--danger,.system-dialog__icon--warning{color:var(--watt-alert);background:var(--watt-alert-background)}.system-dialog__eyebrow{color:var(--watt-text-muted);font-size:11px;font-weight:750;letter-spacing:.07em;text-transform:uppercase}.system-dialog h2,.system-dialog p{margin:0}.system-dialog p{color:var(--watt-text-muted);line-height:1.5}.system-dialog__warning{padding:13px;border:1px solid color-mix(in srgb,var(--watt-alert) 42%,var(--watt-border));border-radius:11px;color:var(--watt-alert)!important;background:var(--watt-alert-background)}.system-dialog__actions{display:flex;justify-content:flex-end;gap:10px;margin-top:8px}.system-dialog__confirm{border:1px solid var(--watt-data);color:#fff;background:var(--watt-data)}.system-dialog__confirm--danger,.system-dialog__confirm--warning{border-color:var(--watt-alert);background:var(--watt-alert)}@media(max-width:600px){.system-toasts{top:auto;right:16px;bottom:max(16px,env(safe-area-inset-bottom));left:16px;width:auto}.system-dialog__panel{padding:20px}.system-dialog__actions{display:grid}.system-dialog__actions .button{width:100%}}
.system-dialog__details{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:10px;margin:3px 0}.system-dialog__details div,.system-dialog__input{display:grid;gap:5px;padding:12px;border:1px solid var(--watt-border);border-radius:10px;background:var(--watt-surface-raised)}.system-dialog__details dt,.system-dialog__input span{color:var(--watt-text-muted);font-size:10px;font-weight:750;text-transform:uppercase}.system-dialog__details dd{margin:0;font-weight:750;overflow-wrap:anywhere}.system-dialog__input input{width:100%}
</style>
