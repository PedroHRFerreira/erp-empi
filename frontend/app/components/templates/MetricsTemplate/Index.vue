<script lang="ts">
import { defineComponent } from 'vue'
import MetricsSummaryGrid from '../../organisms/MetricsSummaryGrid/Index.vue'
import MetricsTableSection from '../../organisms/MetricsTableSection/Index.vue'
import PageHeader from '../../molecules/PageHeader/Index.vue'
import { formatCurrency, formatDateTime } from '../../../utils/format'

export default defineComponent({
  name: 'MetricsTemplate',
  components: {
    MetricsSummaryGrid,
    MetricsTableSection,
    PageHeader
  },
  setup() {
    const metrics = useMetricsStore()

    return {
      formatCurrency,
      formatDateTime,
      metrics
    }
  }
})
</script>

<template>
  <section class="page">
    <PageHeader title="Métricas" subtitle="Visão rápida da operação da oficina." />

    <div v-if="metrics.loading" class="panel empty">Carregando métricas...</div>

    <template v-else-if="metrics.summary">
      <MetricsSummaryGrid :summary="metrics.summary" />

      <div class="metrics-template__domains">
        <section class="metrics-template__domain">
          <header class="metrics-template__domain-header">
            <span>Recibos</span>
            <h2>Pagamentos e pendências</h2>
          </header>

          <div class="metrics-template__domain-grid">
            <MetricsTableSection
              title="Recibos pendentes"
              to="/receipts"
              :empty="!metrics.summary.pendingReceipts.length"
              empty-title="Sem recibos pendentes"
              empty-description="Quando houver orçamentos aguardando pagamento, eles aparecem aqui."
            >
              <table>
                <thead>
                  <tr>
                    <th>Cliente</th>
                    <th>Valor</th>
                    <th>Emitido em</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="receipt in metrics.summary.pendingReceipts" :key="receipt.id">
                    <td>{{ receipt.clientName }}</td>
                    <td>{{ formatCurrency(receipt.priceCents) }}</td>
                    <td>{{ formatDateTime(receipt.createdAt) }}</td>
                  </tr>
                </tbody>
              </table>
            </MetricsTableSection>

            <MetricsTableSection
              title="Recibos pagos recentes"
              to="/receipts"
              :empty="!metrics.summary.paidReceipts.length"
              empty-title="Sem recibos pagos"
              empty-description="Os últimos recibos pagos aparecem aqui."
            >
              <table>
                <thead>
                  <tr>
                    <th>Cliente</th>
                    <th>Valor</th>
                    <th>Emitido em</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="receipt in metrics.summary.paidReceipts" :key="receipt.id">
                    <td>{{ receipt.clientName }}</td>
                    <td>{{ formatCurrency(receipt.priceCents) }}</td>
                    <td>{{ formatDateTime(receipt.createdAt) }}</td>
                  </tr>
                </tbody>
              </table>
            </MetricsTableSection>
          </div>
        </section>

        <section class="metrics-template__domain">
          <header class="metrics-template__domain-header">
            <span>Estoque</span>
            <h2>Uso de produtos e alertas</h2>
          </header>

          <div class="metrics-template__domain-grid">
            <MetricsTableSection
              title="Produtos mais usados"
              to="/stock"
              :empty="!metrics.summary.topProducts.length"
              empty-title="Sem produtos usados"
              empty-description="Os produtos usados em recibos pagos aparecem aqui."
            >
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
            </MetricsTableSection>

            <MetricsTableSection
              title="Estoque baixo"
              to="/stock"
              :empty="!metrics.summary.lowStockProducts.length"
              empty-title="Sem alertas de estoque"
              empty-description="Produtos com até 3 unidades disponíveis aparecem aqui."
            >
              <table>
                <thead>
                  <tr>
                    <th>Produto</th>
                    <th>Disponível</th>
                    <th>Usados</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="product in metrics.summary.lowStockProducts" :key="product.id">
                    <td>{{ product.name }}</td>
                    <td>{{ product.quantity }}</td>
                    <td>{{ product.usedQuantity }}</td>
                  </tr>
                </tbody>
              </table>
            </MetricsTableSection>
          </div>
        </section>

        <section class="metrics-template__domain">
          <header class="metrics-template__domain-header">
            <span>Clientes</span>
            <h2>Movimentação recente</h2>
          </header>

          <div class="metrics-template__domain-grid metrics-template__domain-grid--single">
            <MetricsTableSection
              title="Clientes recentes"
              to="/receipts"
              :empty="!metrics.summary.recentClients.length"
              empty-title="Sem clientes"
              empty-description="Clientes com movimentação recente aparecem aqui."
            >
              <table>
                <thead>
                  <tr>
                    <th>Cliente</th>
                    <th>Recibos</th>
                    <th>Último recibo</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="client in metrics.summary.recentClients" :key="client.id">
                    <td>{{ client.name }}</td>
                    <td>{{ client.receiptsCount }}</td>
                    <td>{{ formatDateTime(client.lastReceiptAt) }}</td>
                  </tr>
                </tbody>
              </table>
            </MetricsTableSection>
          </div>
        </section>
      </div>
    </template>
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
