<script lang="ts">
import { computed, defineComponent, type PropType } from 'vue'
import type { IGoalsSummary } from '../../../../server/contracts/types'
import { formatCurrency } from '../../../utils/format'

type MetricKind = 'currency' | 'number'

interface IChartMetric {
  key: string
  label: string
  actual: number
  target: number
  kind: MetricKind
}

const CHART_BASELINE = 164
const CHART_HEIGHT = 124

export default defineComponent({
  name: 'GoalsProgressChart',
  props: {
    summary: {
      type: Object as PropType<IGoalsSummary>,
      required: true
    }
  },
  setup(props) {
    const metrics = computed<IChartMetric[]>(() => [
      {
        key: 'revenue',
        label: 'Faturamento',
        actual: Math.max(0, props.summary.actual.revenueCents || 0),
        target: Math.max(0, props.summary.targets.revenueTargetCents || 0),
        kind: 'currency'
      },
      {
        key: 'labor',
        label: 'Mão de obra',
        actual: Math.max(0, props.summary.actual.laborCents || 0),
        target: Math.max(0, props.summary.targets.laborTargetCents || 0),
        kind: 'currency'
      },
      {
        key: 'products',
        label: 'Produtos',
        actual: Math.max(0, props.summary.actual.productsCents || 0),
        target: Math.max(0, props.summary.targets.productsTargetCents || 0),
        kind: 'currency'
      },
      {
        key: 'clients',
        label: 'Clientes',
        actual: Math.max(0, props.summary.actual.clients || 0),
        target: Math.max(0, props.summary.targets.clientsTarget || 0),
        kind: 'number'
      },
      {
        key: 'profit',
        label: 'Lucro líquido',
        actual: Math.max(0, props.summary.actual.netProfitCents || 0),
        target: Math.max(0, props.summary.targets.netProfitTargetCents || 0),
        kind: 'currency'
      }
    ])

    const hasData = computed(() => metrics.value.some(({ actual, target }) => actual > 0 || target > 0))
    const highestValue = computed(() => Math.max(1, ...metrics.value.flatMap(({ actual, target }) => [actual, target])))

    const chartPoints = computed(() => metrics.value.map((metric, index) => {
      const x = 26 + index * 82
      return { ...metric, x, actualY: CHART_BASELINE - (metric.actual / highestValue.value) * CHART_HEIGHT, targetY: CHART_BASELINE - (metric.target / highestValue.value) * CHART_HEIGHT }
    }))
    const areaPath = computed(() => (field: 'actualY' | 'targetY') => {
      const points = chartPoints.value
      if (!points.length) return ''
      return `M ${points[0]?.x} ${CHART_BASELINE} L ${points.map((point) => `${point.x} ${point[field]}`).join(' L ')} L ${points.at(-1)?.x} ${CHART_BASELINE} Z`
    })

    function formatValue(value: number, kind: MetricKind) {
      return kind === 'currency' ? formatCurrency(value) : new Intl.NumberFormat('pt-BR').format(value)
    }

    function percentage(metric: IChartMetric) {
      if (!metric.target) return metric.actual ? null : 0
      return Math.round((metric.actual / metric.target) * 100)
    }

    return { areaPath, chartPoints, formatValue, hasData, metrics, percentage }
  }
})
</script>

<template>
  <section class="goals-progress panel" aria-labelledby="goals-progress-title">
    <div class="goals-progress__heading">
      <div>
        <p class="goals-progress__eyebrow">Acompanhamento mensal</p>
        <h2 id="goals-progress-title">Visão mensal: realizado x meta</h2>
      </div>
      <div class="goals-progress__legend" aria-label="Legenda do gráfico">
        <span><i class="goals-progress__legend-target" aria-hidden="true" />Meta</span>
        <span><i class="goals-progress__legend-actual" aria-hidden="true" />Realizado</span>
      </div>
    </div>

    <div v-if="hasData" class="goals-progress__chart-wrap">
      <svg class="goals-progress__chart" viewBox="0 0 380 208" role="img" aria-labelledby="goals-chart-title goals-chart-description">
        <title id="goals-chart-title">Gráfico de área do realizado comparado à meta mensal</title>
        <desc id="goals-chart-description">As áreas mostram o realizado e a meta de faturamento, mão de obra, produtos, clientes e lucro líquido no mês atual.</desc>
        <line class="goals-progress__axis" x1="18" y1="164" x2="366" y2="164" />
        <path class="goals-progress__area goals-progress__area--target" :d="areaPath('targetY')" />
        <path class="goals-progress__area goals-progress__area--actual" :d="areaPath('actualY')" />
        <g v-for="point in chartPoints" :key="point.key">
          <circle class="goals-progress__point goals-progress__point--target" :cx="point.x" :cy="point.targetY" r="3"><title>{{ point.label }}: meta {{ formatValue(point.target, point.kind) }}</title></circle>
          <circle class="goals-progress__point goals-progress__point--actual" :cx="point.x" :cy="point.actualY" r="3"><title>{{ point.label }}: realizado {{ formatValue(point.actual, point.kind) }}</title></circle>
        </g>
      </svg>

      <div class="goals-progress__metrics">
        <article v-for="metric in metrics" :key="metric.key" class="goals-progress__metric">
          <strong>{{ metric.label }}</strong>
          <span class="goals-progress__values">
            <b>{{ formatValue(metric.actual, metric.kind) }}</b>
            <span>de {{ formatValue(metric.target, metric.kind) }}</span>
          </span>
          <span class="goals-progress__percent">
            <template v-if="percentage(metric) !== null">{{ percentage(metric) }}% da meta</template>
            <template v-else>Meta ainda não definida</template>
          </span>
        </article>
      </div>
    </div>

    <div v-else class="goals-progress__empty" role="status">
      <strong>Ainda não há dados para comparar.</strong>
      <p>Defina suas metas e registre recebimentos para acompanhar a evolução deste mês.</p>
    </div>
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
