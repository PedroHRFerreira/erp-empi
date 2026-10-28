<script lang="ts">
import { defineComponent } from 'vue'
import ReceiptCreateTemplate from '../../components/templates/ReceiptCreateTemplate/Index.vue'

export default defineComponent({
  name: 'ReceiptCreatePage',
  components: {
    ReceiptCreateTemplate
  },
  async setup() {
    const stock = useStockStore()
    await stock.loadOptions()

    return {}
  }
})
</script>

<template>
  <ReceiptCreateTemplate />
</template>
