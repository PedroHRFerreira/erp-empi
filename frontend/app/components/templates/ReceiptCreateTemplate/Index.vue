<script lang="ts">
import { ArrowLeft } from '@lucide/vue'
import { computed, defineComponent } from 'vue'
import type { ReceiptForm } from '../../../stores/useReceiptsStore'
import PageHeader from '../../molecules/PageHeader/Index.vue'
import ReceiptsForm from '../../organisms/ReceiptsForm/Index.vue'

export default defineComponent({
  name: 'ReceiptCreateTemplate',
  components: {
    ArrowLeft,
    PageHeader,
    ReceiptsForm
  },
  setup() {
    const router = useRouter()
    const route = useRoute()
    const receipts = useReceiptsStore()
    const stock = useStockStore()
    const isQuickReceipt = computed(() => route.query.quick === '1')
    const pageTitle = computed(() => (isQuickReceipt.value ? 'Recibo rápido' : 'Adicionar recibo'))
    const pageSubtitle = computed(() => {
      if (isQuickReceipt.value) return 'Registre um recibo sem cliente e sem veículo.'
      return 'Registre o serviço, produtos utilizados e gastos internos.'
    })

    async function createReceipt(form: ReceiptForm) {
      const result = await receipts.create(form)

      if (result.status === 'success') {
        await router.push('/receipts')
      }

      return result
    }

    function backToList() {
      return router.push('/receipts')
    }

    return {
      backToList,
      createReceipt,
      isQuickReceipt,
      pageSubtitle,
      pageTitle,
      receipts,
      stock
    }
  }
})
</script>

<template>
  <section class="page">
    <PageHeader :title="pageTitle" :subtitle="pageSubtitle">
      <template #actions>
        <button class="button button--secondary" type="button" @click="backToList">
          <ArrowLeft :size="18" />
          Voltar
        </button>
      </template>
    </PageHeader>

    <ReceiptsForm
      :error="receipts.error"
      :field-errors="receipts.fieldErrors"
      :on-create="createReceipt"
      :quick="isQuickReceipt"
      :stock-items="stock.items"
      @back-to-list="backToList"
      @clear-field-error="receipts.clearFieldError"
    />
  </section>
</template>
