<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import type { ReceiptForm } from '../../../stores/useReceiptsStore'
import { maskVehiclePlate } from '../../../utils/masks'

export default defineComponent({
  name: 'ReceiptVehicleStep',
  props: {
    fieldErrors: {
      type: Object as PropType<Record<string, string>>,
      required: true
    },
    form: {
      type: Object as PropType<ReceiptForm>,
      required: true
    }
  },
  emits: ['clear-field-error'],
  setup(props, { emit }) {
    function clearFieldError(field: string) {
      emit('clear-field-error', field)
    }

    function maskPlate() {
      clearFieldError('vehiclePlate')
      props.form.vehiclePlate = maskVehiclePlate(props.form.vehiclePlate)
    }

    return {
      clearFieldError,
      maskPlate
    }
  }
})
</script>

<template>
  <section class="receipt-step">
    <header class="receipt-step__header">
      <span>Etapa 2</span>
      <h2>Informações do veículo</h2>
    </header>

    <div class="receipt-step__grid receipt-step__grid--three">
      <label class="field" :class="{ 'field--error': fieldErrors.vehicleModel }">
        <span>Modelo do veículo</span>
        <input v-model="form.vehicleModel" required placeholder="Modelo" @input="clearFieldError('vehicleModel')" />
        <small v-if="fieldErrors.vehicleModel" class="field__error">{{ fieldErrors.vehicleModel }}</small>
      </label>

      <label class="field" :class="{ 'field--error': fieldErrors.vehicleYear }">
        <span>Ano</span>
        <input v-model.number="form.vehicleYear" required type="number" min="1950" @input="clearFieldError('vehicleYear')" />
        <small v-if="fieldErrors.vehicleYear" class="field__error">{{ fieldErrors.vehicleYear }}</small>
      </label>

      <label class="field" :class="{ 'field--error': fieldErrors.vehiclePlate }">
        <span>Placa</span>
        <input v-model="form.vehiclePlate" required placeholder="ABC1D23" @input="maskPlate" />
        <small v-if="fieldErrors.vehiclePlate" class="field__error">{{ fieldErrors.vehiclePlate }}</small>
      </label>
    </div>
  </section>
</template>
