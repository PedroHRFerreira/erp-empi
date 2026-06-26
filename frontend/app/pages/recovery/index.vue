<script lang="ts">
import { defineComponent } from 'vue'
import RecoveryTemplate from '../../components/templates/RecoveryTemplate/Index.vue'

export default defineComponent({
  name: 'RecoveryPage',
  components: {
    RecoveryTemplate
  },
  async setup() {
    const receipts = useReceiptsStore()
    await receipts.load(0, 'cancelled')

    return {}
  }
})
</script>

<template>
  <RecoveryTemplate />
</template>
