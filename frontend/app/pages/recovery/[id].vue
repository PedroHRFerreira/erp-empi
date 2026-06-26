<script lang="ts">
import { defineComponent } from 'vue'
import RecoveryDetailTemplate from '../../components/templates/RecoveryDetailTemplate/Index.vue'

export default defineComponent({
  name: 'RecoveryDetailPage',
  components: {
    RecoveryDetailTemplate
  },
  async setup() {
    const route = useRoute()
    const receipts = useReceiptsStore()
    await receipts.loadDetail(String(route.params.id || ''))

    return {}
  }
})
</script>

<template>
  <RecoveryDetailTemplate />
</template>
