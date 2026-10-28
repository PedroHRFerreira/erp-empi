<script setup lang="ts">
import StockTemplate from '../../components/templates/StockTemplate/Index.vue'

const stock = useStockStore()
await stock.load(0)
</script>

<template>
  <StockTemplate view="products" />
</template>
