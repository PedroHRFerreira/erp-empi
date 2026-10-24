<script lang="ts">
import { computed, defineComponent, type PropType } from 'vue'
import type { ReceiptForm } from '../../../stores/useReceiptsStore'
import { formatCurrency } from '../../../utils/format'
import { maskCurrency } from '../../../utils/masks'

export default defineComponent({
  name: 'ReceiptServicesStep',
  props: {
    discountCents: {
      type: Number,
      required: true
    },
    discountInput: {
      type: String,
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
    laborPriceInput: {
      type: String,
      required: true
    },
    stepLabel: {
      type: String,
      default: 'Etapa 3'
    },
    subtotalCents: {
      type: Number,
      required: true
    }
  },
  emits: ['clear-field-error', 'update:discount-input', 'update:labor-price-input'],
  setup(props, { emit }) {
    const laborPrice = computed({
      get: () => props.laborPriceInput,
      set: (value: string) => emit('update:labor-price-input', value)
    })
    const discount = computed({
      get: () => props.discountInput,
      set: (value: string) => emit('update:discount-input', value)
    })

    function clearFieldError(field: string) {
      emit('clear-field-error', field)
    }

    function maskLaborPrice(event: Event) {
      clearFieldError('laborPriceCents')
      const input = event.target as HTMLInputElement
      laborPrice.value = maskCurrency(input.value)
    }

    function maskDiscount(event: Event) {
      clearFieldError('discountCents')
      const input = event.target as HTMLInputElement
      discount.value = maskCurrency(input.value)
    }

    return {
      clearFieldError,
      discount,
      formatCurrency,
      laborPrice,
      maskDiscount,
      maskLaborPrice
    }
  }
})
</script>

<template>
  <section class="receipt-step">
    <header class="receipt-step__header">
      <span>{{ stepLabel }}</span>
      <h2>Serviços</h2>
    </header>

    <div class="receipt-step__grid receipt-step__grid--two">
      <label class="field" :class="{ 'field--error': fieldErrors.laborPriceCents }">
        <span>Valor da mão de obra</span>
        <input v-model="laborPrice" required inputmode="decimal" placeholder="R$ 350,00" @input="maskLaborPrice" />
        <small v-if="fieldErrors.laborPriceCents" class="field__error">{{ fieldErrors.laborPriceCents }}</small>
      </label>

      <label class="field" :class="{ 'field--error': fieldErrors.discountCents }">
        <span>Desconto</span>
        <input v-model="discount" inputmode="decimal" placeholder="R$ 50,00" @input="maskDiscount" />
        <small v-if="fieldErrors.discountCents" class="field__error">{{ fieldErrors.discountCents }}</small>
      </label>

      <label class="field receipt-step__wide" :class="{ 'field--error': fieldErrors.services }">
        <span>Descrição do serviço</span>
        <textarea v-model="form.services" required placeholder="Troca de óleo, diagnóstico elétrico..." @input="clearFieldError('services')" />
        <small v-if="fieldErrors.services" class="field__error">{{ fieldErrors.services }}</small>
      </label>
    </div>

    <section class="receipt-mini-summary" aria-label="Resumo dos serviços">
      <div>
        <span>Subtotal após desconto</span>
        <strong>{{ formatCurrency(subtotalCents) }}</strong>
      </div>
      <div v-if="discountCents">
        <span>Desconto</span>
        <strong>{{ formatCurrency(-discountCents) }}</strong>
      </div>
    </section>
  </section>
</template>
