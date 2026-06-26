<script lang="ts">
import { Copy, Eye, MessageCircle, RotateCcw } from '@lucide/vue'
import { computed, defineComponent } from 'vue'
import type { IReceipt } from '../../../../server/contracts/types'
import { formatCurrency, formatDateTime } from '../../../utils/format'
import { maskPhone } from '../../../utils/masks'
import IconActionButton from '../../atoms/IconActionButton/Index.vue'
import EmptyState from '../../molecules/EmptyState/Index.vue'
import PageHeader from '../../molecules/PageHeader/Index.vue'
import PaginationControls from '../../molecules/PaginationControls/Index.vue'

export default defineComponent({
  name: 'RecoveryTemplate',
  components: {
    Copy,
    EmptyState,
    Eye,
    IconActionButton,
    MessageCircle,
    PageHeader,
    PaginationControls,
    RotateCcw
  },
  setup() {
    const receipts = useReceiptsStore()
    const pages = computed(() => Math.ceil(receipts.total / receipts.limit))
    const currentPage = computed(() => Math.floor(receipts.offset / receipts.limit) + 1)

    function clientPhone(receipt: IReceipt) {
      return receipt.user?.phone ? maskPhone(receipt.user.phone) : '-'
    }

    function openDetail(receipt: IReceipt) {
      return navigateTo(`/recovery/${receipt.id}`)
    }

    async function reopen(receipt: IReceipt) {
      const confirmed = window.confirm(`Retornar o recibo de ${receipt.user.name} para pendente?`)
      if (!confirmed) return

      await receipts.reopen(receipt.id)
    }

    function previousPage() {
      return receipts.load(receipts.offset - receipts.limit, 'cancelled')
    }

    function nextPage() {
      return receipts.load(receipts.offset + receipts.limit, 'cancelled')
    }

    return {
      clientPhone,
      currentPage,
      formatCurrency,
      formatDateTime,
      nextPage,
      openDetail,
      pages,
      previousPage,
      receipts,
      reopen
    }
  }
})
</script>

<template>
  <section class="page recovery-template">
    <PageHeader title="Recuperação" subtitle="Clientes com recibos cancelados e oportunidades de retomada." />

    <div v-if="receipts.loading" class="panel empty">Carregando recuperação...</div>
    <p v-else-if="receipts.error" class="form-error">{{ receipts.error }}</p>

    <section v-else class="recovery-template__list panel table-wrap">
      <table v-if="receipts.receipts.length">
        <thead>
          <tr>
            <th>Cliente</th>
            <th>Telefone</th>
            <th>Veículo</th>
            <th>Valor</th>
            <th>Cancelado em</th>
            <th class="recovery-template__actions-heading">Ações</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="receipt in receipts.receipts" :key="receipt.id">
            <td>
              <strong>{{ receipt.user.name }}</strong>
            </td>
            <td>{{ clientPhone(receipt) }}</td>
            <td>
              <strong>{{ receipt.vehicleModel }}</strong>
              <small>{{ receipt.vehiclePlate }}</small>
            </td>
            <td>{{ formatCurrency(receipt.priceCents) }}</td>
            <td>{{ formatDateTime(receipt.updatedAt) }}</td>
            <td>
              <div class="recovery-template__actions">
                <IconActionButton title="WhatsApp" @click="receipts.shareRecoveryWhatsApp(receipt)">
                  <MessageCircle :size="16" />
                </IconActionButton>
                <IconActionButton title="Copiar mensagem" @click="receipts.copyRecoveryMessage(receipt)">
                  <Copy :size="16" />
                </IconActionButton>
                <IconActionButton title="Detalhes" @click="openDetail(receipt)">
                  <Eye :size="16" />
                </IconActionButton>
                <IconActionButton title="Retornar para pendente" variant="primary" @click="reopen(receipt)">
                  <RotateCcw :size="16" />
                </IconActionButton>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <EmptyState
        v-if="!receipts.receipts.length"
        title="Nenhum recibo cancelado"
        description="Recibos cancelados aparecerão aqui para acompanhamento e retomada."
      />
    </section>

    <PaginationControls :current-page="currentPage" :pages="pages" @next="nextPage" @previous="previousPage" />
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
