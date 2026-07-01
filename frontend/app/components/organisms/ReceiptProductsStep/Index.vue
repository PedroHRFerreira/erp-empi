<script lang="ts">
import { computed, defineComponent, type PropType } from 'vue'
import type { IStockItem } from '../../../../server/contracts/types'
import type { ReceiptForm } from '../../../stores/useReceiptsStore'
import { formatCurrency } from '../../../utils/format'

export default defineComponent({
  name: 'ReceiptProductsStep',
  props: {
    availableQuantity: {
      type: Number,
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
    itemError: {
      type: String,
      default: ''
    },
    itemTotalCents: {
      type: Function as PropType<(item: { stockItemId: string; quantity: number }) => number>,
      required: true
    },
    itemUnitCents: {
      type: Function as PropType<(id: string) => number>,
      required: true
    },
    productsTotalCents: {
      type: Number,
      required: true
    },
    selectedQuantity: {
      type: Number,
      required: true
    },
    selectedStockId: {
      type: String,
      required: true
    },
    selectedStockItem: {
      type: Object as PropType<IStockItem | null>,
      default: null
    },
    stepLabel: {
      type: String,
      default: 'Etapa 4'
    },
    stockItems: {
      type: Array as PropType<IStockItem[]>,
      required: true
    },
    stockName: {
      type: Function as PropType<(id: string) => string>,
      required: true
    }
  },
  emits: [
    'add-item',
    'clear-items-error',
    'remove-item',
    'update:selected-quantity',
    'update:selected-stock-id'
  ],
  setup(props, { emit }) {
    const selectedStock = computed({
      get: () => props.selectedStockId,
      set: (value: string) => emit('update:selected-stock-id', value)
    })
    const quantity = computed({
      get: () => props.selectedQuantity,
      set: (value: number) => emit('update:selected-quantity', Number(value || 1))
    })

    function clearItemsError() {
      emit('clear-items-error')
    }

    return {
      clearItemsError,
      formatCurrency,
      quantity,
      selectedStock
    }
  }
})
</script>

<template>
  <section class="receipt-step">
    <header class="receipt-step__header">
      <span>{{ stepLabel }}</span>
      <h2>Produtos usados</h2>
    </header>

    <div class="receipt-items" :class="{ 'receipt-items--error': fieldErrors.items }">
      <div class="receipt-items__row">
        <label class="field">
          <span>Produto usado</span>
          <select v-model="selectedStock" @change="clearItemsError">
            <option value="">Selecione</option>
            <option v-for="item in stockItems" :key="item.id" :disabled="item.quantity <= 0" :value="item.id">
              {{ item.name }} - disponível {{ item.quantity }}
            </option>
          </select>
        </label>

        <label class="field">
          <span>Quantidade</span>
          <input v-model.number="quantity" type="number" min="1" :max="availableQuantity || 1" @input="clearItemsError" />
        </label>

        <button class="button button--secondary" type="button" @click="$emit('add-item')">Adicionar item</button>
      </div>

      <small v-if="selectedStockItem" class="receipt-items__hint">
        Disponível para este recibo: {{ availableQuantity }}
      </small>

      <ul v-if="form.items.length">
        <li v-for="(item, index) in form.items" :key="`${item.stockItemId}-${index}`">
          <span>
            {{ stockName(item.stockItemId) }} x {{ item.quantity }}
            <small>{{ formatCurrency(itemUnitCents(item.stockItemId)) }} un. / {{ formatCurrency(itemTotalCents(item)) }} total</small>
          </span>
          <button type="button" @click="$emit('remove-item', index)">Remover</button>
        </li>
      </ul>

      <p v-else class="receipt-step__empty">Nenhum produto usado neste recibo.</p>

      <strong class="receipt-items__total">Total dos produtos: {{ formatCurrency(productsTotalCents) }}</strong>
      <small v-if="itemError" class="field__error">{{ itemError }}</small>
      <small v-if="fieldErrors.items" class="field__error">{{ fieldErrors.items }}</small>
    </div>
  </section>
</template>
