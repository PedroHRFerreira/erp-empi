<script lang="ts">
import { computed, defineComponent, type PropType } from 'vue'
import type { ReceiptForm } from '../../../stores/useReceiptsStore'
import { formatCurrency } from '../../../utils/format'
import { maskCurrency } from '../../../utils/masks'

export default defineComponent({
  name: 'ReceiptServicesStep',
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
    installmentOptions: {
      type: Array as PropType<number[]>,
      required: true
    },
    installmentValueFor: {
      type: Function as PropType<(installments: number) => number>,
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
    },
    totalCents: {
      type: Number,
      required: true
    }
  },
  emits: ['clear-field-error', 'payment-change', 'sync-card-fee', 'update:discount-input', 'update:labor-price-input'],
  setup(props, { emit }) {
    const laborPrice = computed({
      get: () => props.laborPriceInput,
      set: (value: string) => emit('update:labor-price-input', value)
    })
    const discount = computed({
      get: () => props.discountInput,
      set: (value: string) => emit('update:discount-input', value)
    })

    const isCardPayment = computed(() => {
      return props.form.paymentMethod === 'credit_card' || props.form.paymentMethod === 'debit_card'
    })

    const activeFeePercent = computed({
      get: () => {
        if (props.form.paymentMethod === 'credit_card' && (props.form.installments || 1) > 1) {
          return props.form.installmentFeePercent
        }
        return props.form.machineFeePercent
      },
      set: (value: number) => {
        if (props.form.paymentMethod === 'credit_card' && (props.form.installments || 1) > 1) {
          props.form.installmentFeePercent = Number(value || 0)
        } else {
          props.form.machineFeePercent = Number(value || 0)
        }
        emit('clear-field-error', 'cardFeePercent')
        emit('sync-card-fee')
      }
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

    function updatePaymentMethod() {
      emit('payment-change')
      emit('sync-card-fee')
    }

    function updateInstallments() {
      clearFieldError('installments')
      emit('sync-card-fee')
    }

    return {
      activeFeePercent,
      clearFieldError,
      discount,
      formatCurrency,
      isCardPayment,
      laborPrice,
      maskDiscount,
      maskLaborPrice,
      updateInstallments,
      updatePaymentMethod
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

      <label class="field" :class="{ 'field--error': fieldErrors.paymentMethod }">
        <span>Forma de pagamento</span>
        <select v-model="form.paymentMethod" required @change="updatePaymentMethod">
          <option value="cash">Dinheiro</option>
          <option value="pix">Pix</option>
          <option value="debit_card">Cartão de débito</option>
          <option value="credit_card">Cartão de crédito</option>
        </select>
        <small v-if="fieldErrors.paymentMethod" class="field__error">{{ fieldErrors.paymentMethod }}</small>
      </label>

      <label v-if="form.paymentMethod === 'credit_card'" class="field" :class="{ 'field--error': fieldErrors.installments }">
        <span>Parcelas</span>
        <select v-model.number="form.installments" @change="updateInstallments">
          <option v-for="installment in installmentOptions" :key="installment" :value="installment">
            {{ installment }}x de {{ formatCurrency(installmentValueFor(installment)) }}
          </option>
        </select>
        <small v-if="fieldErrors.installments" class="field__error">{{ fieldErrors.installments }}</small>
      </label>

      <label v-if="isCardPayment" class="field" :class="{ 'field--error': fieldErrors.cardFeePercent }">
        <span>{{ activeCardFeeLabel }} (%)</span>
        <input v-model.number="activeFeePercent" min="0" step="0.1" type="number" @input="clearFieldError('cardFeePercent')" />
        <small v-if="fieldErrors.cardFeePercent" class="field__error">{{ fieldErrors.cardFeePercent }}</small>
      </label>

      <label class="field receipt-step__wide" :class="{ 'field--error': fieldErrors.services }">
        <span>Descrição do serviço</span>
        <textarea v-model="form.services" required placeholder="Troca de óleo, diagnóstico elétrico..." @input="clearFieldError('services')" />
        <small v-if="fieldErrors.services" class="field__error">{{ fieldErrors.services }}</small>
      </label>
    </div>

    <section class="receipt-mini-summary" aria-label="Resumo dos serviços">
      <div>
        <span>Subtotal</span>
        <strong>{{ formatCurrency(subtotalCents) }}</strong>
      </div>
      <div v-if="discountCents">
        <span>Desconto</span>
        <strong>{{ formatCurrency(-discountCents) }}</strong>
      </div>
      <div v-if="isCardPayment">
        <span>{{ activeCardFeeLabel }}</span>
        <strong>{{ formatCurrency(cardFeeCents) }}</strong>
      </div>
      <div>
        <span>Total parcial</span>
        <strong>{{ formatCurrency(totalCents) }}</strong>
      </div>
    </section>
  </section>
</template>
