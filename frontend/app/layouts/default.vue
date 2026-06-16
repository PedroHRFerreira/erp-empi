<script setup lang="ts">
import { BarChart3, Boxes, FileText, LogOut, UserRound } from '@lucide/vue'

const auth = useAuthStore()

const links = [
  { to: '/', label: 'Metricas', icon: BarChart3 },
  { to: '/receipts', label: 'Recibos', icon: FileText },
  { to: '/stock', label: 'Estoque', icon: Boxes },
  { to: '/profile', label: 'Perfil', icon: UserRound }
]

async function logout() {
  await auth.logout()
  await navigateTo('/login')
}
</script>

<template>
  <div class="shell">
    <aside class="shell__sidebar">
      <NuxtLink class="shell__brand" to="/">
        <span>EA</span>
        <strong>EMPI ERP</strong>
      </NuxtLink>

      <nav class="shell__nav" aria-label="Navegacao principal">
        <NuxtLink v-for="link in links" :key="link.to" :to="link.to">
          <component :is="link.icon" :size="18" />
          {{ link.label }}
        </NuxtLink>
      </nav>

      <button class="shell__logout" type="button" @click="logout">
        <LogOut :size="18" />
        Sair
      </button>
    </aside>

    <main class="shell__content">
      <slot />
    </main>
  </div>
</template>

<style scoped lang="scss">
.shell {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
}

.shell__sidebar {
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding: 24px;
  border-right: 1px solid var(--braip-theme-border);
  background: var(--braip-neutral-0);
}

.shell__brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-weight: 900;
}

.shell__brand span {
  display: grid;
  width: 42px;
  height: 42px;
  place-items: center;
  border-radius: 8px;
  color: var(--braip-neutral-0);
  background: var(--braip-brand-primary);
}

.shell__nav {
  display: grid;
  gap: 8px;
}

.shell__nav a,
.shell__logout {
  display: inline-flex;
  min-height: 42px;
  align-items: center;
  gap: 10px;
  border-radius: 8px;
  padding: 0 12px;
  color: var(--braip-theme-text-muted);
  font-weight: 700;
}

.shell__nav a.router-link-active {
  color: var(--braip-brand-primary);
  background: rgba(215, 25, 32, 0.08);
}

.shell__logout {
  margin-top: auto;
  border: 0;
  background: transparent;
}

.shell__content {
  min-width: 0;
  padding: 32px;
}

@media (max-width: 900px) {
  .shell {
    grid-template-columns: 1fr;
  }

  .shell__sidebar {
    position: sticky;
    top: 0;
    z-index: 10;
    display: grid;
    grid-template-columns: 1fr;
    padding: 16px;
    border-right: 0;
    border-bottom: 1px solid var(--braip-theme-border);
  }

  .shell__nav {
    grid-auto-flow: column;
    overflow-x: auto;
  }

  .shell__content {
    padding: 20px;
  }
}
</style>
