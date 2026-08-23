<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import type { IRealizedExpense } from '../../../../server/contracts/types'
import { formatCurrency, formatDate } from '../../../utils/format'
import EmptyState from '../../molecules/EmptyState/Index.vue'
import { realizedExpenseOriginLabel, realizedExpensePaymentMethodLabel } from './presentation'

export default defineComponent({
  name: 'ExpensesTable',
  components: {
    EmptyState
  },
  props: {
    expenses: {
      type: Array as PropType<IRealizedExpense[]>,
      required: true
    }
  },
  emits: ['edit', 'remove'],
  setup(_, { emit }) {
    function edit(expense: IRealizedExpense) {
      emit('edit', expense)
    }

    function remove(expense: IRealizedExpense) {
      emit('remove', expense)
    }

    return {
      edit,
      formatCurrency,
      formatExpenseDate: formatDate,
      originLabel: realizedExpenseOriginLabel,
      paymentMethodLabel: realizedExpensePaymentMethodLabel,
      remove
    }
  }
})

</script>

<template>
  <section class="expenses-table panel">
    <div v-if="expenses.length" class="expenses-table__head" aria-hidden="true">
      <span>Descrição</span>
      <span>Categoria</span>
      <span>Origem</span>
      <span>Meio</span>
      <span>Valor</span>
      <span>Data</span>
      <span>Ações</span>
    </div>

    <article v-for="expense in expenses" :key="expense.id" class="expenses-table__row">
      <div class="expenses-table__cell">
        <span class="expenses-table__label">Descrição</span>
        <strong>{{ expense.description }}</strong>
        <small v-if="expense.supplierName">Fornecedor: {{ expense.supplierName }}</small>
      </div>
      <div class="expenses-table__cell">
        <span class="expenses-table__label">Categoria</span>
        <strong>{{ expense.category }}</strong>
      </div>
      <div class="expenses-table__cell">
        <span class="expenses-table__label">Origem</span>
        <span class="expenses-table__origin" :class="`expenses-table__origin--${expense.origin}`">
          {{ originLabel(expense) }}
        </span>
        <small v-if="expense.installmentNumber">Parcela {{ expense.installmentNumber }}</small>
      </div>
      <div class="expenses-table__cell">
        <span class="expenses-table__label">Meio</span>
        <strong>{{ paymentMethodLabel(expense.paymentMethod) }}</strong>
      </div>
      <div class="expenses-table__cell expenses-table__cell--money">
        <span class="expenses-table__label">Valor</span>
        <strong>{{ formatCurrency(expense.amountCents) }}</strong>
      </div>
      <div class="expenses-table__cell">
        <span class="expenses-table__label">Data</span>
        <strong>{{ formatExpenseDate(expense.occurredAt) }}</strong>
      </div>
      <div v-if="expense.editable && expense.origin === 'operational'" class="expenses-table__actions">
        <NuxtLink
          v-if="expense.expenseId"
          class="button button--secondary expenses-table__action"
          :to="`/expenses/${expense.expenseId}/history`"
        >
          Ver histórico
        </NuxtLink>
        <button class="button button--secondary expenses-table__action" type="button" @click="edit(expense)">Editar</button>
        <button class="button button--danger expenses-table__action" type="button" @click="remove(expense)">Remover</button>
      </div>
      <div v-else class="expenses-table__readonly">
        <NuxtLink
          v-if="expense.expenseId"
          class="button button--secondary expenses-table__action"
          :to="`/expenses/${expense.expenseId}/history`"
        >
          Ver histórico
        </NuxtLink>
        <NuxtLink
          v-else-if="expense.stockPurchaseId"
          class="button button--secondary expenses-table__action"
          :to="`/expenses/stock-${expense.stockPurchaseId}/history`"
        >
          Ver histórico
        </NuxtLink>
        <span v-else>Somente leitura</span>
      </div>
    </article>

    <EmptyState
      v-if="!expenses.length"
      title="Nenhum gasto encontrado"
      description="Nenhuma saída efetivamente paga foi encontrada neste período."
    />
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
