<script lang="ts">
import { ArrowLeft, MessageCircle, Printer } from '@lucide/vue'
import { defineComponent } from 'vue'
import type { IReceipt } from '../../../../server/contracts/types'
import { formatCurrency, formatDateTime } from '../../../utils/format'
import { maskPhone } from '../../../utils/masks'
import { printReceiptDocument } from '../../../utils/print'
import PageHeader from '../../molecules/PageHeader/Index.vue'

export default defineComponent({
  name: 'ClientDetailTemplate',
  components: {
    ArrowLeft,
    MessageCircle,
    PageHeader,
    Printer
  },
  setup() {
    const clients = useClientsStore()
    const receipts = useReceiptsStore()

    function phone(value: string) {
      return value ? maskPhone(value) : '-'
    }

    function statusLabel(status: IReceipt['status']) {
      if (status === 'paid') return 'Pago'
      if (status === 'cancelled') return 'Cancelado'
      return 'Pendente'
    }

    function paymentLabel(receipt: IReceipt) {
      if (receipt.paymentMethod === 'credit_card') return `Cartão de crédito (${receipt.installments || 1}x)`
      if (receipt.paymentMethod === 'debit_card') return 'Cartão de débito'
      if (receipt.paymentMethod === 'pix') return 'Pix'
      return 'Dinheiro'
    }

    function itemTotal(item: IReceipt['items'][number]) {
      return item.unitResaleCents * item.quantity
    }

    function serviceExpensesTotal(receipt: IReceipt) {
      return receipt.expenses?.reduce((total, expense) => total + expense.amountCents, 0) || 0
    }

    return {
      clients,
      formatCurrency,
      formatDateTime,
      itemTotal,
      paymentLabel,
      phone,
      printReceiptDocument,
      receipts,
      serviceExpensesTotal,
      statusLabel
    }
  }
})
</script>

<template>
  <section class="page">
    <PageHeader
      :title="clients.detail?.client.name || 'Cliente'"
      subtitle="Histórico completo de recibos do cliente."
    >
      <template #actions>
        <NuxtLink class="button button--secondary" to="/clients">
          <ArrowLeft :size="18" />
          Clientes
        </NuxtLink>
      </template>
    </PageHeader>

    <div v-if="clients.loading" class="panel empty">Carregando cliente...</div>
    <p v-else-if="clients.error" class="form-error">{{ clients.error }}</p>

    <template v-else-if="clients.detail">
      <section class="client-detail__profile panel">
        <div>
          <span>Telefone</span>
          <strong>{{ phone(clients.detail.client.phone) }}</strong>
        </div>
        <div>
          <span>Recibos</span>
          <strong>{{ clients.detail.receipts.length }}</strong>
        </div>
      </section>

      <section class="client-detail__receipts">
        <details v-for="receipt in clients.detail.receipts" :key="receipt.id" class="client-receipt panel">
          <summary>
            <div>
              <strong>{{ receipt.vehicleModel }} {{ receipt.vehicleYear }}</strong>
              <span>{{ receipt.vehiclePlate }} · {{ formatDateTime(receipt.createdAt) }}</span>
            </div>
            <div class="client-receipt__summary-total">
              <strong>{{ formatCurrency(receipt.priceCents) }}</strong>
              <span>{{ statusLabel(receipt.status) }}</span>
            </div>
          </summary>

          <div class="client-receipt__body">
            <section>
              <h2>Serviços</h2>
              <p>{{ receipt.services }}</p>
              <p v-if="receipt.notes" class="client-receipt__muted">{{ receipt.notes }}</p>
            </section>

            <section>
              <h2>Produtos</h2>
              <ul v-if="receipt.items.length" class="client-receipt__items">
                <li v-for="item in receipt.items" :key="item.id">
                  <span>{{ item.stockItem?.name || item.stockItemId }}</span>
                  <strong>{{ item.quantity }}x {{ formatCurrency(item.unitResaleCents) }}</strong>
                  <em>{{ formatCurrency(itemTotal(item)) }}</em>
                </li>
              </ul>
              <p v-else class="client-receipt__muted">Nenhum produto vinculado.</p>
            </section>

            <section class="client-receipt__money">
              <div>
                <span>Mão de obra</span>
                <strong>{{ formatCurrency(receipt.laborPriceCents) }}</strong>
              </div>
              <div>
                <span>Produtos</span>
                <strong>{{ formatCurrency(receipt.productsTotalCents) }}</strong>
              </div>
              <div v-if="serviceExpensesTotal(receipt)">
                <span>Gastos do serviço</span>
                <strong>{{ formatCurrency(serviceExpensesTotal(receipt)) }}</strong>
              </div>
              <div v-if="receipt.discountCents">
                <span>Desconto</span>
                <strong>{{ formatCurrency(-receipt.discountCents) }}</strong>
              </div>
              <div>
                <span>Subtotal</span>
                <strong>{{ formatCurrency(receipt.subtotalCents) }}</strong>
              </div>
              <div>
                <span>Pagamento</span>
                <strong>{{ paymentLabel(receipt) }}</strong>
              </div>
              <div class="client-receipt__total">
                <span>Total final</span>
                <strong>{{ formatCurrency(receipt.priceCents) }}</strong>
              </div>
            </section>

            <footer class="client-receipt__actions">
              <button class="button button--secondary" type="button" @click="receipts.shareWhatsApp(receipt)">
                <MessageCircle :size="18" />
                WhatsApp
              </button>
              <button class="button button--secondary" type="button" @click="printReceiptDocument(receipt)">
                <Printer :size="18" />
                PDF
              </button>
            </footer>
          </div>
        </details>
      </section>
    </template>
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
