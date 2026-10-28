<script lang="ts">
import { useRecoveryReceiptsStore } from '../../stores/useReceiptsStore'
import { defineComponent } from 'vue'
import RecoveryTemplate from '../../components/templates/RecoveryTemplate/Index.vue'

export default defineComponent({
  name: 'RecoveryPage',
  components: {
    RecoveryTemplate
  },
  async setup() {
    const receipts = useRecoveryReceiptsStore()
    await receipts.load(0, 'cancelled')

    return {}
  }
})
</script>

<template>
  <RecoveryTemplate />
</template>
