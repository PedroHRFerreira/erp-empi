<script lang="ts">
import {
  BarChart3,
  Bell,
  Boxes,
  CircleDollarSign,
  CreditCard,
  FileText,
  LogOut,
  Menu,
  RotateCcw,
  Target,
  UserRound,
  UsersRound,
  Wallet,
  X,
} from "@lucide/vue";
import { defineComponent, onMounted, ref, watch } from "vue";
import BrandMark from "../../atoms/BrandMark/Index.vue";
import ThemeToggle from "../../atoms/ThemeToggle/Index.vue";
import FinancialNotificationsDrawer from "../FinancialNotificationsDrawer/Index.vue";

export default defineComponent({
  name: "AppSidebar",
  components: {
    BarChart3,
    Bell,
    Boxes,
    BrandMark,
    FileText,
    FinancialNotificationsDrawer,
    CircleDollarSign,
    CreditCard,
    Wallet,
    LogOut,
    Menu,
    RotateCcw,
    Target,
    ThemeToggle,
    UserRound,
    UsersRound,
    X,
  },
  setup() {
    const auth = useAuthStore();
    const route = useRoute();
    const mobileOpen = ref(false);
    const notificationsOpen = ref(false);
    const purchases = usePurchasesStore();
    const links = [
      { to: "/", label: "Métricas", icon: BarChart3 },
      { to: "/goals", label: "Metas", icon: Target },
      { to: "/receipts", label: "Recibos", icon: FileText },
      { to: "/recovery", label: "Recuperação", icon: RotateCcw },
      { to: "/clients", label: "Clientes", icon: UsersRound },
      { to: "/expenses", label: "Gastos", icon: CircleDollarSign },
      { to: "/cash", label: "Caixa", icon: Wallet },
      { to: "/payables", label: "Contas a pagar", icon: CreditCard },
      { to: "/stock", label: "Estoque", icon: Boxes },
      { to: "/profile", label: "Perfil", icon: UserRound },
    ];

    async function logout() {
      await auth.logout();
      await navigateTo("/login");
    }

    async function openNotifications() {
      notificationsOpen.value = true;
      await purchases.loadAlerts(true);
    }

    watch(() => route.fullPath, () => {
      mobileOpen.value = false;
      notificationsOpen.value = false;
    });

    onMounted(() => purchases.loadAlerts());

    return {
      links,
      logout,
      mobileOpen,
      notificationsOpen,
      openNotifications,
      purchases,
    };
  },
});
</script>

<template>
  <aside :class="{ 'app-sidebar--open': mobileOpen }" class="app-sidebar">
    <div class="app-sidebar__header">
      <BrandMark title="EMPI ERP" to="/" />
      <div class="app-sidebar__header-actions">
        <button class="app-sidebar__notification app-sidebar__notification--mobile" type="button" aria-label="Abrir notificações financeiras" aria-controls="financial-notifications" :aria-expanded="notificationsOpen" @click="openNotifications">
          <Bell :size="18" /><span v-if="purchases.alerts.length" class="app-sidebar__notification-count">{{ purchases.alerts.length }}</span>
        </button>
        <ThemeToggle class="app-sidebar__theme-toggle-mobile" />
        <button
          :aria-expanded="mobileOpen"
          aria-controls="primary-navigation"
          :aria-label="mobileOpen ? 'Fechar menu principal' : 'Abrir menu principal'"
          class="app-sidebar__menu"
          type="button"
          @click="mobileOpen = !mobileOpen"
        >
          <X v-if="mobileOpen" :size="22" />
          <Menu v-else :size="22" />
        </button>
      </div>
    </div>

    <nav id="primary-navigation" class="app-sidebar__nav" aria-label="Navegação principal">
      <NuxtLink v-for="link in links" :key="link.to" :to="link.to">
        <component :is="link.icon" :size="18" />
        <span class="app-sidebar__label">{{ link.label }}</span>
      </NuxtLink>
    </nav>

    <div class="app-sidebar__footer">
      <button class="app-sidebar__notification app-sidebar__notification--desktop" type="button" aria-label="Abrir notificações financeiras" aria-controls="financial-notifications" :aria-expanded="notificationsOpen" @click="openNotifications">
        <Bell :size="18" />
        <span class="app-sidebar__label">Notificações</span>
        <span v-if="purchases.alerts.length" class="app-sidebar__notification-count">{{ purchases.alerts.length }}</span>
      </button>
      <div class="app-sidebar__theme-row">
        <ThemeToggle />
        <span class="app-sidebar__label">Alternar tema</span>
      </div>
      <button class="app-sidebar__logout" type="button" @click="logout">
        <LogOut :size="18" />
        <span class="app-sidebar__label">Sair</span>
      </button>
    </div>

    <FinancialNotificationsDrawer :open="notificationsOpen" :alerts="purchases.alerts" :loading="purchases.alertsLoading" @close="notificationsOpen = false" />
  </aside>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
