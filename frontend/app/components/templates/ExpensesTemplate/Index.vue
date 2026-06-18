<script lang="ts">
import { ArrowLeft, FileDown, Plus } from '@lucide/vue'
import { computed, defineComponent, reactive, ref } from 'vue'
import type { IExpense } from '../../../../server/contracts/types'
import { expenseCategories, type ExpenseForm } from '../../../stores/useExpensesStore'
import { formatCurrency } from '../../../utils/format'
import { currencyMaskToCents, formatCentsAsCurrency } from '../../../utils/masks'
import PageHeader from '../../molecules/PageHeader/Index.vue'
import PaginationControls from '../../molecules/PaginationControls/Index.vue'
import ExpensesForm from '../../organisms/ExpensesForm/Index.vue'
import ExpensesTable from '../../organisms/ExpensesTable/Index.vue'
import FinancialSummaryGrid from '../../organisms/FinancialSummaryGrid/Index.vue'

function makeExpenseForm(): ExpenseForm {
  return {
    receiptId: null,
    description: '',
    category: '',
    amountCents: 0,
    spentAt: toDateInputValue(new Date()),
    notes: ''
  }
}

export default defineComponent({
  name: 'ExpensesTemplate',
  components: {
    ArrowLeft,
    ExpensesForm,
    ExpensesTable,
    FileDown,
    FinancialSummaryGrid,
    PageHeader,
    PaginationControls,
    Plus
  },
  setup() {
    const expenses = useExpensesStore()
    const receipts = useReceiptsStore()
    const showForm = ref(false)
    const amountInput = ref('')
    const form = reactive<ExpenseForm>(makeExpenseForm())
    const pages = computed(() => Math.ceil(expenses.total / expenses.limit))
    const currentPage = computed(() => Math.floor(expenses.offset / expenses.limit) + 1)
    const categoryTotals = computed(() => expenses.summary?.expensesByCategory || [])
    const receiptCosts = computed(() => expenses.summary?.receiptCosts || [])
    const isEditing = computed(() => Boolean(form.id))
    const pageTitle = computed(() => {
      return showForm.value ? (isEditing.value ? 'Editar gasto' : 'Adicionar gasto') : 'Gastos'
    })
    const pageSubtitle = computed(() => {
      return showForm.value
        ? 'Preencha somente as informações do gasto para salvar no controle financeiro.'
        : 'Controle despesas, lucro e situação financeira da oficina.'
    })

    function resetForm() {
      Object.assign(form, makeExpenseForm())
      amountInput.value = ''
      expenses.error = ''
      expenses.fieldErrors = {}
    }

    function startCreate() {
      resetForm()
      showForm.value = true
    }

    function startEdit(expense: IExpense) {
      Object.assign(form, {
        id: expense.id,
        receiptId: expense.receiptId || null,
        description: expense.description,
        category: expense.category,
        amountCents: expense.amountCents,
        spentAt: toDateInputValue(new Date(expense.spentAt)),
        notes: expense.notes || ''
      })
      amountInput.value = formatCentsAsCurrency(expense.amountCents)
      expenses.error = ''
      expenses.fieldErrors = {}
      showForm.value = true
    }

    function cancelForm() {
      resetForm()
      showForm.value = false
    }

    async function save() {
      form.amountCents = currencyMaskToCents(amountInput.value)
      const result = await expenses.save({ ...form })

      if (result.status === 'success') {
        cancelForm()
      }
    }

    async function remove(expense: IExpense) {
      const confirmed = window.confirm(`Remover o gasto "${expense.description}"?`)
      if (!confirmed) return
      await expenses.remove(expense.id)
    }

    function applyPeriod() {
      return expenses.load(0)
    }

    function previousPage() {
      return expenses.load(expenses.offset - expenses.limit)
    }

    function nextPage() {
      return expenses.load(expenses.offset + expenses.limit)
    }

    return {
      amountInput,
      applyPeriod,
      cancelForm,
      categories: expenseCategories,
      categoryTotals,
      currentPage,
      expenses,
      form,
      formatCurrency,
      isEditing,
      nextPage,
      pageSubtitle,
      pageTitle,
      pages,
      previousPage,
      remove,
      receiptCosts,
      receipts,
      save,
      showForm,
      startCreate,
      startEdit
    }
  }
})

function toDateInputValue(date: Date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}
</script>

<template>
  <section class="page expenses-template">
    <PageHeader :title="pageTitle" :subtitle="pageSubtitle">
      <template #actions>
        <div v-if="!showForm" class="expenses-template__header-actions">
          <button class="button button--secondary" type="button" @click="expenses.exportPdf">
            <FileDown :size="18" />
            PDF
          </button>
          <button class="button button--primary" type="button" @click="startCreate">
            <Plus :size="18" />
            Adicionar
          </button>
        </div>
        <button v-else class="button button--secondary" type="button" @click="cancelForm">
          <ArrowLeft :size="18" />
          Voltar
        </button>
      </template>
    </PageHeader>

    <ExpensesForm
      v-if="showForm"
      :amount-input="amountInput"
      :categories="categories"
      :error="expenses.error"
      :field-errors="expenses.fieldErrors"
      :form="form"
      :receipt-options="receipts.receiptOptions"
      :saving="expenses.saving"
      @cancel="cancelForm"
      @clear-field-error="expenses.clearFieldError"
      @save="save"
      @update:amount-input="(value) => (amountInput = value)"
    />

    <template v-else>
      <form class="expenses-template__filters panel" @submit.prevent="applyPeriod">
        <label class="field">
          <span>Início</span>
          <input v-model="expenses.startDate" type="date" />
        </label>
        <label class="field">
          <span>Fim</span>
          <input v-model="expenses.endDate" type="date" />
        </label>
        <button class="button button--secondary" type="submit">Aplicar período</button>
      </form>

      <FinancialSummaryGrid :summary="expenses.summary" />

      <section class="expenses-template__categories panel">
        <header>
          <span>Categorias</span>
          <strong>Distribuição dos gastos</strong>
        </header>
        <div v-if="categoryTotals.length" class="expenses-template__category-list">
          <div v-for="category in categoryTotals" :key="category.category" class="expenses-template__category-row">
            <span>{{ category.category }}</span>
            <strong>{{ formatCurrency(category.amountCents) }}</strong>
            <small>{{ category.count }} lançamento{{ category.count === 1 ? '' : 's' }}</small>
          </div>
        </div>
        <p v-else>Nenhum gasto registrado no período.</p>
      </section>

      <section class="expenses-template__receipt-costs panel">
        <header>
          <span>Recibos</span>
          <strong>Maiores custos internos</strong>
        </header>
        <div v-if="receiptCosts.length" class="expenses-template__receipt-cost-list">
          <div v-for="receipt in receiptCosts" :key="receipt.receiptId" class="expenses-template__receipt-cost-row">
            <span>{{ receipt.clientName }}</span>
            <strong>{{ formatCurrency(receipt.totalCostCents) }}</strong>
            <small>
              {{ receipt.vehicleModel }} {{ receipt.vehiclePlate }} /
              Gastos {{ formatCurrency(receipt.serviceExpensesCents) }} /
              Produtos {{ formatCurrency(receipt.productCostCents) }}
            </small>
          </div>
        </div>
        <p v-else>Nenhum custo vinculado a recibos no período.</p>
      </section>

      <ExpensesTable :expenses="expenses.expenses" @edit="startEdit" @remove="remove" />

      <PaginationControls :current-page="currentPage" :pages="pages" @next="nextPage" @previous="previousPage" />
    </template>
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
