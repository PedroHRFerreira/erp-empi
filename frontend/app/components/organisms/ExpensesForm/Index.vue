<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import type { IReceipt } from '../../../../server/contracts/types'
import type { ExpenseForm as IExpenseForm } from '../../../stores/useExpensesStore'
import { maskCurrency } from '../../../utils/masks'
import { receiptClientName, receiptVehicleLine } from '../../../utils/receiptDisplay'

export default defineComponent({
  name: 'ExpensesForm',
  props: {
    form: {
      type: Object as PropType<IExpenseForm>,
      required: true
    },
    amountInput: {
      type: String,
      required: true
    },
    categories: {
      type: Array as PropType<string[]>,
      required: true
    },
    receiptOptions: {
      type: Array as PropType<IReceipt[]>,
      default: () => []
    },
    fieldErrors: {
      type: Object as PropType<Record<string, string>>,
      required: true
    },
    error: {
      type: String,
      default: ''
    },
    saving: {
      type: Boolean,
      default: false
    }
  },
  emits: ['cancel', 'clear-field-error', 'save', 'update:amount-input'],
  setup(_, { emit }) {
    function clearFieldError(field: string) {
      emit('clear-field-error', field)
    }

    function save() {
      emit('save')
    }

    function cancel() {
      emit('cancel')
    }

    function updateAmountInput(event: Event) {
      clearFieldError('amountCents')
      const input = event.target as HTMLInputElement
      emit('update:amount-input', maskCurrency(input.value))
    }

    function receiptLabel(receipt: IReceipt) {
      return `${receiptClientName(receipt)} - ${receiptVehicleLine(receipt)}`
    }

    return {
      cancel,
      clearFieldError,
      receiptLabel,
      save,
      updateAmountInput
    }
  }
})
</script>

<template>
  <form class="expenses-form panel" novalidate @submit.prevent="save">
    <label class="field" :class="{ 'field--error': fieldErrors.description }">
      <span>Descrição</span>
      <input v-model="form.description" required placeholder="Conta de luz, gasolina, manutenção..." @input="clearFieldError('description')" />
      <small v-if="fieldErrors.description" class="field__error">{{ fieldErrors.description }}</small>
    </label>

    <label class="field" :class="{ 'field--error': fieldErrors.category }">
      <span>Categoria</span>
      <select v-model="form.category" required @change="clearFieldError('category')">
        <option value="" disabled>Selecione</option>
        <option v-for="category in categories" :key="category" :value="category">
          {{ category }}
        </option>
      </select>
      <small v-if="fieldErrors.category" class="field__error">{{ fieldErrors.category }}</small>
    </label>

    <label class="field" :class="{ 'field--error': fieldErrors.amountCents }">
      <span>Valor</span>
      <input :value="amountInput" required inputmode="numeric" placeholder="R$ 200,00" @input="updateAmountInput" />
      <small v-if="fieldErrors.amountCents" class="field__error">{{ fieldErrors.amountCents }}</small>
    </label>

    <label class="field" :class="{ 'field--error': fieldErrors.spentAt }">
      <span>Data</span>
      <input v-model="form.spentAt" required type="date" @input="clearFieldError('spentAt')" />
      <small v-if="fieldErrors.spentAt" class="field__error">{{ fieldErrors.spentAt }}</small>
    </label>

    <label class="field">
      <span>Recibo vinculado</span>
      <select v-model="form.receiptId" @change="clearFieldError('receiptId')">
        <option :value="null">Sem recibo</option>
        <option v-for="receipt in receiptOptions" :key="receipt.id" :value="receipt.id">
          {{ receiptLabel(receipt) }}
        </option>
      </select>
    </label>

    <label class="field expenses-form__wide">
      <span>Observações</span>
      <textarea v-model="form.notes" placeholder="Detalhes adicionais do gasto" @input="clearFieldError('notes')" />
    </label>

    <p v-if="error" class="form-error expenses-form__wide">{{ error }}</p>

    <div class="expenses-form__actions expenses-form__wide">
      <button class="button button--secondary" type="button" @click="cancel">Cancelar</button>
      <button class="button button--primary" type="submit" :disabled="saving">
        {{ saving ? 'Salvando...' : 'Salvar gasto' }}
      </button>
    </div>
  </form>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
