<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import type { ReceiptForm } from '../../../stores/useReceiptsStore'
import { maskPhone } from '../../../utils/masks'

export default defineComponent({
  name: 'ReceiptClientStep',
  props: {
    fieldErrors: {
      type: Object as PropType<Record<string, string>>,
      required: true
    },
    form: {
      type: Object as PropType<ReceiptForm>,
      required: true
    },
    stepLabel: {
      type: String,
      default: 'Etapa 1'
    }
  },
  emits: ['clear-field-error'],
  setup(props, { emit }) {
    function clearFieldError(field: string) {
      emit('clear-field-error', field)
    }

    function maskClientPhone() {
      clearFieldError('client.phone')
      props.form.client.phone = maskPhone(props.form.client.phone)
    }

    return {
      clearFieldError,
      maskClientPhone
    }
  }
})
</script>

<template>
  <section class="receipt-step">
    <header class="receipt-step__header">
      <span>{{ stepLabel }}</span>
      <h2>Informações do cliente</h2>
    </header>

    <div class="receipt-step__grid receipt-step__grid--two">
      <label class="field" :class="{ 'field--error': fieldErrors['client.name'] }">
        <span>Nome do cliente</span>
        <input v-model="form.client.name" required placeholder="Nome do cliente" @input="clearFieldError('client.name')" />
        <small v-if="fieldErrors['client.name']" class="field__error">{{ fieldErrors['client.name'] }}</small>
      </label>

      <label class="field" :class="{ 'field--error': fieldErrors['client.phone'] }">
        <span>Telefone</span>
        <input v-model="form.client.phone" inputmode="numeric" placeholder="(33) 98735-1922" @input="maskClientPhone" />
        <small v-if="fieldErrors['client.phone']" class="field__error">{{ fieldErrors['client.phone'] }}</small>
      </label>
    </div>
  </section>
</template>
