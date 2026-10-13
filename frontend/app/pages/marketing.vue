<script setup lang="ts">
import MarketingTemplate from '../components/templates/MarketingTemplate/Index.vue'

definePageMeta({ middleware: async () => {
  const auth = useAuthStore()
  await auth.bootstrap()
  if (auth.user?.type !== 'admin') return navigateTo('/')
} })
</script>

<template><MarketingTemplate /></template>
