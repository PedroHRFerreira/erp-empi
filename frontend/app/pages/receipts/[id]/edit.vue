<script lang="ts">
import { defineComponent } from 'vue'
import ReceiptEditTemplate from '../../../components/templates/ReceiptEditTemplate/Index.vue'
import { receiptToForm } from '../../../stores/useReceiptsStore'

export default defineComponent({
  name: 'ReceiptEditPage',
  components: {
    ReceiptEditTemplate
  },
  async setup() {
    const route = useRoute()
    const receipts = useReceiptsStore()
    const stock = useStockStore()

    await Promise.all([stock.load(), receipts.loadDetail(String(route.params.id || ''))])

    if (receipts.receiptDetail) {
      receipts.setReceiptDraft(receiptToForm(receipts.receiptDetail))
    }

    return {}
  }
})
</script>

<template>
  <ReceiptEditTemplate />
</template>
