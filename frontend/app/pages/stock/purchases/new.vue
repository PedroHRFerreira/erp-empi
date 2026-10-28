<script setup lang="ts">
const route = useRoute()
const productId = typeof route.query.productId === 'string' ? route.query.productId : ''
const query: Record<string, string> = {}
for (const key of ['supplierName', 'supplierMode', 'returnTo']) {
  if (typeof route.query[key] === 'string') query[key] = route.query[key]
}
if (productId) query.replenish = '1'
await navigateTo({ path: productId ? `/stock/${encodeURIComponent(productId)}/edit` : '/stock/new', query }, { replace: true })
</script>

<template><span role="status">Abrindo produto...</span></template>
