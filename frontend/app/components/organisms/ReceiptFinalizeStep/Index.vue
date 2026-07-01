<script lang="ts">
import { computed, defineComponent, type PropType } from 'vue'
import type { ReceiptForm } from '../../../stores/useReceiptsStore'
import { formatCurrency } from '../../../utils/format'

export default defineComponent({
  name: 'ReceiptFinalizeStep',
  props: {
    activeCardFeeLabel: {
      type: String,
      required: true
    },
    cardFeeCents: {
      type: Number,
      required: true
    },
    discountCents: {
      type: Number,
      required: true
    },
    error: {
      type: String,
      default: ''
    },
    form: {
      type: Object as PropType<ReceiptForm>,
      required: true
    },
    installmentValueCents: {
      type: Number,
      required: true
    },
    laborPriceCents: {
      type: Number,
      required: true
    },
    productsTotalCents: {
      type: Number,
      required: true
    },
    serviceExpensesTotalCents: {
      type: Number,
      required: true
    },
    subtotalCents: {
      type: Number,
      required: true
    },
    totalCents: {
      type: Number,
      required: true
    }
  },
  emits: ['clear-field-error'],
  setup(props, { emit }) {
    const isCardPayment = computed(() => {
      return props.form.paymentMethod === 'credit_card' || props.form.paymentMethod === 'debit_card'
    })

    function clearFieldError(field: string) {
      emit('clear-field-error', field)
    }

    return {
      clearFieldError,
      formatCurrency,
      isCardPayment
    }
  }
})
</script>

<template>
  <section class="receipt-step">
    <header class="receipt-step__header">
      <span>Etapa 6</span>
      <h2>Finalizar recibo</h2>
    </header>

    <label class="field">
      <span>Descrição do recibo</span>
      <textarea v-model="form.notes" placeholder="Informações adicionais" @input="clearFieldError('notes')" />
    </label>

    <section class="receipt-summary" aria-label="Resumo do recibo">
      <div>
        <span>Mão de obra</span>
        <strong>{{ formatCurrency(laborPriceCents) }}</strong>
      </div>
      <div>
        <span>Produtos utilizados</span>
        <strong>{{ formatCurrency(productsTotalCents) }}</strong>
      </div>
      <div v-if="serviceExpensesTotalCents">
        <span>Gastos do serviço</span>
        <strong>{{ formatCurrency(serviceExpensesTotalCents) }}</strong>
      </div>
      <div v-if="discountCents">
        <span>Desconto</span>
        <strong>{{ formatCurrency(-discountCents) }}</strong>
      </div>
      <div>
        <span>Subtotal</span>
        <strong>{{ formatCurrency(subtotalCents) }}</strong>
      </div>
      <div v-if="isCardPayment">
        <span>{{ activeCardFeeLabel }} ({{ form.cardFeePercent }}%)</span>
        <strong>{{ formatCurrency(cardFeeCents) }}</strong>
      </div>
      <div class="receipt-summary__total">
        <span>Total</span>
        <strong>{{ formatCurrency(totalCents) }}</strong>
      </div>
      <div v-if="form.paymentMethod === 'credit_card'" class="receipt-summary__installment">
        <span>Parcelamento</span>
        <strong>{{ form.installments }}x de {{ formatCurrency(installmentValueCents) }}</strong>
      </div>
    </section>

    <p v-if="error" class="form-error">{{ error }}</p>
  </section>
</template>
