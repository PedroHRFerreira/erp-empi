<script lang="ts">
import { computed, defineComponent, reactive, ref, type PropType } from 'vue'
import type { IStockItem } from '../../../../server/contracts/types'
import type { ReceiptForm } from '../../../stores/useReceiptsStore'
import type { IStoreActionResult } from '../../../stores/types'
import { formatCurrency } from '../../../utils/format'
import { currencyMaskToCents, maskCpf, maskCurrency, maskPhone, maskVehiclePlate } from '../../../utils/masks'

function makeReceiptForm(): ReceiptForm {
  return {
    client: { name: '', cpf: '', phone: '', email: '' },
    vehicleModel: '',
    vehicleYear: new Date().getFullYear(),
    vehiclePlate: '',
    services: '',
    laborPriceCents: 0,
    priceCents: 0,
    paymentMethod: 'cash',
    installments: 1,
    notes: '',
    items: []
  }
}

export default defineComponent({
  name: 'ReceiptsForm',
  props: {
    stockItems: {
      type: Array as PropType<IStockItem[]>,
      required: true
    },
    fieldErrors: {
      type: Object as PropType<Record<string, string>>,
      required: true
    },
    error: {
      type: String,
      default: ''
    },
    onCreate: {
      type: Function as PropType<(form: ReceiptForm) => Promise<IStoreActionResult>>,
      required: true
    }
  },
  emits: ['clear-field-error'],
  setup(props, { emit }) {
    const auth = useAuthStore()
    const form = reactive<ReceiptForm>(makeReceiptForm())
    const selectedStockId = ref('')
    const selectedQuantity = ref(1)
    const laborPriceInput = ref('')
    const itemError = ref('')
    const installmentOptions = Array.from({ length: 12 }, (_, index) => index + 1)

    const selectedStockItem = computed(() => {
      return props.stockItems.find((item) => item.id === selectedStockId.value) || null
    })

    const availableQuantity = computed(() => {
      if (!selectedStockItem.value) return 0

      return Math.max(selectedStockItem.value.quantity - usedQuantityInForm(selectedStockItem.value.id), 0)
    })

    const laborPriceCents = computed(() => currencyMaskToCents(laborPriceInput.value))
    const productsTotalCents = computed(() => {
      return form.items.reduce((total, item) => total + itemTotalCents(item), 0)
    })
    const subtotalCents = computed(() => laborPriceCents.value + productsTotalCents.value)
    const machineFeePercent = computed(() => Number(auth.user?.machineFeePercent || 0))
    const installmentFeePercent = computed(() => Number(auth.user?.installmentFeePercent || 0))
    const isCardPayment = computed(() => {
      return form.paymentMethod === 'credit_card' || form.paymentMethod === 'debit_card'
    })
    const activeCardFeePercent = computed(() => {
      return cardFeePercentFor(form.paymentMethod, form.installments || 1)
    })
    const activeCardFeeLabel = computed(() => {
      if (form.paymentMethod === 'credit_card' && (form.installments || 1) > 1) {
        return 'Juros de parcelamento'
      }
      return 'Juros da maquininha'
    })
    const cardFeeCents = computed(() => {
      return calculateFeeCents(subtotalCents.value, activeCardFeePercent.value)
    })
    const totalCents = computed(() => subtotalCents.value + cardFeeCents.value)
    const installmentValueCents = computed(() => {
      const installments = form.paymentMethod === 'credit_card' ? form.installments || 1 : 1
      return Math.ceil(totalCents.value / installments)
    })

    function stockItemById(id: string) {
      return props.stockItems.find((item) => item.id === id) || null
    }

    function stockName(id: string) {
      return stockItemById(id)?.name || 'Produto'
    }

    function itemTotalCents(item: { stockItemId: string; quantity: number }) {
      return (stockItemById(item.stockItemId)?.resalePriceCents || 0) * item.quantity
    }

    function itemUnitCents(id: string) {
      return stockItemById(id)?.resalePriceCents || 0
    }

    function cardFeePercentFor(paymentMethod: ReceiptForm['paymentMethod'], installments: number) {
      if (paymentMethod === 'debit_card') return machineFeePercent.value
      if (paymentMethod === 'credit_card') {
        return installments > 1 ? installmentFeePercent.value : machineFeePercent.value
      }
      return 0
    }

    function calculateFeeCents(value: number, percent: number) {
      if (percent <= 0) return 0
      return Math.trunc(value * (percent / 100))
    }

    function totalCentsForInstallment(installments: number) {
      const percent = cardFeePercentFor('credit_card', installments)
      return subtotalCents.value + calculateFeeCents(subtotalCents.value, percent)
    }

    function installmentValueFor(installments: number) {
      return Math.ceil(totalCentsForInstallment(installments) / installments)
    }

    function clearFieldError(field: string) {
      emit('clear-field-error', field)
    }

    function clearItemsError() {
      itemError.value = ''
      clearFieldError('items')
    }

    function usedQuantityInForm(id: string) {
      return form.items
        .filter((item) => item.stockItemId === id)
        .reduce((total, item) => total + item.quantity, 0)
    }

    function addItem() {
      itemError.value = ''
      clearFieldError('items')

      if (!selectedStockItem.value) {
        itemError.value = 'Selecione um produto.'
        return
      }
      if (selectedStockItem.value.quantity <= 0) {
        itemError.value = 'Este produto não possui estoque disponível.'
        return
      }
      if (selectedQuantity.value <= 0) {
        itemError.value = 'Informe uma quantidade maior que zero.'
        return
      }
      if (selectedQuantity.value > availableQuantity.value) {
        itemError.value = `Quantidade máxima disponível: ${availableQuantity.value}.`
        return
      }

      form.items.push({ stockItemId: selectedStockId.value, quantity: selectedQuantity.value })
      selectedStockId.value = ''
      selectedQuantity.value = 1
    }

    function removeItem(index: number) {
      form.items.splice(index, 1)
      itemError.value = ''
      clearFieldError('items')
    }

    function resetForm() {
      Object.assign(form, makeReceiptForm())
      laborPriceInput.value = ''
      itemError.value = ''
    }

    async function createReceipt() {
      form.laborPriceCents = laborPriceCents.value
      form.priceCents = totalCents.value
      if (form.paymentMethod !== 'credit_card') {
        form.installments = 1
      }

      const result = await props.onCreate(form)

      if (result.status === 'success') {
        resetForm()
      }
    }

    function maskClientCpf() {
      clearFieldError('client.cpf')
      clearFieldError('client.phone')
      form.client.cpf = maskCpf(form.client.cpf)
    }

    function maskClientPhone() {
      clearFieldError('client.phone')
      form.client.phone = maskPhone(form.client.phone)
    }

    function maskPlate() {
      clearFieldError('vehiclePlate')
      form.vehiclePlate = maskVehiclePlate(form.vehiclePlate)
    }

    function maskLaborPrice() {
      clearFieldError('laborPriceCents')
      laborPriceInput.value = maskCurrency(laborPriceInput.value)
    }

    function updatePaymentMethod() {
      clearFieldError('paymentMethod')
      clearFieldError('installments')
      if (form.paymentMethod !== 'credit_card') {
        form.installments = 1
      }
    }

    return {
      addItem,
      activeCardFeeLabel,
      activeCardFeePercent,
      availableQuantity,
      cardFeeCents,
      clearFieldError,
      clearItemsError,
      createReceipt,
      form,
      formatCurrency,
      installmentOptions,
      installmentValueCents,
      installmentValueFor,
      isCardPayment,
      itemTotalCents,
      itemUnitCents,
      itemError,
      laborPriceCents,
      laborPriceInput,
      maskClientCpf,
      maskClientPhone,
      maskLaborPrice,
      maskPlate,
      machineFeePercent,
      productsTotalCents,
      removeItem,
      selectedQuantity,
      selectedStockId,
      selectedStockItem,
      stockName,
      subtotalCents,
      totalCents,
      updatePaymentMethod
    }
  }
})
</script>

