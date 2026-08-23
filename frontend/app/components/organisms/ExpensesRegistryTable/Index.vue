<script setup lang="ts">
import type { IExpense } from '../../../../server/contracts/types'
import { formatCurrency, formatDate } from '../../../utils/format'
import EmptyState from '../../molecules/EmptyState/Index.vue'

defineProps<{ expenses: IExpense[] }>()
const emit = defineEmits<{ edit: [expense: IExpense]; remove: [expense: IExpense] }>()

function hasPaidInstallment(expense: IExpense) {
  return expense.installments?.some((installment) => installment.status === 'paid') || false
}

function statusLabel(expense: IExpense) {
  if (!expense.installments?.length) return 'Legado'
  const paid = expense.installments.filter((installment) => installment.status === 'paid').length
  if (paid === expense.installments.length) return 'Pago'
  if (paid > 0) return 'Parcial'
  return 'Pendente'
}

function statusClass(expense: IExpense) {
  return statusLabel(expense).toLocaleLowerCase('pt-BR')
}
</script>

<template>
  <section class="expenses-registry panel">
    <div v-if="expenses.length" class="expenses-registry__head" aria-hidden="true">
      <span>Descrição</span><span>Categoria</span><span>Valor</span><span>Lançamento</span><span>Situação</span><span>Ações</span>
    </div>

    <article v-for="expense in expenses" :key="expense.id" class="expenses-registry__row">
      <div class="expenses-registry__cell">
        <span class="expenses-registry__label">Descrição</span>
        <strong>{{ expense.description }}</strong>
        <small v-if="expense.notes">{{ expense.notes }}</small>
      </div>
      <div class="expenses-registry__cell">
        <span class="expenses-registry__label">Categoria</span><strong>{{ expense.category }}</strong>
      </div>
      <div class="expenses-registry__cell expenses-registry__cell--money">
        <span class="expenses-registry__label">Valor</span><strong>{{ formatCurrency(expense.amountCents) }}</strong>
      </div>
      <div class="expenses-registry__cell">
        <span class="expenses-registry__label">Lançamento</span><strong>{{ formatDate(expense.spentAt) }}</strong>
      </div>
      <div class="expenses-registry__cell">
        <span class="expenses-registry__label">Situação</span>
        <span class="expenses-registry__status" :class="`expenses-registry__status--${statusClass(expense)}`">{{ statusLabel(expense) }}</span>
        <small v-if="expense.installments?.length">{{ expense.installments.length }} parcela{{ expense.installments.length === 1 ? '' : 's' }}</small>
      </div>
      <div class="expenses-registry__actions">
        <NuxtLink class="button button--secondary expenses-registry__action expenses-registry__action--history" :to="`/expenses/${expense.id}/history`">Ver histórico</NuxtLink>
        <button class="button button--secondary expenses-registry__action" type="button" :disabled="hasPaidInstallment(expense)" :title="hasPaidInstallment(expense) ? 'Gastos com pagamentos não podem ser editados.' : undefined" @click="emit('edit', expense)">Editar</button>
        <button class="button button--danger expenses-registry__action" type="button" :disabled="hasPaidInstallment(expense)" :title="hasPaidInstallment(expense) ? 'Gastos com pagamentos não podem ser removidos.' : undefined" @click="emit('remove', expense)">Remover</button>
      </div>
    </article>

    <EmptyState v-if="!expenses.length" title="Nenhum lançamento encontrado" description="Nenhum gasto operacional foi lançado neste período." />
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
