<script lang="ts">
import { ArrowLeft } from '@lucide/vue'
import { computed, defineComponent } from 'vue'
import type { ReceiptForm } from '../../../stores/useReceiptsStore'
import PageHeader from '../../molecules/PageHeader/Index.vue'
import ReceiptsForm from '../../organisms/ReceiptsForm/Index.vue'

export default defineComponent({
  name: 'ReceiptEditTemplate',
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
    const receiptId = computed(() => String(route.params.id || ''))
    const receipt = computed(() => receipts.receiptDetail)
    const isPending = computed(() => receipt.value?.status === 'pending')

    async function updateReceipt(form: ReceiptForm) {
      const result = await receipts.update(receiptId.value, form)

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
      isPending,
      receipt,
      receipts,
      stock,
      updateReceipt
    }
  }
})
</script>

<template>
  <section class="page">
    <PageHeader title="Editar recibo" subtitle="Atualize os dados do recibo pendente.">
      <template #actions>
        <button class="button button--secondary" type="button" @click="backToList">
          <ArrowLeft :size="18" />
          Voltar
        </button>
      </template>
    </PageHeader>

    <p v-if="receipts.loading" class="panel empty">Carregando recibo...</p>
    <p v-else-if="receipts.error" class="form-error">{{ receipts.error }}</p>
    <p v-else-if="receipt && !isPending" class="panel empty">Apenas recibos pendentes podem ser editados.</p>
    <ReceiptsForm
      v-else-if="receipt"
      :error="receipts.error"
      :field-errors="receipts.fieldErrors"
      mode="edit"
      :on-create="updateReceipt"
      :quick="!receipt.user"
      :stock-items="stock.items"
      @back-to-list="backToList"
      @clear-field-error="receipts.clearFieldError"
    />
  </section>
</template>
