<script lang="ts">
import { Pencil, Trash2 } from '@lucide/vue'
import { defineComponent, type PropType } from 'vue'
import type { IExpense } from '../../../../server/contracts/types'
import { formatCurrency } from '../../../utils/format'
import { receiptClientName, receiptVehicleName, receiptVehiclePlate } from '../../../utils/receiptDisplay'
import IconActionButton from '../../atoms/IconActionButton/Index.vue'
import EmptyState from '../../molecules/EmptyState/Index.vue'

export default defineComponent({
  name: 'ExpensesTable',
  components: {
    EmptyState,
    IconActionButton,
    Pencil,
    Trash2
  },
  props: {
    expenses: {
      type: Array as PropType<IExpense[]>,
      required: true
    }
  },
  emits: ['edit', 'remove'],
  setup(_, { emit }) {
    function edit(expense: IExpense) {
      emit('edit', expense)
    }

    function remove(expense: IExpense) {
      emit('remove', expense)
    }

    function receiptLabel(expense: IExpense) {
      if (!expense.receipt) return 'Sem vínculo'
      return `${receiptClientName(expense.receipt)} - ${receiptVehiclePlate(expense.receipt)}`
    }

    return {
      edit,
      formatCurrency,
      formatExpenseDate,
      receiptLabel,
      receiptVehicleName,
      remove
    }
  }
})

function formatExpenseDate(value: string) {
  if (!value) return '-'
  const date = new Date(value)
  return new Intl.DateTimeFormat('pt-BR').format(date)
}
</script>

<template>
  <section class="expenses-table panel">
    <div v-if="expenses.length" class="expenses-table__head" aria-hidden="true">
      <span>Descrição</span>
      <span>Categoria</span>
      <span>Recibo</span>
      <span>Valor</span>
      <span>Data</span>
      <span>Ações</span>
    </div>

    <article v-for="expense in expenses" :key="expense.id" class="expenses-table__row">
      <div class="expenses-table__cell">
        <span class="expenses-table__label">Descrição</span>
        <strong>{{ expense.description }}</strong>
        <small v-if="expense.notes">{{ expense.notes }}</small>
      </div>
      <div class="expenses-table__cell">
        <span class="expenses-table__label">Categoria</span>
        <strong>{{ expense.category }}</strong>
      </div>
      <div class="expenses-table__cell">
        <span class="expenses-table__label">Recibo</span>
        <strong>{{ receiptLabel(expense) }}</strong>
        <small v-if="expense.receipt">{{ receiptVehicleName(expense.receipt) }}</small>
      </div>
      <div class="expenses-table__cell expenses-table__cell--money">
        <span class="expenses-table__label">Valor</span>
        <strong>{{ formatCurrency(expense.amountCents) }}</strong>
      </div>
      <div class="expenses-table__cell">
        <span class="expenses-table__label">Data</span>
        <strong>{{ formatExpenseDate(expense.spentAt) }}</strong>
      </div>
      <div class="expenses-table__actions">
        <IconActionButton title="Editar gasto" @click="edit(expense)">
          <Pencil :size="16" />
        </IconActionButton>
        <IconActionButton title="Remover gasto" variant="danger" @click="remove(expense)">
          <Trash2 :size="16" />
        </IconActionButton>
      </div>
    </article>

    <EmptyState
      v-if="!expenses.length"
      title="Nenhum gasto encontrado"
      description="Adicione gastos do período para acompanhar o lucro real da oficina."
    />
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
