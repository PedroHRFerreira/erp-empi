<script lang="ts">
import { Plus } from '@lucide/vue'
import { computed, defineComponent } from 'vue'
import type { IReceipt } from '../../../../server/contracts/types'
import PageHeader from '../../molecules/PageHeader/Index.vue'
import PaginationControls from '../../molecules/PaginationControls/Index.vue'
import ReceiptsTable from '../../organisms/ReceiptsTable/Index.vue'
import { prepareReceiptInvoiceIssue, printReceiptDocument } from '../../../utils/print'

export default defineComponent({
  name: 'ReceiptsTemplate',
  components: {
    PageHeader,
    PaginationControls,
    Plus,
    ReceiptsTable
  },
  setup() {
    const router = useRouter()
    const receipts = useReceiptsStore()
    const auth = useAuthStore()
    const pages = computed(() => Math.ceil(receipts.total / receipts.limit))
    const currentPage = computed(() => Math.floor(receipts.offset / receipts.limit) + 1)

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
      const confirmed = window.confirm(`Cancelar o recibo de ${receipt.user.name}? Ele não reservará mais produtos no estoque.`)
      if (!confirmed) return

      await receipts.cancel(receipt.id)
    }

    function startCreate() {
      return router.push('/receipts/new')
    }

    function previousPage() {
      return receipts.load(receipts.offset - receipts.limit)
    }

    function nextPage() {
      return receipts.load(receipts.offset + receipts.limit)
    }

    return {
      cancelReceipt,
      currentPage,
      nextPage,
      pages,
      printInvoiceData,
      previousPage,
      printReceipt,
      receipts,
      shareWhatsApp,
      startCreate
    }
  }
})
</script>

<template>
  <section class="page">
    <PageHeader title="Recibos" subtitle="Crie orçamentos, acompanhe pagamentos e baixe produtos do estoque.">
      <template #actions>
        <button class="button button--primary" type="button" @click="startCreate">
          <Plus :size="18" />
          Adicionar
        </button>
      </template>
    </PageHeader>

    <ReceiptsTable
      :receipts="receipts.receipts"
      @cancel="cancelReceipt"
      @copy-instagram="receipts.copyInstagramText"
      @mark-paid="(receipt) => receipts.markPaid(receipt.id)"
      @print-invoice-data="printInvoiceData"
      @print="printReceipt"
      @share-whatsapp="shareWhatsApp"
    />

    <PaginationControls :current-page="currentPage" :pages="pages" @next="nextPage" @previous="previousPage" />
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
