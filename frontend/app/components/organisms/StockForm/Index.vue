<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import type { StockForm as IStockForm } from '../../../stores/useStockStore'
import { maskCurrency } from '../../../utils/masks'

export default defineComponent({
  name: 'StockForm',
  props: {
    form: {
      type: Object as PropType<IStockForm>,
      required: true
    },
    costInput: {
      type: String,
      required: true
    },
    fieldErrors: {
      type: Object as PropType<Record<string, string>>,
      required: true
    },
    error: {
      type: String,
      default: ''
    }
  },
  emits: ['clear-field-error', 'save', 'update:cost-input'],
  setup(_, { emit }) {
    function clearFieldError(field: string) {
      emit('clear-field-error', field)
    }

    function save() {
      emit('save')
    }

    function updateCostInput(event: Event) {
      clearFieldError('costCents')
      const input = event.target as HTMLInputElement
      emit('update:cost-input', maskCurrency(input.value))
    }

    return {
      clearFieldError,
      save,
      updateCostInput
    }
  }
})
</script>

<template>
  <form class="stock-form panel" novalidate @submit.prevent="save">
    <label class="field" :class="{ 'field--error': fieldErrors.name }">
      <span>Produto</span>
      <input v-model="form.name" required placeholder="Óleo, filtro, pastilha..." @input="clearFieldError('name')" />
      <small v-if="fieldErrors.name" class="field__error">{{ fieldErrors.name }}</small>
    </label>

    <label class="field" :class="{ 'field--error': fieldErrors.costCents }">
      <span>Custo</span>
      <input :value="costInput" required inputmode="numeric" placeholder="R$ 120,00" @input="updateCostInput" />
      <small v-if="fieldErrors.costCents" class="field__error">{{ fieldErrors.costCents }}</small>
    </label>

    <label class="field" :class="{ 'field--error': fieldErrors.markupPercent }">
      <span>Margem de revenda (%)</span>
      <input v-model.number="form.markupPercent" required type="number" min="0" step="0.1" @input="clearFieldError('markupPercent')" />
      <small v-if="fieldErrors.markupPercent" class="field__error">{{ fieldErrors.markupPercent }}</small>
    </label>

    <label class="field" :class="{ 'field--error': fieldErrors.quantity }">
      <span>Quantidade</span>
      <input v-model.number="form.quantity" required type="number" min="0" @input="clearFieldError('quantity')" />
      <small v-if="fieldErrors.quantity" class="field__error">{{ fieldErrors.quantity }}</small>
    </label>

    <label class="field stock-form__wide">
      <span>Descrição</span>
      <textarea v-model="form.description" placeholder="Detalhes do produto" @input="clearFieldError('description')" />
    </label>

    <p v-if="error" class="form-error stock-form__wide">{{ error }}</p>
    <button class="button button--primary stock-form__wide" type="submit">Salvar produto</button>
  </form>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
