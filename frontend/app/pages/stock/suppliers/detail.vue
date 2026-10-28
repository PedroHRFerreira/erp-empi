<script setup lang="ts">
import StockSupplierDetailTemplate from '../../../components/templates/StockSupplierDetailTemplate/Index.vue'

const route = useRoute()
const supplierName = computed(() => typeof route.query.supplierName === 'string' ? route.query.supplierName : '')
</script>

<template><StockSupplierDetailTemplate :key="supplierName" :supplier-name="supplierName" /></template>
