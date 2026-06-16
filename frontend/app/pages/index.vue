<script setup lang="ts">
import { Banknote, Clock3, UsersRound, Wrench } from '@lucide/vue'

const metrics = useMetricsStore()
await metrics.load()
</script>

<template>
  <section class="page">
    <header class="page__header">
      <div>
        <h1 class="page__title">Metricas</h1>
        <p class="page__subtitle">Visao rapida da operacao da oficina.</p>
      </div>
    </header>

    <div v-if="metrics.loading" class="panel empty">Carregando metricas...</div>

    <template v-else-if="metrics.summary">
      <div class="metrics-grid">
        <article class="metric panel">
          <UsersRound :size="22" />
          <span>Clientes</span>
          <strong>{{ metrics.summary.clientsTotal }}</strong>
        </article>
        <article class="metric panel">
          <Wrench :size="22" />
          <span>Recibos</span>
          <strong>{{ metrics.summary.receiptsTotal }}</strong>
        </article>
        <article class="metric panel">
          <Banknote :size="22" />
          <span>Receita paga</span>
          <strong>{{ formatCurrency(metrics.summary.revenuePaidCents) }}</strong>
        </article>
        <article class="metric panel">
          <Clock3 :size="22" />
          <span>Pendentes</span>
          <strong>{{ metrics.summary.receiptsPending }}</strong>
        </article>
      </div>

      <div class="dashboard-grid">
        <section class="panel table-wrap">
          <header class="block-header">
            <h2>Produtos mais usados</h2>
            <NuxtLink class="button button--secondary" to="/stock">Ver mais</NuxtLink>
          </header>
          <table>
            <thead>
              <tr>
                <th>Produto</th>
                <th>Usados</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="product in metrics.summary.topProducts" :key="product.id">
                <td>{{ product.name }}</td>
                <td>{{ product.usedQuantity }}</td>
              </tr>
            </tbody>
          </table>
        </section>

        <section class="panel table-wrap">
          <header class="block-header">
            <h2>Recibos pendentes</h2>
            <NuxtLink class="button button--secondary" to="/receipts">Ver mais</NuxtLink>
          </header>
          <table>
            <thead>
              <tr>
                <th>Cliente</th>
                <th>Valor</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="receipt in metrics.summary.pendingReceipts" :key="receipt.id">
                <td>{{ receipt.clientName }}</td>
                <td>{{ formatCurrency(receipt.priceCents) }}</td>
              </tr>
            </tbody>
          </table>
        </section>
      </div>
    </template>
  </section>
</template>

<style scoped lang="scss">
.metrics-grid,
.dashboard-grid {
  display: grid;
  gap: 16px;
}

.metrics-grid {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.dashboard-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.metric {
  display: grid;
  gap: 8px;
  padding: 18px;
}

.metric svg {
  color: var(--braip-brand-primary);
}

.metric span {
  color: var(--braip-theme-text-muted);
  font-weight: 700;
}

.metric strong {
  font-size: 28px;
}

.block-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px;
}

.block-header h2 {
  margin: 0;
  font-size: 18px;
}

@media (max-width: 980px) {
  .metrics-grid,
  .dashboard-grid {
    grid-template-columns: 1fr;
  }
}
</style>
