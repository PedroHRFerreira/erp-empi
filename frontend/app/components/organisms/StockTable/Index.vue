<script lang="ts">
import { Edit3, History, PackagePlus, Trash2 } from '@lucide/vue'
import { defineComponent, type PropType } from 'vue'
import type { IStockItem, IStockPurchase } from '../../../../server/contracts/types'
import IconActionButton from '../../atoms/IconActionButton/Index.vue'
import EmptyState from '../../molecules/EmptyState/Index.vue'
import { formatCurrency, formatDateTime } from '../../../utils/format'

export default defineComponent({
  name: 'StockTable',
  components: {
    Edit3,
    EmptyState,
    IconActionButton,
    History,
    PackagePlus,
    Trash2
  },
  props: {
    items: {
      type: Array as PropType<IStockItem[]>,
      required: true
    },
    purchases: { type: Array as PropType<IStockPurchase[]>, required: true }
  },
  emits: ['add-stock', 'edit', 'remove'],
  setup(props, { emit }) {
    function edit(item: IStockItem) {
      emit('edit', item)
    }

    function remove(item: IStockItem) {
      emit('remove', item)
    }

    function addStock(item: IStockItem) { emit('add-stock', item) }

    function entries(item: IStockItem) {
      return props.purchases.filter((purchase) => purchase.items.some((entry) => entry.stockItemId === item.id))
    }

    function lastEntry(item: IStockItem) { return entries(item).find((purchase) => purchase.status === 'confirmed') }

    function paymentLabel(item: IStockItem) {
      const installments = entries(item).filter((purchase) => purchase.status === 'confirmed').flatMap((purchase) => purchase.installments)
      if (!installments.length) return 'Sem contas'
      if (installments.every((row) => row.status === 'paid')) return 'Quitado'
      const pending = installments.filter((row) => row.status === 'pending')
      return `${pending.length} ${pending.length === 1 ? 'parcela pendente' : 'parcelas pendentes'}`
    }

    function resaleTotalCents(item: IStockItem) {
      return item.resalePriceCents * item.quantity
    }

    return {
      addStock,
      edit,
      entries,
      formatCurrency,
      formatDateTime,
      lastEntry,
      paymentLabel,
      remove,
      resaleTotalCents
    }
  }
})
</script>

<template>
  <section class="panel table-wrap" aria-label="Tabela de estoque" tabindex="0">
    <table class="stock-table">
      <thead>
        <tr>
          <th>Produto</th>
          <th>Última entrada</th>
          <th>Fornecedor</th>
          <th>Custo atual</th>
          <th>Revenda estimada (un.)</th>
          <th>Revenda estimada (total)</th>
          <th>Qtd.</th>
          <th>Usados</th>
          <th>Pagamento</th>
          <th>Ações</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in items" :key="item.id">
          <td>{{ item.name }}</td>
          <td>{{ lastEntry(item) ? formatDateTime(lastEntry(item)!.purchasedAt) : 'Sem entrada' }}</td>
          <td>{{ lastEntry(item)?.supplierName || '—' }}</td>
          <td>{{ formatCurrency(item.costCents) }}</td>
          <td>{{ formatCurrency(item.resalePriceCents) }}</td>
          <td>{{ formatCurrency(resaleTotalCents(item)) }}</td>
          <td>{{ item.quantity }}</td>
          <td>{{ item.usedQuantity }}</td>
          <td><span class="stock-status">{{ paymentLabel(item) }}</span></td>
          <td>
            <div class="stock-table__actions">
              <button class="button button--secondary stock-action" type="button" @click="addStock(item)"><PackagePlus :size="16" /> Adicionar estoque</button>
              <NuxtLink class="button button--secondary stock-action" :to="`/stock/${item.id}/history`"><History :size="15" /> Histórico</NuxtLink>
              <IconActionButton title="Editar" @click="edit(item)">
                <Edit3 :size="16" />
              </IconActionButton>
              <IconActionButton title="Remover produto" variant="danger" @click="remove(item)">
                <Trash2 :size="16" />
              </IconActionButton>
            </div>
          </td>
        </tr>
      </tbody>
    </table>

    <div class="stock-cards">
      <article v-for="item in items" :key="item.id" class="stock-card">
        <header><div><span>Produto</span><strong>{{ item.name }}</strong></div><span class="stock-status">{{ paymentLabel(item) }}</span></header>
        <dl>
          <div><dt>Quantidade</dt><dd>{{ item.quantity }} disponíveis</dd></div>
          <div><dt>Custo atual</dt><dd>{{ formatCurrency(item.costCents) }}</dd></div>
          <div><dt>Último fornecedor</dt><dd>{{ lastEntry(item)?.supplierName || 'Sem entrada' }}</dd></div>
          <div><dt>Última entrada</dt><dd>{{ lastEntry(item) ? formatDateTime(lastEntry(item)!.purchasedAt) : '—' }}</dd></div>
          <div><dt>Revenda unitária</dt><dd>{{ formatCurrency(item.resalePriceCents) }}</dd></div>
          <div><dt>Usados</dt><dd>{{ item.usedQuantity }}</dd></div>
        </dl>
        <div class="stock-card__actions">
          <button class="button button--primary" type="button" @click="addStock(item)"><PackagePlus :size="16" /> Adicionar estoque</button>
          <button class="button button--secondary" type="button" @click="edit(item)"><Edit3 :size="16" /> Editar</button>
          <button class="button button--danger" type="button" @click="remove(item)"><Trash2 :size="16" /> Remover</button>
        </div>
        <NuxtLink class="stock-card__history-link" :to="`/stock/${item.id}/history`"><History :size="16" /> Ver histórico completo</NuxtLink>
      </article>
    </div>

    <EmptyState
      v-if="!items.length"
      title="Nenhum produto cadastrado"
      description="Adicione produtos ao estoque para usar nos recibos e acompanhar consumo."
    />
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
