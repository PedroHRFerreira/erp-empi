<script lang="ts">
import { ArrowLeft, FileDown, Plus } from '@lucide/vue'
import { computed, defineComponent, reactive, ref } from 'vue'
import type { IExpense, IRealizedExpense, RealizedExpenseOrigin } from '../../../../server/contracts/types'
import { expenseCategories, type ExpenseForm } from '../../../stores/useExpensesStore'
import { formatCurrency } from '../../../utils/format'
import { currencyMaskToCents, formatCentsAsCurrency } from '../../../utils/masks'
import { generateInstallmentSchedule } from '../../../utils/purchaseInstallments'
import PageHeader from '../../molecules/PageHeader/Index.vue'
import PaginationControls from '../../molecules/PaginationControls/Index.vue'
import ExpensesForm from '../../organisms/ExpensesForm/Index.vue'
import ExpensesRegistryTable from '../../organisms/ExpensesRegistryTable/Index.vue'
import ExpensesTable from '../../organisms/ExpensesTable/Index.vue'
import FinancialSummaryGrid from '../../organisms/FinancialSummaryGrid/Index.vue'

function makeExpenseForm(): ExpenseForm {
  return {
    receiptId: null,
    description: '',
    category: '',
    amountCents: 0,
    spentAt: toDateInputValue(new Date()),
    notes: '',
    installments: []
  }
}

