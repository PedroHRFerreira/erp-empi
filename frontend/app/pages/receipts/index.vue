<script lang="ts">
import { defineComponent } from 'vue'
import ReceiptsTemplate from '../../components/templates/ReceiptsTemplate/Index.vue'

export default defineComponent({
  name: 'ReceiptsPage',
  components: {
    ReceiptsTemplate
  },
  async setup() {
    const receipts = useReceiptsStore()

    await receipts.load()

    return {}
  }
})
</script>

<template>
  <ReceiptsTemplate />
</template>
