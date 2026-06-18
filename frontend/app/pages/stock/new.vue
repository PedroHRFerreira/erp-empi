<script lang="ts">
import { defineComponent } from 'vue'
import StockCreateTemplate from '../../components/templates/StockCreateTemplate/Index.vue'

export default defineComponent({
  name: 'StockNewPage',
  components: {
    StockCreateTemplate
  }
})
</script>

<template>
  <StockCreateTemplate />
</template>
