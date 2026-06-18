<script lang="ts">
import { ArrowLeft } from '@lucide/vue'
import { defineComponent } from 'vue'
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
    const receipts = useReceiptsStore()
    const stock = useStockStore()

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
      receipts,
      stock
    }
  }
})
</script>

<template>
  <section class="page">
    <PageHeader title="Adicionar recibo" subtitle="Registre o serviço, produtos utilizados e gastos internos.">
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
      :stock-items="stock.items"
      @clear-field-error="receipts.clearFieldError"
    />
  </section>
</template>
