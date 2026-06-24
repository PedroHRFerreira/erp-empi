<script lang="ts">
import { computed, defineComponent, reactive, ref, watch, type PropType } from 'vue'
import type { IStockItem } from '../../../../server/contracts/types'
import { expenseCategories } from '../../../stores/useExpensesStore'
import {
  makeReceiptServiceExpense,
  receiptWizardSteps,
  type ReceiptForm,
  type ReceiptServiceExpenseForm
} from '../../../stores/useReceiptsStore'
import type { IStoreActionResult } from '../../../stores/types'
import { currencyMaskToCents } from '../../../utils/masks'
import ReceiptClientStep from '../ReceiptClientStep/Index.vue'
import ReceiptFinalizeStep from '../ReceiptFinalizeStep/Index.vue'
import ReceiptProductsStep from '../ReceiptProductsStep/Index.vue'
import ReceiptServiceExpensesStep from '../ReceiptServiceExpensesStep/Index.vue'
import ReceiptServicesStep from '../ReceiptServicesStep/Index.vue'
import ReceiptVehicleStep from '../ReceiptVehicleStep/Index.vue'
import ReceiptWizardProgress from '../ReceiptWizardProgress/Index.vue'

export default defineComponent({
  name: 'ReceiptsForm',
  components: {
    ReceiptClientStep,
    ReceiptFinalizeStep,
    ReceiptProductsStep,
    ReceiptServiceExpensesStep,
    ReceiptServicesStep,
    ReceiptVehicleStep,
    ReceiptWizardProgress
  },
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
  emits: ['back-to-list', 'clear-field-error'],
  setup(props, { emit }) {
    const auth = useAuthStore()
    const receipts = useReceiptsStore()

    receipts.resetReceiptWizard(defaultFees())

    const form = computed(() => receipts.receiptDraft)
    const serviceExpense = reactive<ReceiptServiceExpenseForm>(makeReceiptServiceExpense())
    const selectedStockId = ref('')
    const selectedQuantity = ref(1)
    const laborPriceInput = ref('')
    const serviceExpenseAmountInput = ref('')
    const serviceExpenseError = ref('')
    const itemError = ref('')
    const hasAppliedProfileFees = ref(Boolean(auth.user))
    const installmentOptions = Array.from({ length: 12 }, (_, index) => index + 1)

    const isFinalStep = computed(() => receipts.receiptWizardStep === receiptWizardSteps.length - 1)
    const selectedStockItem = computed(() => {
      return props.stockItems.find((item) => item.id === selectedStockId.value) || null
    })
    const availableQuantity = computed(() => {
      if (!selectedStockItem.value) return 0

      return Math.max(selectedStockItem.value.quantity - usedQuantityInForm(selectedStockItem.value.id), 0)
    })
    const laborPriceCents = computed(() => currencyMaskToCents(laborPriceInput.value))
    const productsTotalCents = computed(() => {
      return form.value.items.reduce((total, item) => total + itemTotalCents(item), 0)
    })
    const serviceExpensesTotalCents = computed(() => {
      return form.value.serviceExpenses.reduce((total, expense) => total + expense.amountCents, 0)
    })
    const subtotalCents = computed(() => laborPriceCents.value + productsTotalCents.value + serviceExpensesTotalCents.value)
    const isCardPayment = computed(() => {
      return form.value.paymentMethod === 'credit_card' || form.value.paymentMethod === 'debit_card'
    })
    const activeCardFeePercent = computed(() => {
      return cardFeePercentFor(form.value.paymentMethod, form.value.installments || 1)
    })
    const activeCardFeeLabel = computed(() => {
      if (form.value.paymentMethod === 'credit_card' && (form.value.installments || 1) > 1) {
        return 'Juros de parcelamento'
      }
      return 'Juros da maquininha'
    })
    const cardFeeCents = computed(() => {
      return calculateFeeCents(subtotalCents.value, activeCardFeePercent.value)
    })
    const totalCents = computed(() => subtotalCents.value + cardFeeCents.value)
    const installmentValueCents = computed(() => {
      const installments = form.value.paymentMethod === 'credit_card' ? form.value.installments || 1 : 1
      return Math.ceil(totalCents.value / installments)
    })

    watch(
      () => auth.user?.id,
      () => {
        if (!auth.user || hasAppliedProfileFees.value) return
        receipts.applyReceiptDefaultFees(defaultFees())
        hasAppliedProfileFees.value = true
        syncCardFeePercent()
      }
    )

    watch(
      () => [
        form.value.paymentMethod,
        form.value.installments,
        form.value.machineFeePercent,
        form.value.installmentFeePercent
      ],
      () => syncCardFeePercent(),
      { immediate: true }
    )

    function defaultFees() {
      return {
        machineFeePercent: Number(auth.user?.machineFeePercent || 0),
        installmentFeePercent: Number(auth.user?.installmentFeePercent || 0)
      }
    }

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
      if (paymentMethod === 'debit_card') return Number(form.value.machineFeePercent || 0)
      if (paymentMethod === 'credit_card') {
        return installments > 1 ? Number(form.value.installmentFeePercent || 0) : Number(form.value.machineFeePercent || 0)
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

    function syncCardFeePercent() {
      form.value.cardFeePercent = isCardPayment.value ? activeCardFeePercent.value : 0
    }

    function syncCalculatedFields() {
      syncCardFeePercent()
      form.value.laborPriceCents = laborPriceCents.value
      form.value.priceCents = totalCents.value
      if (form.value.paymentMethod !== 'credit_card') {
        form.value.installments = 1
      }
    }

    function clearFieldError(field: string) {
      emit('clear-field-error', field)
    }

    function clearItemsError() {
      itemError.value = ''
      clearFieldError('items')
    }

    function clearServiceExpensesError() {
      serviceExpenseError.value = ''
      clearFieldError('serviceExpenses')
    }

    function usedQuantityInForm(id: string) {
      return form.value.items
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

      form.value.items.push({ stockItemId: selectedStockId.value, quantity: selectedQuantity.value })
      selectedStockId.value = ''
      selectedQuantity.value = 1
    }

    function removeItem(index: number) {
      form.value.items.splice(index, 1)
      clearItemsError()
    }

    function resetServiceExpenseForm() {
      Object.assign(serviceExpense, makeReceiptServiceExpense())
      serviceExpenseAmountInput.value = ''
      serviceExpenseError.value = ''
    }

    function addServiceExpense() {
      clearServiceExpensesError()
      const amountCents = currencyMaskToCents(serviceExpenseAmountInput.value)

      if (!serviceExpense.description.trim()) {
        serviceExpenseError.value = 'Informe a descrição do gasto.'
        return
      }
      if (!serviceExpense.category.trim()) {
        serviceExpenseError.value = 'Selecione a categoria.'
        return
      }
      if (amountCents <= 0) {
        serviceExpenseError.value = 'Informe um valor maior que zero.'
        return
      }
      if (!serviceExpense.spentAt) {
        serviceExpenseError.value = 'Informe a data do gasto.'
        return
      }

      form.value.serviceExpenses.push({
        description: serviceExpense.description.trim(),
        category: serviceExpense.category,
        amountCents,
        spentAt: serviceExpense.spentAt,
        notes: serviceExpense.notes.trim()
      })
      resetServiceExpenseForm()
    }

    function removeServiceExpense(index: number) {
      form.value.serviceExpenses.splice(index, 1)
      clearServiceExpensesError()
    }

    function resetForm() {
      receipts.resetReceiptWizard(defaultFees())
      selectedStockId.value = ''
      selectedQuantity.value = 1
      laborPriceInput.value = ''
      itemError.value = ''
      resetServiceExpenseForm()
    }

    function updatePaymentMethod() {
      clearFieldError('paymentMethod')
      clearFieldError('installments')
      if (form.value.paymentMethod !== 'credit_card') {
        form.value.installments = 1
      }
      syncCardFeePercent()
    }

    function previousStep() {
      if (receipts.receiptWizardStep === 0) {
        emit('back-to-list')
        return
      }
      receipts.previousReceiptStep()
    }

    function nextStep() {
      syncCalculatedFields()
      if (!receipts.validateReceiptStep()) return
      receipts.nextReceiptStep()
    }

    async function createReceipt() {
      syncCalculatedFields()
      if (!receipts.validateReceiptStep()) return

      const result = await props.onCreate(form.value)

      if (result.status === 'success') {
        resetForm()
      }
    }

    function submitStep() {
      if (isFinalStep.value) {
        return createReceipt()
      }
      nextStep()
    }

    return {
      activeCardFeeLabel,
      addItem,
      addServiceExpense,
      availableQuantity,
      cardFeeCents,
      categories: expenseCategories,
      clearFieldError,
      clearItemsError,
      clearServiceExpensesError,
      createReceipt,
      form,
      installmentOptions,
      installmentValueCents,
      installmentValueFor,
      isFinalStep,
      itemError,
      itemTotalCents,
      itemUnitCents,
      laborPriceCents,
      laborPriceInput,
      nextStep,
      previousStep,
      productsTotalCents,
      receiptWizardSteps,
      receipts,
      removeItem,
      removeServiceExpense,
      selectedQuantity,
      selectedStockId,
      selectedStockItem,
      serviceExpense,
      serviceExpenseAmountInput,
      serviceExpenseError,
      serviceExpensesTotalCents,
      stockName,
      submitStep,
      subtotalCents,
      syncCardFeePercent,
      totalCents,
      updatePaymentMethod
    }
  }
})
</script>

<template>
  <form class="receipts-form panel" novalidate @submit.prevent="submitStep">
    <ReceiptWizardProgress :active-index="receipts.receiptWizardStep" :steps="receiptWizardSteps" />

    <ReceiptClientStep
      v-if="receipts.receiptWizardStep === 0"
      :field-errors="fieldErrors"
      :form="form"
      @clear-field-error="clearFieldError"
    />

    <ReceiptVehicleStep
      v-else-if="receipts.receiptWizardStep === 1"
      :field-errors="fieldErrors"
      :form="form"
      @clear-field-error="clearFieldError"
    />

    <ReceiptServicesStep
      v-else-if="receipts.receiptWizardStep === 2"
      v-model:labor-price-input="laborPriceInput"
      :active-card-fee-label="activeCardFeeLabel"
      :card-fee-cents="cardFeeCents"
      :field-errors="fieldErrors"
      :form="form"
      :installment-options="installmentOptions"
      :installment-value-for="installmentValueFor"
      :subtotal-cents="subtotalCents"
      :total-cents="totalCents"
      @clear-field-error="clearFieldError"
      @payment-change="updatePaymentMethod"
      @sync-card-fee="syncCardFeePercent"
    />

    <ReceiptProductsStep
      v-else-if="receipts.receiptWizardStep === 3"
      v-model:selected-quantity="selectedQuantity"
      v-model:selected-stock-id="selectedStockId"
      :available-quantity="availableQuantity"
      :field-errors="fieldErrors"
      :form="form"
      :item-error="itemError"
      :item-total-cents="itemTotalCents"
      :item-unit-cents="itemUnitCents"
      :products-total-cents="productsTotalCents"
      :selected-stock-item="selectedStockItem"
      :stock-items="stockItems"
      :stock-name="stockName"
      @add-item="addItem"
      @clear-items-error="clearItemsError"
      @remove-item="removeItem"
    />

    <ReceiptServiceExpensesStep
      v-else-if="receipts.receiptWizardStep === 4"
      v-model:service-expense-amount-input="serviceExpenseAmountInput"
      :categories="categories"
      :field-errors="fieldErrors"
      :form="form"
      :service-expense="serviceExpense"
      :service-expense-error="serviceExpenseError"
      :service-expenses-total-cents="serviceExpensesTotalCents"
      @add-service-expense="addServiceExpense"
      @clear-service-expenses-error="clearServiceExpensesError"
      @remove-service-expense="removeServiceExpense"
    />

    <ReceiptFinalizeStep
      v-else
      :active-card-fee-label="activeCardFeeLabel"
      :card-fee-cents="cardFeeCents"
      :error="error"
      :form="form"
      :installment-value-cents="installmentValueCents"
      :labor-price-cents="laborPriceCents"
      :products-total-cents="productsTotalCents"
      :service-expenses-total-cents="serviceExpensesTotalCents"
      :subtotal-cents="subtotalCents"
      :total-cents="totalCents"
      @clear-field-error="clearFieldError"
    />

    <footer class="receipt-wizard-actions">
      <button class="button button--secondary" type="button" @click="previousStep">
        {{ receipts.receiptWizardStep === 0 ? 'Voltar' : 'Anterior' }}
      </button>
      <button v-if="!isFinalStep" class="button button--primary" type="button" @click="nextStep">Avançar</button>
      <button v-else class="button button--primary" type="submit">Salvar recibo</button>
    </footer>
  </form>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
