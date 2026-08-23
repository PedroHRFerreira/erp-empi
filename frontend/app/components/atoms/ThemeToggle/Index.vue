<script setup lang="ts">
import { Moon, Sun } from '@lucide/vue'

const { theme, toggle } = useTheme()
</script>

<template>
  <button
    class="theme-toggle"
    type="button"
    :aria-label="theme === 'dark' ? 'Ativar tema claro' : 'Ativar tema escuro'"
    :title="theme === 'dark' ? 'Tema claro' : 'Tema escuro'"
    @click="toggle"
  >
    <Sun v-if="theme === 'dark'" :size="18" aria-hidden="true" />
    <Moon v-else :size="18" aria-hidden="true" />
  </button>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
