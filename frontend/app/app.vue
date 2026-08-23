<script setup lang="ts">
import { onMounted } from 'vue'
import ThemeToggle from './components/atoms/ThemeToggle/Index.vue'

const { initialise } = useTheme()
const route = useRoute()

useHead({
  htmlAttrs: { lang: 'pt-BR' },
  title: 'EMPI ERP',
  titleTemplate: (title) => (title && title !== 'EMPI ERP' ? `${title} · EMPI ERP` : 'EMPI ERP')
})

onMounted(initialise)
</script>

<template>
  <ThemeToggle v-if="route.path === '/login'" class="theme-toggle--auth" />
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>