export default defineComponent({
  name: 'ExpensesTemplate',
  components: {
    ArrowLeft,
    ExpensesForm,
    ExpensesRegistryTable,
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
    const originOptions: Array<{ value: RealizedExpenseOrigin; label: string }> = [
      { value: 'all', label: 'Todos' },
      { value: 'operational', label: 'Operacionais' },
      { value: 'stock', label: 'Estoque' }
    ]
    const amountInput = ref('')
    const installmentAmountInputs = ref<string[]>([])
    const installmentCount = ref(1)
    const firstDueDate = ref(toDateInputValue(new Date()))
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
        ? 'Cadastre o gasto e programe as parcelas que seguirão para Contas a pagar.'
        : 'Acompanhe somente as saídas efetivamente pagas, separadas por origem.'
    })

    function resetForm() {
      Object.assign(form, makeExpenseForm())
      amountInput.value = ''
      installmentCount.value = 1
      firstDueDate.value = toDateInputValue(new Date())
      generateInstallments()
      expenses.error = ''
      expenses.fieldErrors = {}
    }

    function startCreate() {
      resetForm()
      showForm.value = true
    }

    function generateInstallments() {
      const count = Math.min(48, Math.max(1, Math.trunc(installmentCount.value || 1)))
      installmentCount.value = count
      const schedule = generateInstallmentSchedule(form.amountCents, count, firstDueDate.value)
      form.installments = schedule.map((row, index) => ({
        ...row,
        plannedMethod: form.installments[index]?.plannedMethod || 'boleto'
      }))
      installmentAmountInputs.value = form.installments.map((row) => formatCentsAsCurrency(row.amountCents))
      expenses.clearFieldError('installments')
    }

    function setAmountInput(value: string) {
      amountInput.value = value
      form.amountCents = currencyMaskToCents(value)
      generateInstallments()
    }

    function setInstallmentCount(value: number) {
      installmentCount.value = value
      generateInstallments()
    }

    function setFirstDueDate(value: string) {
      firstDueDate.value = value
      generateInstallments()
    }

    function setInstallmentAmount(index: number, value: string) {
      const installment = form.installments[index]
      if (!installment) return
      installment.amountCents = currencyMaskToCents(value)
      installmentAmountInputs.value[index] = value
    }

    function removeInstallment(index: number) {
      if (form.installments.length <= 1) return
      form.installments.splice(index, 1)
      installmentCount.value = form.installments.length
      generateInstallments()
    }

    function operationalExpense(row: IRealizedExpense) {
      return expenses.expenses.find((expense) => expense.id === (row.expenseId || row.id))
    }

    function startEdit(row: IRealizedExpense) {
      const expense = operationalExpense(row)
      if (!expense || !row.editable || row.origin !== 'operational') return
      Object.assign(form, {
        id: expense.id,
        receiptId: expense.receiptId || null,
        description: expense.description,
        category: expense.category,
        amountCents: expense.amountCents,
        spentAt: toDateInputValue(new Date(expense.spentAt)),
        notes: expense.notes || '',
        installments: (expense.installments || []).map((installment) => ({
          amountCents: installment.amountCents,
          dueDate: installment.dueDate.slice(0, 10),
          plannedMethod: installment.plannedMethod
        }))
      })
      amountInput.value = formatCentsAsCurrency(expense.amountCents)
      installmentCount.value = form.installments.length || 1
      firstDueDate.value = form.installments[0]?.dueDate || expense.spentAt.slice(0, 10)
      installmentAmountInputs.value = form.installments.map((installment) => formatCentsAsCurrency(installment.amountCents))
      if (!form.installments.length) generateInstallments()
      expenses.error = ''
      expenses.fieldErrors = {}
      showForm.value = true
    }

    function startEditExpense(expense: IExpense) {
      const paid = expense.installments?.some((installment) => installment.status === 'paid')
      if (paid) return
      Object.assign(form, {
        id: expense.id,
        receiptId: expense.receiptId || null,
        description: expense.description,
        category: expense.category,
        amountCents: expense.amountCents,
        spentAt: toDateInputValue(new Date(expense.spentAt)),
        notes: expense.notes || '',
        installments: (expense.installments || []).map((installment) => ({ amountCents: installment.amountCents, dueDate: installment.dueDate.slice(0, 10), plannedMethod: installment.plannedMethod }))
      })
      amountInput.value = formatCentsAsCurrency(expense.amountCents)
      installmentCount.value = form.installments.length || 1
      firstDueDate.value = form.installments[0]?.dueDate || expense.spentAt.slice(0, 10)
      installmentAmountInputs.value = form.installments.map((installment) => formatCentsAsCurrency(installment.amountCents))
      if (!form.installments.length) generateInstallments()
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

    async function remove(row: IRealizedExpense) {
      if (!row.editable || row.origin !== 'operational') return
      const confirmed = window.confirm(`Remover o gasto "${row.description}"?`)
      if (!confirmed) return
      await expenses.remove(row.expenseId || row.id)
    }

    async function removeExpense(expense: IExpense) {
      if (expense.installments?.some((installment) => installment.status === 'paid')) return
      const confirmed = window.confirm(`Remover o gasto "${expense.description}"?`)
      if (!confirmed) return
      await expenses.remove(expense.id)
    }

    function applyPeriod() {
      return expenses.load(0, true)
    }

    function selectOrigin(origin: RealizedExpenseOrigin) {
      expenses.origin = origin
      return expenses.load(0, true)
    }

    function previousPage() {
      return expenses.load(expenses.offset - expenses.limit, true)
    }

    function nextPage() {
      return expenses.load(expenses.offset + expenses.limit, true)
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
      firstDueDate,
      formatCurrency,
      generateInstallments,
      installmentAmountInputs,
      installmentCount,
      isEditing,
      nextPage,
      originOptions,
      pageSubtitle,
      pageTitle,
      pages,
      previousPage,
      remove,
      removeInstallment,
      receiptCosts,
      receipts,
      save,
      selectOrigin,
      setAmountInput,
      setFirstDueDate,
      setInstallmentAmount,
      setInstallmentCount,
      showForm,
      startCreate,
      startEdit,
      startEditExpense,
      removeExpense
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
      :first-due-date="firstDueDate"
      :installment-amount-inputs="installmentAmountInputs"
      :installment-count="installmentCount"
      :receipt-options="receipts.receiptOptions"
      :saving="expenses.saving"
      @cancel="cancelForm"
      @clear-field-error="expenses.clearFieldError"
      @generate-installments="generateInstallments"
      @remove-installment="removeInstallment"
      @save="save"
      @update:amount-input="setAmountInput"
      @update:first-due-date="setFirstDueDate"
      @update:installment-amount="setInstallmentAmount"
      @update:installment-count="setInstallmentCount"
    />

    <template v-else>
      <form class="expenses-template__filters panel" @submit.prevent="applyPeriod">
        <fieldset class="expenses-template__origin-filter">
          <legend>Origem do gasto</legend>
          <button
            v-for="option in originOptions"
            :key="option.value"
            class="expenses-template__origin-button"
            :class="{ 'expenses-template__origin-button--active': expenses.origin === option.value }"
            :aria-pressed="expenses.origin === option.value"
            :disabled="expenses.loading"
            type="button"
            @click="selectOrigin(option.value)"
          >
            {{ option.label }}
          </button>
        </fieldset>
        <label class="field">
          <span>Início</span>
          <input v-model="expenses.startDate" type="date" />
        </label>
        <label class="field">
          <span>Fim</span>
          <input v-model="expenses.endDate" type="date" />
        </label>
        <button class="button button--secondary" type="submit" :disabled="expenses.loading">
          {{ expenses.loading ? 'Atualizando…' : 'Aplicar período' }}
        </button>
      </form>

      <FinancialSummaryGrid :summary="expenses.summary" />

      <template v-if="expenses.origin !== 'stock'">
        <section class="expenses-template__section-heading">
          <div>
            <span>Controle de lançamentos</span>
            <h2>Gastos operacionais cadastrados</h2>
          </div>
          <p>Inclui parcelas pendentes. Os valores só entram nos indicadores e no caixa depois da quitação em Contas a pagar.</p>
        </section>
        <ExpensesRegistryTable :expenses="expenses.expenses" @edit="startEditExpense" @remove="removeExpense" />
      </template>

      <section class="expenses-template__section-heading">
        <div>
          <span>Saídas realizadas</span>
          <h2>Histórico de gastos pagos</h2>
        </div>
        <p>Compras de estoque aparecem aqui somente depois da quitação e não podem ser alteradas nesta tela.</p>
      </section>

      <div
        v-if="expenses.loading"
        class="expenses-template__state panel"
        role="status"
        aria-live="polite"
      >
        Carregando saídas realizadas…
      </div>
      <div v-else-if="expenses.error" class="expenses-template__state expenses-template__state--error panel" role="alert">
        <strong>Não foi possível atualizar os gastos.</strong>
        <span>{{ expenses.error }}</span>
        <button class="button button--secondary" type="button" @click="expenses.load(expenses.offset, true)">
          Tentar novamente
        </button>
      </div>
      <ExpensesTable v-else :expenses="expenses.realizedExpenses" @edit="startEdit" @remove="remove" />

      <PaginationControls
        v-if="!expenses.loading && !expenses.error && pages > 1"
        :current-page="currentPage"
        :pages="pages"
        @next="nextPage"
        @previous="previousPage"
      />

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

    </template>
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
