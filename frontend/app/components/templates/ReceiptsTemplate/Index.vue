<script lang="ts">
import { Plus, Zap } from '@lucide/vue'
import { computed, defineComponent, reactive, ref } from 'vue'
import type { IReceipt } from '../../../../server/contracts/types'
import { emptyReceiptFilters, type ReceiptFilters } from '../../../stores/useReceiptsStore'
import EmptyState from '../../molecules/EmptyState/Index.vue'
import PageHeader from '../../molecules/PageHeader/Index.vue'
import PaginationControls from '../../molecules/PaginationControls/Index.vue'
import ReceiptsTable from '../../organisms/ReceiptsTable/Index.vue'
import { prepareReceiptInvoiceIssue, printReceiptDocument } from '../../../utils/print'
import { receiptClientName } from '../../../utils/receiptDisplay'

export default defineComponent({
  name: 'ReceiptsTemplate',
  components: {
    EmptyState,
    PageHeader,
    PaginationControls,
    Plus,
    Zap,
    ReceiptsTable
  },
  setup() {
    const router = useRouter()
    const receipts = useReceiptsStore()
    const auth = useAuthStore()
    const feedback = useSystemFeedback()
    const pages = computed(() => Math.ceil(receipts.total / receipts.limit))
    const currentPage = computed(() => Math.floor(receipts.offset / receipts.limit) + 1)
    const draft = reactive<ReceiptFilters>({ ...receipts.filters })
    const filterError = ref('')

    async function applyFilters() {
      filterError.value = ''
      if (draft.startDate && draft.endDate && draft.startDate > draft.endDate) {
        filterError.value = 'A data final deve ser igual ou posterior à data inicial.'
        return
      }
      await receipts.applyFilters(draft)
    }

    async function clearFilters() {
      Object.assign(draft, emptyReceiptFilters())
      filterError.value = ''
      await receipts.clearFilters()
    }

    function printReceipt(receipt: IReceipt) {
      printReceiptDocument(receipt, auth.user)
    }

    function printInvoiceData(receipt: IReceipt) {
      return prepareReceiptInvoiceIssue(receipt, auth.user)
    }

    function shareWhatsApp(receipt: IReceipt) {
      return receipts.shareWhatsApp(receipt, auth.user)
    }

    async function cancelReceipt(receipt: IReceipt) {
      const confirmed = await feedback.confirm({ title: 'Cancelar recibo?', message: `Cancelar o recibo de ${receiptClientName(receipt)}? Ele não reservará mais produtos no estoque.`, tone: 'danger', confirmLabel: 'Cancelar recibo' })
      if (!confirmed) return

      await receipts.cancel(receipt.id)
    }

    function startCreate() {
      return router.push('/receipts/new')
    }

    function startQuickCreate() {
      return router.push('/receipts/new?quick=1')
    }

    function editReceipt(receipt: IReceipt) {
      return router.push(`/receipts/${receipt.id}/edit`)
    }

    function previousPage() {
      return receipts.load(receipts.offset - receipts.limit)
    }

    function nextPage() {
      return receipts.load(receipts.offset + receipts.limit)
    }

    return {
      applyFilters,
      cancelReceipt,
      clearFilters,
      currentPage,
      draft,
      editReceipt,
      filterError,
      nextPage,
      pages,
      printInvoiceData,
      previousPage,
      printReceipt,
      receipts,
      shareWhatsApp,
      startCreate,
      startQuickCreate
    }
  }
})
</script>

<template>
  <section class="page">
    <PageHeader title="Recibos" subtitle="Crie orçamentos, acompanhe pagamentos e baixe produtos do estoque.">
      <template #actions>
        <div class="receipts-template__actions">
          <button class="button button--secondary" type="button" @click="startQuickCreate">
            <Zap :size="18" />
            Recibo rápido
          </button>
          <button class="button button--primary" type="button" @click="startCreate">
            <Plus :size="18" />
            Adicionar
          </button>
        </div>
      </template>
    </PageHeader>

    <form class="receipts-template__filters panel" aria-label="Filtros de recibos" @submit.prevent="applyFilters">
      <fieldset :disabled="receipts.loading">
        <legend>Filtrar recibos</legend>
        <div class="receipts-template__filter-grid">
          <label class="field receipts-template__search">
            <span>Cliente, telefone, placa ou serviço</span>
            <input v-model="draft.q" type="search" placeholder="Buscar recibos" />
          </label>
          <label class="field">
            <span>Data inicial de criação</span>
            <input v-model="draft.startDate" type="date" :aria-invalid="Boolean(filterError)" :aria-describedby="filterError ? 'receipt-filter-error' : undefined" @input="filterError = ''" />
          </label>
          <label class="field">
            <span>Data final de criação</span>
            <input v-model="draft.endDate" type="date" :aria-invalid="Boolean(filterError)" :aria-describedby="filterError ? 'receipt-filter-error' : undefined" @input="filterError = ''" />
          </label>
          <label class="field">
            <span>Situação</span>
            <select v-model="draft.status" aria-label="Situação">
              <option value="">Todas as situações</option>
              <option value="pending">Pendente</option>
              <option value="paid">Pago</option>
              <option value="cancelled">Cancelado</option>
            </select>
          </label>
          <label class="field">
            <span>Forma de pagamento</span>
            <select v-model="draft.paymentMethod" aria-label="Forma de pagamento">
              <option value="">Todas as formas</option>
              <option value="cash">Dinheiro</option>
              <option value="pix">PIX</option>
              <option value="credit_card">Cartão de crédito</option>
              <option value="debit_card">Cartão de débito</option>
              <option value="mixed">Pagamentos mistos</option>
              <option value="legacy">Sem forma registrada</option>
            </select>
          </label>
          <label class="field">
            <span>Tipo de recibo</span>
            <select v-model="draft.type" aria-label="Tipo de recibo">
              <option value="">Todos os tipos</option>
              <option value="normal">Normal</option>
              <option value="quick">Rápido</option>
            </select>
          </label>
        </div>
        <p id="receipt-filter-error" v-if="filterError" class="form-error" role="alert">{{ filterError }}</p>
        <div class="receipts-template__filter-actions">
          <button class="button button--secondary" type="button" @click="clearFilters">Limpar filtros</button>
          <button class="button button--primary" type="submit">Aplicar filtros</button>
        </div>
      </fieldset>
    </form>

    <p v-if="receipts.error" class="form-error" role="alert">{{ receipts.error }}</p>
    <div v-if="receipts.loading" class="panel empty" role="status">Carregando recibos...</div>
    <div v-else-if="!receipts.error && !receipts.receipts.length" class="panel">
      <EmptyState title="Nenhum recibo encontrado" description="Revise os filtros ou limpe a seleção para consultar todos os recibos." />
    </div>

    <ReceiptsTable
      v-else-if="!receipts.error"
      :receipts="receipts.receipts"
      @cancel="cancelReceipt"
      @copy-instagram="receipts.copyInstagramText"
      @edit="editReceipt"
      @mark-paid="(receipt) => receipts.markPaid(receipt.id)"
      @print-invoice-data="printInvoiceData"
      @print="printReceipt"
      @share-whatsapp="shareWhatsApp"
    />

    <div v-if="!receipts.loading && !receipts.error" aria-live="polite">
      <p class="receipts-template__total">{{ receipts.total }} {{ receipts.total === 1 ? 'recibo encontrado' : 'recibos encontrados' }}</p>
      <PaginationControls :current-page="currentPage" :pages="pages" @next="nextPage" @previous="previousPage" />
    </div>
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
