<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import type { ReceiptForm } from '../../../stores/useReceiptsStore'
import { formatCurrency } from '../../../utils/format'
import ReceiptPaymentsEditor from '../ReceiptPaymentsEditor/Index.vue'

export default defineComponent({
  name: 'ReceiptFinalizeStep',
  components: { ReceiptPaymentsEditor },
  props: {
    fieldErrors: {
      type: Object as PropType<Record<string, string>>,
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
    stepLabel: {
      type: String,
      default: 'Etapa 6'
    },
    totalCents: {
      type: Number,
      required: true
    }
  },
  emits: ['clear-field-error'],
  setup(props, { emit }) {
    function clearFieldError(field: string) {
      emit('clear-field-error', field)
    }

    return {
      clearFieldError,
      formatCurrency
    }
  }
})
</script>

<template>
  <section class="receipt-step">
    <header class="receipt-step__header">
      <span>{{ stepLabel }}</span>
      <h2>Finalizar recibo</h2>
    </header>

    <label class="field">
      <span>Descrição do recibo</span>
      <textarea v-model="form.notes" placeholder="Informações adicionais" @input="clearFieldError('notes')" />
    </label>

    <ReceiptPaymentsEditor
      :form="form"
      :field-errors="fieldErrors"
      :subtotal-cents="subtotalCents"
      @clear-field-error="clearFieldError"
    />

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
      <div v-if="cardFeeCents">
        <span>Taxas dos cartões</span>
        <strong>{{ formatCurrency(cardFeeCents) }}</strong>
      </div>
      <div class="receipt-summary__total">
        <span>Total</span>
        <strong>{{ formatCurrency(totalCents) }}</strong>
      </div>
    </section>

    <p v-if="error" class="form-error">{{ error }}</p>
  </section>
</template>
