<script setup lang="ts">
import StockEditTemplate from '../../../components/templates/StockEditTemplate/Index.vue'

const route = useRoute()
</script>

<template>
  <StockEditTemplate :product-id="String(route.params.id || '')" />
</template>
