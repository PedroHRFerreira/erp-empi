<script lang="ts">
import { computed, defineComponent, type PropType } from 'vue'
import type { ReceiptForm, ReceiptServiceExpenseForm } from '../../../stores/useReceiptsStore'
import { formatCurrency } from '../../../utils/format'
import { maskCurrency } from '../../../utils/masks'

export default defineComponent({
  name: 'ReceiptServiceExpensesStep',
  props: {
    categories: {
      type: Array as PropType<string[]>,
      required: true
    },
    fieldErrors: {
      type: Object as PropType<Record<string, string>>,
      required: true
    },
    form: {
      type: Object as PropType<ReceiptForm>,
      required: true
    },
    serviceExpense: {
      type: Object as PropType<ReceiptServiceExpenseForm>,
      required: true
    },
    serviceExpenseAmountInput: {
      type: String,
      required: true
    },
    serviceExpenseError: {
      type: String,
      default: ''
    },
    serviceExpensesTotalCents: {
      type: Number,
      required: true
    }
  },
  emits: [
    'add-service-expense',
    'clear-service-expenses-error',
    'remove-service-expense',
    'update:service-expense-amount-input'
  ],
  setup(props, { emit }) {
    const amountInput = computed({
      get: () => props.serviceExpenseAmountInput,
      set: (value: string) => emit('update:service-expense-amount-input', value)
    })

    function clearServiceExpensesError() {
      emit('clear-service-expenses-error')
    }

    function maskServiceExpenseAmount() {
      clearServiceExpensesError()
      amountInput.value = maskCurrency(amountInput.value)
    }

    return {
      amountInput,
      clearServiceExpensesError,
      formatCurrency,
      maskServiceExpenseAmount
    }
  }
})
</script>

<template>
  <section class="receipt-step">
    <header class="receipt-step__header">
      <span>Etapa 5</span>
      <h2>Gastos do serviço</h2>
    </header>

    <div class="service-expenses" :class="{ 'service-expenses--error': fieldErrors.serviceExpenses }">
      <header class="service-expenses__header">
        <span>Total de gastos</span>
        <strong>{{ formatCurrency(serviceExpensesTotalCents) }}</strong>
      </header>

      <div class="service-expenses__form">
        <label class="field">
          <span>Descrição</span>
          <input v-model="serviceExpense.description" placeholder="Gasolina, peça avulsa..." @input="clearServiceExpensesError" />
        </label>

        <label class="field">
          <span>Categoria</span>
          <select v-model="serviceExpense.category" @change="clearServiceExpensesError">
            <option value="">Selecione</option>
            <option v-for="category in categories" :key="category" :value="category">
              {{ category }}
            </option>
          </select>
        </label>

        <label class="field">
          <span>Valor</span>
          <input v-model="amountInput" inputmode="numeric" placeholder="R$ 80,00" @input="maskServiceExpenseAmount" />
        </label>

        <label class="field">
          <span>Data</span>
          <input v-model="serviceExpense.spentAt" type="date" @input="clearServiceExpensesError" />
        </label>

        <button class="button button--secondary" type="button" @click="$emit('add-service-expense')">Adicionar gasto</button>

        <label class="field service-expenses__notes">
          <span>Observações</span>
          <input v-model="serviceExpense.notes" placeholder="Detalhes do gasto" @input="clearServiceExpensesError" />
        </label>
      </div>

      <ul v-if="form.serviceExpenses.length">
        <li v-for="(expense, index) in form.serviceExpenses" :key="`${expense.description}-${index}`">
          <span>
            {{ expense.description }}
            <small>{{ expense.category }} / {{ formatCurrency(expense.amountCents) }}</small>
          </span>
          <button type="button" @click="$emit('remove-service-expense', index)">Remover</button>
        </li>
      </ul>

      <p v-else class="receipt-step__empty">Nenhum gasto vinculado a este serviço.</p>
      <small v-if="serviceExpenseError" class="field__error">{{ serviceExpenseError }}</small>
      <small v-if="fieldErrors.serviceExpenses" class="field__error">{{ fieldErrors.serviceExpenses }}</small>
    </div>
  </section>
</template>
