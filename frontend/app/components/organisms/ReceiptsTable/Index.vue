<script lang="ts">
import { Ban, CheckCircle2, Copy, FileText, MessageCircle, Pencil, Printer } from '@lucide/vue'
import { defineComponent, type PropType } from 'vue'
import type { IReceipt } from '../../../../server/contracts/types'
import IconActionButton from '../../atoms/IconActionButton/Index.vue'
import EmptyState from '../../molecules/EmptyState/Index.vue'
import { formatCurrency, formatDateTime } from '../../../utils/format'
import { receiptClientName, receiptVehicleName, receiptVehiclePlate } from '../../../utils/receiptDisplay'

export default defineComponent({
  name: 'ReceiptsTable',
  components: {
    Ban,
    CheckCircle2,
    Copy,
    EmptyState,
    FileText,
    IconActionButton,
    MessageCircle,
    Pencil,
    Printer
  },
  props: {
    receipts: {
      type: Array as PropType<IReceipt[]>,
      required: true
    }
  },
  emits: ['cancel', 'copy-instagram', 'edit', 'mark-paid', 'print', 'print-invoice-data', 'share-whatsapp'],
  setup(_, { emit }) {
    function cancel(receipt: IReceipt) {
      emit('cancel', receipt)
    }

    function copyInstagram(receipt: IReceipt) {
      emit('copy-instagram', receipt)
    }

    function markPaid(receipt: IReceipt) {
      emit('mark-paid', receipt)
    }

    function edit(receipt: IReceipt) {
      emit('edit', receipt)
    }

    function print(receipt: IReceipt) {
      emit('print', receipt)
    }

    function printInvoiceData(receipt: IReceipt) {
      emit('print-invoice-data', receipt)
    }

    function shareWhatsApp(receipt: IReceipt) {
      emit('share-whatsapp', receipt)
    }

    function paymentMethodLabel(receipt: IReceipt) {
      if (receipt.paymentMethod === 'credit_card') return `Crédito ${receipt.installments || 1}x`
      if (receipt.paymentMethod === 'debit_card') return 'Débito'
      if (receipt.paymentMethod === 'pix') return 'Pix'
      return 'Dinheiro'
    }

    function statusBadgeClass(status: IReceipt['status']) {
      if (status === 'paid') return 'badge--paid'
      if (status === 'cancelled') return 'badge--cancelled'
      return 'badge--pending'
    }

    function statusLabel(status: IReceipt['status']) {
      if (status === 'paid') return 'Pago'
      if (status === 'cancelled') return 'Cancelado'
      return 'Pendente'
    }

    return {
      cancel,
      copyInstagram,
      edit,
      formatCurrency,
      formatDateTime,
      markPaid,
      paymentMethodLabel,
      receiptClientName,
      receiptVehicleName,
      receiptVehiclePlate,
      print,
      printInvoiceData,
      shareWhatsApp,
      statusBadgeClass,
      statusLabel
    }
  }
})
</script>

<template>
  <section class="receipts-list panel">
    <div v-if="receipts.length" class="receipts-list__head" aria-hidden="true">
      <span>Cliente</span>
      <span>Veículo</span>
      <span>Total</span>
      <span>Pagamento</span>
      <span>Emitido em</span>
      <span>Status</span>
      <span>Ações</span>
    </div>

    <article v-for="receipt in receipts" :key="receipt.id" class="receipts-list__row">
      <div class="receipts-list__cell receipts-list__cell--client">
        <span class="receipts-list__label">Cliente</span>
        <strong>{{ receiptClientName(receipt) }}</strong>
      </div>
      <div class="receipts-list__cell">
        <span class="receipts-list__label">Veículo</span>
        <strong>{{ receiptVehicleName(receipt) }}</strong>
        <small>{{ receiptVehiclePlate(receipt) }}</small>
      </div>
      <div class="receipts-list__cell receipts-list__cell--money">
        <span class="receipts-list__label">Total</span>
        <strong>{{ formatCurrency(receipt.priceCents) }}</strong>
      </div>
      <div class="receipts-list__cell">
        <span class="receipts-list__label">Pagamento</span>
        <strong>{{ paymentMethodLabel(receipt) }}</strong>
      </div>
      <div class="receipts-list__cell">
        <span class="receipts-list__label">Emitido em</span>
        <strong>{{ formatDateTime(receipt.createdAt) }}</strong>
      </div>
      <div class="receipts-list__cell receipts-list__cell--status">
        <span class="receipts-list__label">Status</span>
        <span class="badge" :class="statusBadgeClass(receipt.status)">
          {{ statusLabel(receipt.status) }}
        </span>
      </div>
      <div class="receipts-list__actions">
        <IconActionButton title="WhatsApp" @click="shareWhatsApp(receipt)">
          <MessageCircle :size="16" />
        </IconActionButton>
        <IconActionButton title="Copiar para Instagram" @click="copyInstagram(receipt)">
          <Copy :size="16" />
        </IconActionButton>
        <IconActionButton title="PDF / imprimir" @click="print(receipt)">
          <Printer :size="16" />
        </IconActionButton>
        <IconActionButton title="Preparar emissão NFS-e" @click="printInvoiceData(receipt)">
          <FileText :size="16" />
        </IconActionButton>
        <IconActionButton v-if="receipt.status === 'pending'" title="Editar recibo" @click="edit(receipt)">
          <Pencil :size="16" />
        </IconActionButton>
        <IconActionButton v-if="receipt.status === 'pending'" title="Cancelar recibo" variant="danger" @click="cancel(receipt)">
          <Ban :size="16" />
        </IconActionButton>
        <IconActionButton v-if="receipt.status === 'pending'" title="Marcar como pago" variant="primary" @click="markPaid(receipt)">
          <CheckCircle2 :size="16" />
        </IconActionButton>
      </div>
    </article>

    <EmptyState
      v-if="!receipts.length"
      title="Nenhum recibo encontrado"
      description="Crie o primeiro recibo para acompanhar pagamentos, serviços e consumo de estoque."
    />
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
