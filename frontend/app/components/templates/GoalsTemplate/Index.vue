<script lang="ts">
import { computed, defineComponent, reactive, ref, watch } from 'vue'
import type { IGoalsSummary, IMonthlyGoal } from '../../../../server/contracts/types'
import { formatCurrency } from '../../../utils/format'
import { currencyMaskToCents, formatCentsAsCurrency, maskCurrency } from '../../../utils/masks'
import PageHeader from '../../molecules/PageHeader/Index.vue'
import GoalsProgressChart from '../../organisms/GoalsProgressChart/Index.vue'

type GoalKey = 'revenue' | 'labor' | 'products' | 'clients' | 'netProfit'

interface GoalMetric {
  key: GoalKey
  label: string
  shortLabel: string
  previous: number
  target: number
  actual: number
  projection: number
  currency: boolean
}

function emptyTargets(): IMonthlyGoal {
  return {
    month: '', revenueTargetCents: 0, laborTargetCents: 0, productsTargetCents: 0,
    clientsTarget: 0, netProfitTargetCents: 0
  }
}

export default defineComponent({
  name: 'GoalsTemplate',
  components: { GoalsProgressChart, PageHeader },
  setup() {
    const goals = useGoalsStore()
    const form = reactive<IMonthlyGoal>(emptyTargets())
    const saveState = ref<'idle' | 'saving' | 'saved' | 'error'>('idle')
    const periodPreset = ref<'previous-month' | 'current-month' | 'current-week' | 'custom'>('previous-month')
    const customStartDate = ref('')
    const customEndDate = ref('')
    let saveTimer: ReturnType<typeof setTimeout> | undefined

    function syncForm(summary: IGoalsSummary | null) {
      if (summary) Object.assign(form, summary.targets)
    }

    watch(() => goals.summary, syncForm, { immediate: true })

    const metrics = computed<GoalMetric[]>(() => {
      const summary = goals.summary
      if (!summary) return []
      return [
        metric('revenue', 'Faturamento', 'Faturamento', summary.previous.revenueCents, summary.targets.revenueTargetCents, summary.actual.revenueCents, summary.projection.revenueCents, true),
        metric('labor', 'Mão de obra', 'Mão de obra', summary.previous.laborCents, summary.targets.laborTargetCents, summary.actual.laborCents, summary.projection.laborCents, true),
        metric('products', 'Produtos', 'Produtos', summary.previous.productsCents, summary.targets.productsTargetCents, summary.actual.productsCents, summary.projection.productsCents, true),
        metric('clients', 'Clientes', 'Clientes', summary.previous.clients, summary.targets.clientsTarget, summary.actual.clients, summary.projection.clients, false),
        metric('netProfit', 'Lucro líquido', 'Lucro', summary.previous.netProfitCents, summary.targets.netProfitTargetCents, summary.actual.netProfitCents, summary.projection.netProfitCents, true)
      ]
    })

    const monthLabel = computed(() => {
      const month = goals.summary?.month
      if (!month) return ''
      return new Intl.DateTimeFormat('pt-BR', { month: 'long', year: 'numeric' }).format(new Date(`${month}-02T12:00:00`))
    })
    const periodLabel = computed(() => {
      const summary = goals.summary
      if (!summary) return ''
      const format = (value: string) => new Intl.DateTimeFormat('pt-BR').format(new Date(`${value}T12:00:00`))
      return summary.periodStart === summary.periodEnd ? format(summary.periodStart) : `${format(summary.periodStart)} a ${format(summary.periodEnd)}`
    })

    const required = computed(() => {
      const summary = goals.summary
      if (!summary) return []
      return [
        { label: 'Ticket médio necessário', value: formatCurrency(summary.requirements.averageTicketCents) },
        { label: 'Mão de obra por atendimento', value: formatCurrency(summary.requirements.laborPerAppointmentCents) },
        { label: 'Produtos por atendimento', value: formatCurrency(summary.requirements.productsPerAppointmentCents) }
      ]
    })

    function metric(key: GoalKey, label: string, shortLabel: string, previous: number, target: number, actual: number, projection: number, currency: boolean): GoalMetric {
      return { key, label, shortLabel, previous, target, actual, projection, currency }
    }

    function value(metric: GoalMetric, number: number) {
      return metric.currency ? formatCurrency(number) : new Intl.NumberFormat('pt-BR').format(number)
    }

    function progress(metric: GoalMetric) {
      if (!metric.target) return 0
      return Math.round((metric.actual / metric.target) * 100)
    }

    function remaining(metric: GoalMetric) {
      return Math.max(metric.target - metric.actual, 0)
    }

    function barWidth(metric: GoalMetric, value: number) {
      const maximum = Math.max(metric.previous, metric.target, metric.actual, 1)
      return `${Math.max((value / maximum) * 100, value ? 4 : 0)}%`
    }

    function currencyInput(value: number) { return formatCentsAsCurrency(value) }
    function updateCents(field: keyof IMonthlyGoal, event: Event) {
      const input = event.target as HTMLInputElement
      input.value = maskCurrency(input.value)
      form[field] = currencyMaskToCents(input.value) as never
      scheduleSave()
    }

    async function save() {
      saveState.value = 'saving'
      const saved = await goals.save({ ...form })
      saveState.value = saved ? 'saved' : 'error'
    }

    function scheduleSave() {
      if (saveTimer) clearTimeout(saveTimer)
      saveState.value = 'idle'
      saveTimer = setTimeout(() => { void save() }, 650)
    }

    function applyPeriod() {
      const now = new Date()
      if (periodPreset.value === 'previous-month') {
        const previous = new Date(now.getFullYear(), now.getMonth() - 1, 1)
        void goals.load(monthValue(previous), '', '', true); return
      }
      if (periodPreset.value === 'current-month') { void goals.load(monthValue(now), '', '', true); return }
      if (periodPreset.value === 'current-week') {
        const start = new Date(now)
        start.setDate(now.getDate() - ((now.getDay() + 6) % 7))
        const end = new Date(start); end.setDate(start.getDate() + 6)
        void goals.load(monthValue(now), dateValue(start), dateValue(end), true); return
      }
      if (customStartDate.value && customEndDate.value) void goals.load(customStartDate.value.slice(0, 7), customStartDate.value, customEndDate.value, true)
    }

    return { applyPeriod, barWidth, currencyInput, customEndDate, customStartDate, formatCurrency, form, goals, metrics, monthLabel, periodLabel, periodPreset, progress, remaining, required, saveState, scheduleSave, updateCents, value }
  }
})
</script>

