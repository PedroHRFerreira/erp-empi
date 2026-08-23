<script lang="ts">
import { computed, defineComponent, type PropType } from 'vue'
import type { IFinancialSummary } from '../../../../server/contracts/types'
import { formatCurrency } from '../../../utils/format'

export default defineComponent({
  name: 'FinancialSummaryGrid',
  props: {
    summary: {
      type: Object as PropType<IFinancialSummary | null>,
      default: null
    }
  },
  setup(props) {
    const health = computed(() => {
      if (!props.summary) {
        return {
          label: 'Sem dados',
          title: 'Resumo ainda não carregado',
          description: 'Carregando informações financeiras.',
          modifier: 'neutral'
        }
      }
      if (props.summary.healthStatus === 'red') {
        return {
          label: 'Vermelho',
          title: 'Operação no vermelho',
          description: 'Os custos e gastos passaram da receita recebida no período.',
          modifier: 'red'
        }
      }
      if (props.summary.healthStatus === 'yellow') {
        return {
          label: 'Atenção',
          title: 'Margem apertada',
          description: 'Existe lucro, mas a margem líquida está abaixo de 15%.',
          modifier: 'yellow'
        }
      }
      return {
        label: 'Verde',
        title: 'Operação saudável',
        description: 'A margem líquida do período está em um nível confortável.',
        modifier: 'green'
      }
    })

    const cards = computed(() => {
      const summary = props.summary
      if (!summary) return []

      return [
        {
          label: 'Total pago no período',
          value: formatCurrency(summary.totalRealizedExpensesCents || 0),
          detail: 'Somente saídas efetivamente realizadas',
          tone: 'negative'
        },
        {
          label: 'Gastos operacionais',
          value: formatCurrency(summary.operationalExpensesCents || 0),
          detail: `${summary.expensesCount || 0} lançamento${summary.expensesCount === 1 ? '' : 's'}`,
          tone: 'warning'
        },
        {
          label: 'Compras de estoque pagas',
          value: formatCurrency(summary.stockExpensesCents || 0),
          detail: `${summary.stockPaymentsCount || 0} pagamento${summary.stockPaymentsCount === 1 ? '' : 's'}`,
          tone: 'neutral'
        },
        {
          label: 'Resultado líquido',
          value: formatCurrency(summary.netProfitCents),
          detail: `Margem de ${(summary.netMarginPercent || 0).toFixed(1)}%`,
          tone: summary.netProfitCents < 0 ? 'negative' : 'positive'
        }
      ]
    })

    return {
      cards,
      health
    }
  }
})
</script>

<template>
  <section class="financial-summary">
    <article class="financial-summary__health panel" :class="`financial-summary__health--${health.modifier}`">
      <span>{{ health.label }}</span>
      <strong>{{ health.title }}</strong>
      <p>{{ health.description }}</p>
    </article>

    <div class="financial-summary__grid">
      <article v-for="card in cards" :key="card.label" class="financial-summary__card panel" :class="`financial-summary__card--${card.tone}`">
        <span>{{ card.label }}</span>
        <strong>{{ card.value }}</strong>
        <small>{{ card.detail }}</small>
      </article>
    </div>
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