<template>
  <form class="receipts-form panel" novalidate @submit.prevent="createReceipt">
    <label class="field" :class="{ 'field--error': fieldErrors['client.name'] }">
      <span>Cliente</span>
      <input v-model="form.client.name" required placeholder="Nome do cliente" @input="clearFieldError('client.name')" />
      <small v-if="fieldErrors['client.name']" class="field__error">{{ fieldErrors['client.name'] }}</small>
    </label>

    <label class="field" :class="{ 'field--error': fieldErrors['client.cpf'] }">
      <span>CPF opcional</span>
      <input v-model="form.client.cpf" inputmode="numeric" placeholder="000.000.000-00" @input="maskClientCpf" />
      <small v-if="fieldErrors['client.cpf']" class="field__error">{{ fieldErrors['client.cpf'] }}</small>
    </label>

    <label class="field" :class="{ 'field--error': fieldErrors['client.phone'] }">
      <span>Telefone</span>
      <input v-model="form.client.phone" inputmode="numeric" placeholder="(33) 98735-1922" @input="maskClientPhone" />
      <small v-if="fieldErrors['client.phone']" class="field__error">{{ fieldErrors['client.phone'] }}</small>
    </label>

    <label class="field">
      <span>E-mail opcional</span>
      <input v-model="form.client.email" type="email" placeholder="cliente@email.com" @input="clearFieldError('client.email')" />
    </label>

    <label class="field" :class="{ 'field--error': fieldErrors.vehicleModel }">
      <span>Veículo</span>
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

    <label class="field" :class="{ 'field--error': fieldErrors.laborPriceCents }">
      <span>Mão de obra</span>
      <input v-model="laborPriceInput" required inputmode="decimal" placeholder="R$ 350,00" @input="maskLaborPrice" />
      <small v-if="fieldErrors.laborPriceCents" class="field__error">{{ fieldErrors.laborPriceCents }}</small>
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

    <label class="field" :class="{ 'field--error': fieldErrors.installments }">
      <span>Parcelas</span>
      <select v-model.number="form.installments" :disabled="form.paymentMethod !== 'credit_card'" @change="clearFieldError('installments')">
        <option v-for="installment in installmentOptions" :key="installment" :value="installment">
          {{ installment }}x de {{ formatCurrency(installmentValueFor(installment)) }}
        </option>
      </select>
      <small v-if="fieldErrors.installments" class="field__error">{{ fieldErrors.installments }}</small>
    </label>

    <label class="field receipts-form__wide" :class="{ 'field--error': fieldErrors.services }">
      <span>Serviços</span>
      <textarea v-model="form.services" required placeholder="Troca de óleo, diagnóstico elétrico..." @input="clearFieldError('services')" />
      <small v-if="fieldErrors.services" class="field__error">{{ fieldErrors.services }}</small>
    </label>

    <div class="receipts-form__wide receipt-items" :class="{ 'receipt-items--error': fieldErrors.items }">
      <div class="receipt-items__row">
        <label class="field">
          <span>Produto usado</span>
          <select v-model="selectedStockId" @change="clearItemsError">
            <option value="">Selecione</option>
            <option v-for="item in stockItems" :key="item.id" :disabled="item.quantity <= 0" :value="item.id">
              {{ item.name }} - disponível {{ item.quantity }}
            </option>
          </select>
        </label>

        <label class="field">
          <span>Quantidade</span>
          <input v-model.number="selectedQuantity" type="number" min="1" :max="availableQuantity || 1" @input="clearItemsError" />
        </label>

        <button class="button button--secondary" type="button" @click="addItem">Adicionar item</button>
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
          <button type="button" @click="removeItem(index)">Remover</button>
        </li>
      </ul>

      <small v-if="itemError" class="field__error">{{ itemError }}</small>
      <small v-if="fieldErrors.items" class="field__error">{{ fieldErrors.items }}</small>
    </div>

    <label class="field receipts-form__wide">
      <span>Observações</span>
      <textarea v-model="form.notes" placeholder="Informações adicionais" @input="clearFieldError('notes')" />
    </label>

    <section class="receipt-summary receipts-form__wide" aria-label="Resumo do recibo">
      <div>
        <span>Mão de obra</span>
        <strong>{{ formatCurrency(laborPriceCents) }}</strong>
      </div>
      <div>
        <span>Produtos utilizados</span>
        <strong>{{ formatCurrency(productsTotalCents) }}</strong>
      </div>
      <div>
        <span>Subtotal</span>
        <strong>{{ formatCurrency(subtotalCents) }}</strong>
      </div>
      <div v-if="isCardPayment">
        <span>{{ activeCardFeeLabel }} ({{ activeCardFeePercent }}%)</span>
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

    <p v-if="error" class="form-error receipts-form__wide">{{ error }}</p>
    <button class="button button--primary receipts-form__wide" type="submit">Salvar recibo</button>
  </form>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
