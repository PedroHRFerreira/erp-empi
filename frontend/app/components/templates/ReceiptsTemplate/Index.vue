<script lang="ts">
import { Plus } from '@lucide/vue'
import { computed, defineComponent, ref } from 'vue'
import type { IReceipt } from '../../../../server/contracts/types'
import type { ReceiptForm } from '../../../stores/useReceiptsStore'
import PageHeader from '../../molecules/PageHeader/Index.vue'
import PaginationControls from '../../molecules/PaginationControls/Index.vue'
import ReceiptsForm from '../../organisms/ReceiptsForm/Index.vue'
import ReceiptsTable from '../../organisms/ReceiptsTable/Index.vue'
import { printReceiptDocument } from '../../../utils/print'

export default defineComponent({
  name: 'ReceiptsTemplate',
  components: {
    PageHeader,
    PaginationControls,
    Plus,
    ReceiptsForm,
    ReceiptsTable
  },
  setup() {
    const receipts = useReceiptsStore()
    const stock = useStockStore()
    const showForm = ref(false)
    const pages = computed(() => Math.ceil(receipts.total / receipts.limit))
    const currentPage = computed(() => Math.floor(receipts.offset / receipts.limit) + 1)

    async function createReceipt(form: ReceiptForm) {
      const result = await receipts.create(form)

      if (result.status === 'success') {
        showForm.value = false
      }

      return result
    }

    function printReceipt(receipt: IReceipt) {
      printReceiptDocument(receipt)
    }

    function previousPage() {
      return receipts.load(receipts.offset - receipts.limit)
    }

    function nextPage() {
      return receipts.load(receipts.offset + receipts.limit)
    }

    return {
      createReceipt,
      currentPage,
      nextPage,
      pages,
      previousPage,
      printReceipt,
      receipts,
      showForm,
      stock
    }
  }
})
</script>

<template>
  <section class="page">
    <PageHeader title="Recibos" subtitle="Crie orçamentos, acompanhe pagamentos e baixe produtos do estoque.">
      <template #actions>
        <button class="button button--primary" type="button" @click="showForm = !showForm">
          <Plus :size="18" />
          Adicionar
        </button>
      </template>
    </PageHeader>

    <ReceiptsForm
      v-if="showForm"
      :error="receipts.error"
      :field-errors="receipts.fieldErrors"
      :on-create="createReceipt"
      :stock-items="stock.items"
      @clear-field-error="receipts.clearFieldError"
    />

    <ReceiptsTable
      :receipts="receipts.receipts"
      @copy-instagram="receipts.copyInstagramText"
      @mark-paid="(receipt) => receipts.markPaid(receipt.id)"
      @print="printReceipt"
      @share-whatsapp="receipts.shareWhatsApp"
    />

    <PaginationControls :current-page="currentPage" :pages="pages" @next="nextPage" @previous="previousPage" />
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
