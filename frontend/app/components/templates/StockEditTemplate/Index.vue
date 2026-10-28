<script setup lang="ts">
import StockCreateTemplate from '../StockCreateTemplate/Index.vue'
defineProps<{ productId: string | number }>()
</script>

<template>
  <StockCreateTemplate :key="productId" :product-id="productId" />
</template>
