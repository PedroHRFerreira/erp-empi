<script lang="ts">
import { BarChart3, Boxes, CircleDollarSign, FileText, LogOut, RotateCcw, Target, UserRound, UsersRound } from '@lucide/vue'
import { defineComponent } from 'vue'
import BrandMark from '../../atoms/BrandMark/Index.vue'

export default defineComponent({
  name: 'AppSidebar',
  components: {
    BarChart3,
    Boxes,
    BrandMark,
    FileText,
    CircleDollarSign,
    LogOut,
    RotateCcw,
    Target,
    UserRound,
    UsersRound
  },
  setup() {
    const auth = useAuthStore()
    const links = [
      { to: '/', label: 'Métricas', icon: BarChart3 },
      { to: '/goals', label: 'Metas', icon: Target },
      { to: '/receipts', label: 'Recibos', icon: FileText },
      { to: '/recovery', label: 'Recuperação', icon: RotateCcw },
      { to: '/clients', label: 'Clientes', icon: UsersRound },
      { to: '/expenses', label: 'Gastos', icon: CircleDollarSign },
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
        <span class="app-sidebar__label">{{ link.label }}</span>
      </NuxtLink>
    </nav>

    <button class="app-sidebar__logout" type="button" @click="logout">
      <LogOut :size="18" />
      <span class="app-sidebar__label">Sair</span>
    </button>
  </aside>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
