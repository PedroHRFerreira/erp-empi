<script lang="ts">
import { Banknote, Boxes, CalendarClock, UsersRound } from '@lucide/vue'
import { defineComponent, type PropType } from 'vue'
import type { IMetricsSummary } from '../../../../server/contracts/types'
import { formatCurrency, formatDateTime } from '../../../utils/format'

export default defineComponent({
  name: 'MetricsSummaryGrid',
  components: {
    Banknote,
    Boxes,
    CalendarClock,
    UsersRound
  },
  props: {
    summary: {
      type: Object as PropType<IMetricsSummary>,
      required: true
    }
  },
  setup() {
    return {
      formatCurrency,
      formatDateTime,
      receiptStatusLabel
    }
  }
})

function receiptStatusLabel(status: string | undefined) {
  if (status === 'paid') return 'Pago'
  if (status === 'cancelled') return 'Cancelado'
  if (status === 'pending') return 'Pendente'
  return '-'
}
</script>

<template>
  <div class="metrics-summary-grid">
    <section class="metrics-summary-group panel">
      <header class="metrics-summary-group__header">
        <UsersRound :size="20" />
        <div>
          <span>Operação</span>
          <strong>Volume da oficina</strong>
        </div>
      </header>

      <dl class="metrics-summary-group__items">
        <div class="metrics-summary-item">
          <dt>Clientes</dt>
          <dd>{{ summary.clientsTotal }}</dd>
        </div>
        <div class="metrics-summary-item">
          <dt>Recibos</dt>
          <dd>{{ summary.receiptsTotal }}</dd>
        </div>
        <div class="metrics-summary-item">
          <dt>Pagos</dt>
          <dd>{{ summary.receiptsPaid }}</dd>
        </div>
        <div class="metrics-summary-item">
          <dt>Pendentes</dt>
          <dd>{{ summary.receiptsPending }}</dd>
        </div>
      </dl>
    </section>

    <section class="metrics-summary-group panel">
      <header class="metrics-summary-group__header">
        <Banknote :size="20" />
        <div>
          <span>Financeiro</span>
          <strong>Receita e ticket</strong>
        </div>
      </header>

      <dl class="metrics-summary-group__items">
        <div class="metrics-summary-item metrics-summary-item--stack">
          <dt>Receita paga</dt>
          <dd>{{ formatCurrency(summary.revenuePaidCents) }}</dd>
        </div>
        <div class="metrics-summary-item metrics-summary-item--stack">
          <dt>Receita pendente</dt>
          <dd>{{ formatCurrency(summary.revenuePendingCents) }}</dd>
        </div>
        <div class="metrics-summary-item metrics-summary-item--stack">
          <dt>Ticket médio</dt>
          <dd>{{ formatCurrency(summary.averageTicketPaidCents) }}</dd>
        </div>
      </dl>
    </section>

    <section class="metrics-summary-group panel">
      <header class="metrics-summary-group__header">
        <Boxes :size="20" />
        <div>
          <span>Estoque</span>
          <strong>Disponibilidade</strong>
        </div>
      </header>

      <dl class="metrics-summary-group__items">
        <div class="metrics-summary-item">
          <dt>Itens cadastrados</dt>
          <dd>{{ summary.stockItemsTotal }}</dd>
        </div>
        <div class="metrics-summary-item">
          <dt>Unidades disponíveis</dt>
          <dd>{{ summary.stockUnitsAvailableTotal }}</dd>
        </div>
        <div class="metrics-summary-item">
          <dt>Unidades usadas</dt>
          <dd>{{ summary.stockUnitsUsedTotal }}</dd>
        </div>
        <div class="metrics-summary-item metrics-summary-item--stack">
          <dt>Último item</dt>
          <dd>{{ summary.lastStockItem?.name || '-' }}</dd>
          <small>{{ formatDateTime(summary.lastStockItem?.createdAt || '') }}</small>
        </div>
      </dl>
    </section>

    <section class="metrics-summary-group panel">
      <header class="metrics-summary-group__header">
        <CalendarClock :size="20" />
        <div>
          <span>Último recibo</span>
          <strong>Movimentação recente</strong>
        </div>
      </header>

      <dl class="metrics-summary-group__items">
        <div class="metrics-summary-item metrics-summary-item--stack">
          <dt>Emitido em</dt>
          <dd>{{ formatDateTime(summary.lastReceipt?.createdAt || '') }}</dd>
        </div>
        <div class="metrics-summary-item">
          <dt>Status</dt>
          <dd>{{ receiptStatusLabel(summary.lastReceipt?.status) }}</dd>
        </div>
        <div class="metrics-summary-item metrics-summary-item--stack">
          <dt>Cliente</dt>
          <dd>{{ summary.lastReceipt?.clientName || '-' }}</dd>
        </div>
        <div class="metrics-summary-item metrics-summary-item--stack">
          <dt>Valor</dt>
          <dd>{{ summary.lastReceipt ? formatCurrency(summary.lastReceipt.priceCents) : '-' }}</dd>
        </div>
      </dl>
    </section>
  </div>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
