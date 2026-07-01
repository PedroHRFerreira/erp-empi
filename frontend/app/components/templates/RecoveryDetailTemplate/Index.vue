<script lang="ts">
import { ArrowLeft, Copy, MessageCircle, Printer, RotateCcw } from '@lucide/vue'
import { computed, defineComponent } from 'vue'
import type { IReceipt } from '../../../../server/contracts/types'
import { formatCurrency, formatDateTime } from '../../../utils/format'
import { maskPhone } from '../../../utils/masks'
import { printReceiptDocument } from '../../../utils/print'
import { receiptClientName, receiptClientPhone, receiptVehicleName, receiptVehiclePlate } from '../../../utils/receiptDisplay'
import PageHeader from '../../molecules/PageHeader/Index.vue'

export default defineComponent({
  name: 'RecoveryDetailTemplate',
  components: {
    ArrowLeft,
    Copy,
    MessageCircle,
    PageHeader,
    Printer,
    RotateCcw
  },
  setup() {
    const receipts = useReceiptsStore()
    const auth = useAuthStore()
    const receipt = computed(() => receipts.receiptDetail)

    function phone(value: string) {
      return value ? maskPhone(value) : '-'
    }

    function paymentLabel(value: IReceipt) {
      if (value.paymentMethod === 'credit_card') return `Cartão de crédito (${value.installments || 1}x)`
      if (value.paymentMethod === 'debit_card') return 'Cartão de débito'
      if (value.paymentMethod === 'pix') return 'Pix'
      return 'Dinheiro'
    }

    function itemTotal(item: IReceipt['items'][number]) {
      return item.unitResaleCents * item.quantity
    }

    function serviceExpensesTotal(value: IReceipt) {
      return value.expenses?.reduce((total, expense) => total + expense.amountCents, 0) || 0
    }

    async function reopen(value: IReceipt) {
      const confirmed = window.confirm(`Retornar o recibo de ${receiptClientName(value)} para pendente?`)
      if (!confirmed) return

      const result = await receipts.reopen(value.id)
      if (result.status === 'success') {
        await navigateTo('/receipts')
      }
    }

    return {
      auth,
      formatCurrency,
      formatDateTime,
      itemTotal,
      paymentLabel,
      phone,
      printReceiptDocument,
      receipt,
      receiptClientName,
      receiptClientPhone,
      receiptVehicleName,
      receiptVehiclePlate,
      receipts,
      reopen,
      serviceExpensesTotal
    }
  }
})
</script>

<template>
  <section class="page recovery-detail">
    <PageHeader
      :title="receipt ? receiptClientName(receipt) : 'Recuperação'"
      subtitle="Detalhes do recibo cancelado e ações para retomada."
    >
      <template #actions>
        <NuxtLink class="button button--secondary" to="/recovery">
          <ArrowLeft :size="18" />
          Recuperação
        </NuxtLink>
      </template>
    </PageHeader>

    <div v-if="receipts.loading" class="panel empty">Carregando recibo...</div>
    <p v-else-if="receipts.error" class="form-error">{{ receipts.error }}</p>

    <template v-else-if="receipt">
      <section class="recovery-detail__overview panel">
        <div>
          <span>Status</span>
          <strong class="badge badge--cancelled">Cancelado</strong>
        </div>
        <div>
          <span>Cliente</span>
          <strong>{{ receiptClientName(receipt) }}</strong>
        </div>
        <div>
          <span>Telefone</span>
          <strong>{{ phone(receiptClientPhone(receipt)) }}</strong>
        </div>
        <div>
          <span>Total</span>
          <strong>{{ formatCurrency(receipt.priceCents) }}</strong>
        </div>
        <div>
          <span>Emitido em</span>
          <strong>{{ formatDateTime(receipt.createdAt) }}</strong>
        </div>
        <div>
          <span>Cancelado em</span>
          <strong>{{ formatDateTime(receipt.updatedAt) }}</strong>
        </div>
      </section>

      <section class="recovery-detail__actions panel">
        <button class="button button--secondary" type="button" @click="receipts.shareRecoveryWhatsApp(receipt)">
          <MessageCircle :size="18" />
          WhatsApp
        </button>
        <button class="button button--secondary" type="button" @click="receipts.copyRecoveryMessage(receipt)">
          <Copy :size="18" />
          Copiar mensagem
        </button>
        <button class="button button--secondary" type="button" @click="printReceiptDocument(receipt, auth.user)">
          <Printer :size="18" />
          PDF
        </button>
        <button class="button button--primary" type="button" @click="reopen(receipt)">
          <RotateCcw :size="18" />
          Retornar para pendente
        </button>
      </section>

      <section class="recovery-detail__section panel">
        <h2>Recibo cancelado</h2>
        <div class="recovery-detail__vehicle">
          <div>
            <span>Veículo</span>
            <strong>{{ receiptVehicleName(receipt) }}</strong>
          </div>
          <div>
            <span>Placa</span>
            <strong>{{ receiptVehiclePlate(receipt) }}</strong>
          </div>
          <div>
            <span>Pagamento</span>
            <strong>{{ paymentLabel(receipt) }}</strong>
          </div>
        </div>
        <p>{{ receipt.services }}</p>
        <p v-if="receipt.notes" class="recovery-detail__muted">{{ receipt.notes }}</p>
      </section>

      <section class="recovery-detail__section panel">
        <h2>Produtos</h2>
        <ul v-if="receipt.items.length" class="recovery-detail__items">
          <li v-for="item in receipt.items" :key="item.id">
            <span>{{ item.stockItem?.name || item.stockItemId }}</span>
            <strong>{{ item.quantity }}x {{ formatCurrency(item.unitResaleCents) }}</strong>
            <em>{{ formatCurrency(itemTotal(item)) }}</em>
          </li>
        </ul>
        <p v-else class="recovery-detail__muted">Nenhum produto vinculado.</p>
      </section>

      <section class="recovery-detail__money panel">
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
          <span>Total</span>
          <strong>{{ formatCurrency(receipt.priceCents) }}</strong>
        </div>
      </section>
    </template>
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
