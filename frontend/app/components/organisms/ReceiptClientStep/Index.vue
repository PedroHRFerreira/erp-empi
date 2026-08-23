<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import type { IUser } from '../../../../server/contracts/types'
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
    },
    suggestions: { type: Array as PropType<IUser[]>, default: () => [] },
    suggestionLoading: { type: Boolean, default: false },
    selectedClientId: { type: String, default: '' },
    suggestionNotice: { type: String, default: '' }
  },
  emits: ['clear-field-error', 'client-input', 'clear-selection', 'select-client'],
  setup(props, { emit }) {
    function clearFieldError(field: string) {
      emit('clear-field-error', field)
    }

    function maskClientPhone() {
      clearFieldError('client.phone')
      props.form.client.phone = maskPhone(props.form.client.phone)
      emit('client-input')
    }

    function clientNameInput() { clearFieldError('client.name'); emit('client-input') }

    return {
      clientNameInput,
      clearFieldError,
      maskPhone,
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
        <input v-model="form.client.name" required autocomplete="off" placeholder="Nome do cliente" role="combobox" :aria-expanded="suggestions.length > 0" aria-controls="receipt-client-suggestions" @input="clientNameInput" />
        <small v-if="fieldErrors['client.name']" class="field__error">{{ fieldErrors['client.name'] }}</small>
      </label>

      <label class="field" :class="{ 'field--error': fieldErrors['client.phone'] }">
        <span>Telefone</span>
        <input v-model="form.client.phone" inputmode="numeric" placeholder="(33) 98735-1922" @input="maskClientPhone" />
        <small v-if="fieldErrors['client.phone']" class="field__error">{{ fieldErrors['client.phone'] }}</small>
      </label>
    </div>

    <div v-if="suggestionLoading" class="client-suggestion-state">Buscando dados do cliente...</div>
    <ul v-else-if="suggestions.length && !selectedClientId" id="receipt-client-suggestions" class="client-suggestions" role="listbox">
      <li v-for="client in suggestions" :key="client.id"><button type="button" role="option" @click="$emit('select-client', client)"><strong>{{ client.name }}</strong><span>{{ maskPhone(client.phone) }}</span></button></li>
    </ul>
    <div v-if="selectedClientId" class="client-suggestion-applied">
      <span>{{ suggestionNotice || 'Dados sugeridos a partir do recibo mais recente.' }}</span>
      <button type="button" @click="$emit('clear-selection')">Limpar seleção</button>
    </div>
  </section>
</template>
