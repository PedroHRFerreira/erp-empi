<script lang="ts">
import { Trash2 } from '@lucide/vue'
import { computed, defineComponent, type PropType } from 'vue'
import type { IReceipt } from '../../../../server/contracts/types'
import type { ExpenseForm as IExpenseForm } from '../../../stores/useExpensesStore'
import { maskCurrency } from '../../../utils/masks'
import { receiptClientName, receiptVehicleLine } from '../../../utils/receiptDisplay'

export default defineComponent({
  name: 'ExpensesForm',
  components: { Trash2 },
  props: {
    form: {
      type: Object as PropType<IExpenseForm>,
      required: true
    },
    amountInput: {
      type: String,
      required: true
    },
    installmentAmountInputs: {
      type: Array as PropType<string[]>,
      required: true
    },
    installmentCount: {
      type: Number,
      required: true
    },
    firstDueDate: {
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
  emits: [
    'cancel',
    'clear-field-error',
    'generate-installments',
    'remove-installment',
    'save',
    'update:amount-input',
    'update:first-due-date',
    'update:installment-amount',
    'update:installment-count'
  ],
  setup(props, { emit }) {
    const installmentTotalCents = computed(() => props.form.installments.reduce((sum, row) => sum + row.amountCents, 0))
    const installmentTotalInput = computed(() => maskCurrency(String(installmentTotalCents.value)))
    const installmentsMatchTotal = computed(() => installmentTotalCents.value === props.form.amountCents)

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

    function updateInstallmentAmount(index: number, event: Event) {
      clearFieldError('installments')
      const input = event.target as HTMLInputElement
      emit('update:installment-amount', index, maskCurrency(input.value))
    }

    function updateInstallmentCount(event: Event) {
      emit('update:installment-count', Number((event.target as HTMLInputElement).value))
    }

    function updateFirstDueDate(event: Event) {
      emit('update:first-due-date', (event.target as HTMLInputElement).value)
    }

    function receiptLabel(receipt: IReceipt) {
      return `${receiptClientName(receipt)} - ${receiptVehicleLine(receipt)}`
    }

    return {
      cancel,
      clearFieldError,
      installmentTotalCents,
      installmentTotalInput,
      installmentsMatchTotal,
      receiptLabel,
      save,
      updateAmountInput,
      updateFirstDueDate,
      updateInstallmentAmount,
      updateInstallmentCount
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

    <section class="expenses-form__installments expenses-form__wide" aria-labelledby="expense-installments-title">
      <header>
        <div>
          <span>Contas a pagar</span>
          <h2 id="expense-installments-title">Parcelamento</h2>
        </div>
        <p>As parcelas serão reconhecidas como gasto e saída de caixa somente quando forem quitadas.</p>
      </header>

      <div class="expenses-form__generator">
        <label class="field">
          <span>Número de parcelas</span>
          <input
            :value="installmentCount"
            type="number"
            min="1"
            max="48"
            @input="updateInstallmentCount"
          />
        </label>
        <label class="field">
          <span>Primeiro vencimento</span>
          <input
            :value="firstDueDate"
            type="date"
            @input="updateFirstDueDate"
          />
        </label>
        <button class="button button--secondary" type="button" @click="$emit('generate-installments')">Gerar parcelas</button>
      </div>

      <div class="expenses-form__installment-list">
        <div v-for="(installment, index) in form.installments" :key="index" class="expenses-form__installment-row">
          <strong>{{ index + 1 }}ª</strong>
          <label class="field">
            <span>Valor</span>
            <input
              :value="installmentAmountInputs[index]"
              inputmode="numeric"
              @input="updateInstallmentAmount(index, $event)"
            />
          </label>
          <label class="field">
            <span>Vencimento</span>
            <input v-model="installment.dueDate" type="date" @input="clearFieldError('installments')" />
          </label>
          <label class="field">
            <span>Forma prevista</span>
            <select v-model="installment.plannedMethod" @change="clearFieldError('installments')">
              <option value="boleto">Boleto</option>
              <option value="pix">PIX</option>
              <option value="cash">Dinheiro</option>
              <option value="debit_card">Cartão de débito</option>
              <option value="credit_card">Cartão de crédito</option>
            </select>
          </label>
          <button
            class="expenses-form__remove-installment"
            type="button"
            :aria-label="`Remover ${index + 1}ª parcela`"
            :disabled="form.installments.length === 1"
            @click="$emit('remove-installment', index)"
          >
            <Trash2 :size="17" />
          </button>
        </div>
      </div>

      <div class="expenses-form__installment-total" :class="{ 'expenses-form__installment-total--invalid': !installmentsMatchTotal }">
        <span>Total das parcelas</span>
        <strong>{{ installmentTotalInput }}</strong>
      </div>
      <small v-if="fieldErrors.installments" class="field__error">{{ fieldErrors.installments }}</small>
    </section>

    <p class="expenses-form__notice expenses-form__wide">
      O gasto será criado como pendente. A forma e a data efetivas do pagamento serão informadas em Contas a pagar.
    </p>

    <p v-if="error" class="form-error expenses-form__wide" role="alert">{{ error }}</p>

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
