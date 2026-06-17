<script lang="ts">
import { BarChart3, Boxes, FileText, LogOut, UserRound, UsersRound } from '@lucide/vue'
import { defineComponent } from 'vue'
import BrandMark from '../../atoms/BrandMark/Index.vue'

export default defineComponent({
  name: 'AppSidebar',
  components: {
    BarChart3,
    Boxes,
    BrandMark,
    FileText,
    LogOut,
    UserRound,
    UsersRound
  },
  setup() {
    const auth = useAuthStore()
    const links = [
      { to: '/', label: 'Métricas', icon: BarChart3 },
      { to: '/receipts', label: 'Recibos', icon: FileText },
      { to: '/clients', label: 'Clientes', icon: UsersRound },
      { to: '/stock', label: 'Estoque', icon: Boxes },
      { to: '/profile', label: 'Perfil', icon: UserRound }
    ]

    async function logout() {
      await auth.logout()
      await navigateTo('/login')
    }

    return {
      links,
      logout
    }
  }
})
</script>

<template>
  <aside class="app-sidebar">
    <BrandMark title="EMPI ERP" to="/" />

    <nav class="app-sidebar__nav" aria-label="Navegação principal">
      <NuxtLink v-for="link in links" :key="link.to" :to="link.to">
        <component :is="link.icon" :size="18" />
        {{ link.label }}
      </NuxtLink>
    </nav>

    <button class="app-sidebar__logout" type="button" @click="logout">
      <LogOut :size="18" />
      Sair
    </button>
  </aside>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