function dateValue(value: Date) { return `${value.getFullYear()}-${String(value.getMonth() + 1).padStart(2, '0')}-${String(value.getDate()).padStart(2, '0')}` }
function monthValue(value: Date) { return dateValue(value).slice(0, 7) }

<template>
  <section class="page goals-template">
    <PageHeader title="Metas" subtitle="Defina o resultado desejado, registre os recebimentos e acompanhe o quanto falta para chegar lá." />

    <div v-if="goals.loading && !goals.summary" class="panel goals-template__empty">Carregando metas...</div>
    <div v-else-if="goals.error && !goals.summary" class="panel goals-template__empty goals-template__empty--error">{{ goals.error }}</div>

    <template v-else-if="goals.summary">
      <form class="goals-template__filters panel" @submit.prevent="applyPeriod">
        <label class="field"><span>Período analisado</span><select v-model="periodPreset" @change="applyPeriod"><option value="previous-month">Mês anterior</option><option value="current-month">Mês atual</option><option value="current-week">Semana atual</option><option value="custom">Intervalo personalizado</option></select></label>
        <template v-if="periodPreset === 'custom'"><label class="field"><span>Início</span><input v-model="customStartDate" type="date" /></label><label class="field"><span>Fim</span><input v-model="customEndDate" type="date" /></label><button class="button button--secondary" type="submit">Aplicar</button></template>
        <p>Indicadores, gráfico e dicas consideram: <strong>{{ periodLabel }}</strong>.</p>
      </form>
      <section class="goals-template__intro panel">
        <div>
          <span class="goals-template__eyebrow">Planejamento de {{ monthLabel }}</span>
          <h2>{{ goals.summary.saved ? 'Sua meta está definida' : 'Sugestão pronta para você' }}</h2>
          <p>1. Ajuste os valores abaixo. 2. As mudanças são salvas automaticamente. 3. Conforme recibos forem pagos, o acompanhamento mostra o realizado, o que falta e a projeção do mês.</p>
        </div>
        <div class="goals-template__opportunity">
          <span>Oportunidade em pendências</span>
          <strong>{{ formatCurrency(goals.summary.pendingOpportunityCents) }}</strong>
          <small>Recebimentos que podem acelerar o resultado do mês.</small>
        </div>
      </section>

      <section class="goals-template__targets panel">
        <header class="goals-template__section-heading">
          <div><span>Defina a direção</span><h2>Metas mensais</h2></div>
          <small v-if="saveState === 'saving'">Salvando alterações...</small>
          <small v-else-if="saveState === 'saved'">Alterações salvas automaticamente</small>
          <small v-else-if="saveState === 'error'" class="goals-template__save-error">Não foi possível salvar. Tente editar novamente.</small>
          <small v-else>Digite em reais; a máscara formata automaticamente.</small>
        </header>
        <div class="goals-template__target-fields">
          <label class="field"><span>Faturamento</span><input :value="currencyInput(form.revenueTargetCents)" type="text" inputmode="decimal" @input="updateCents('revenueTargetCents', $event)" /></label>
          <label class="field"><span>Mão de obra</span><input :value="currencyInput(form.laborTargetCents)" type="text" inputmode="decimal" @input="updateCents('laborTargetCents', $event)" /></label>
          <label class="field"><span>Produtos</span><input :value="currencyInput(form.productsTargetCents)" type="text" inputmode="decimal" @input="updateCents('productsTargetCents', $event)" /></label>
          <label class="field"><span>Clientes atendidos</span><input v-model.number="form.clientsTarget" type="number" min="0" step="1" inputmode="numeric" @input="scheduleSave" /></label>
          <label class="field"><span>Lucro líquido</span><input :value="currencyInput(form.netProfitTargetCents)" type="text" inputmode="decimal" @input="updateCents('netProfitTargetCents', $event)" /></label>
        </div>
      </section>

      <section class="goals-template__progress-section">
        <header class="goals-template__section-heading"><div><span>Acompanhamento</span><h2>Como você está indo</h2></div><small>Realizado no mês atual</small></header>
        <div class="goals-template__progress-grid">
          <article v-for="item in metrics" :key="item.key" class="goals-template__progress-card panel">
            <div class="goals-template__progress-top"><span>{{ item.label }}</span><strong>{{ progress(item) }}%</strong></div>
            <b>{{ value(item, item.actual) }}</b>
            <div class="goals-template__progress-track"><i :style="{ width: `${progress(item)}%` }" /></div>
            <p><span>Faltam {{ value(item, remaining(item)) }}</span><span>Meta {{ value(item, item.target) }}</span></p>
            <small>Projeção: {{ value(item, item.projection) }}</small>
          </article>
        </div>
      </section>

      <GoalsProgressChart v-if="goals.summary" :summary="goals.summary" />

      <section class="goals-template__insights-grid">
        <article class="goals-template__comparison panel">
          <header class="goals-template__section-heading"><div><span>Comparativo</span><h2>Meta versus resultado</h2></div></header>
          <div v-for="item in metrics" :key="item.key" class="goals-template__chart-row">
            <div class="goals-template__chart-label"><strong>{{ item.shortLabel }}</strong><small>{{ value(item, item.actual) }} de {{ value(item, item.target) }}</small></div>
            <div class="goals-template__chart-bars" aria-hidden="true">
              <i class="goals-template__bar goals-template__bar--previous" :style="{ width: barWidth(item, item.previous) }" /><i class="goals-template__bar goals-template__bar--target" :style="{ width: barWidth(item, item.target) }" /><i class="goals-template__bar goals-template__bar--actual" :style="{ width: barWidth(item, item.actual) }" />
            </div>
          </div>
          <footer class="goals-template__legend"><span><i class="goals-template__bar--previous" />Mês anterior</span><span><i class="goals-template__bar--target" />Meta</span><span><i class="goals-template__bar--actual" />Atual</span></footer>
        </article>

        <article class="goals-template__requirements panel">
          <header class="goals-template__section-heading"><div><span>Ritmo necessário</span><h2>Referências para decidir</h2></div></header>
          <div v-for="item in required" :key="item.label" class="goals-template__requirement"><span>{{ item.label }}</span><strong>{{ item.value }}</strong></div>
        </article>
      </section>

      <section class="goals-template__bottom-grid">
        <article class="goals-template__tips panel">
          <header class="goals-template__section-heading"><div><span>Ações recomendadas</span><h2>Dicas para este mês</h2></div></header>
          <div v-if="goals.summary.tips.length" class="goals-template__tips-list"><article v-for="tip in goals.summary.tips" :key="`${tip.kind}-${tip.title}`"><span>{{ tip.kind }}</span><h3>{{ tip.title }}</h3><p>{{ tip.description }}</p></article></div>
          <p v-else class="goals-template__quiet">Continue registrando pagamentos e gastos: em breve teremos recomendações mais específicas.</p>
        </article>

        <article class="goals-template__pricing panel">
          <header class="goals-template__section-heading"><div><span>Precificação</span><h2>Produtos para revisar</h2></div></header>
          <div v-if="goals.summary.pricingRecommendations.length" class="goals-template__pricing-list">
            <article v-for="product in goals.summary.pricingRecommendations" :key="product.stockItemId" :class="{ 'goals-template__price-row--alert': product.belowMinimum }" class="goals-template__price-row">
              <div><strong>{{ product.name }}</strong><small>Vender {{ product.suggestedQuantity }} un. no mês · margem {{ product.markupPercent }}%</small></div>
              <div><small>Preço mínimo</small><b>{{ formatCurrency(product.minimumPriceCents) }}</b><em v-if="product.belowMinimum">Abaixo do mínimo</em></div>
            </article>
          </div>
          <p v-else class="goals-template__quiet">Ainda não há produtos vendidos no mês anterior para gerar recomendações.</p>
        </article>
      </section>
    </template>
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
