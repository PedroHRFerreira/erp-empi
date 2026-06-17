<script lang="ts">
import { Edit3, Trash2 } from '@lucide/vue'
import { defineComponent, type PropType } from 'vue'
import type { IStockItem } from '../../../../server/contracts/types'
import IconActionButton from '../../atoms/IconActionButton/Index.vue'
import EmptyState from '../../molecules/EmptyState/Index.vue'
import { formatCurrency, formatDateTime } from '../../../utils/format'

export default defineComponent({
  name: 'StockTable',
  components: {
    Edit3,
    EmptyState,
    IconActionButton,
    Trash2
  },
  props: {
    items: {
      type: Array as PropType<IStockItem[]>,
      required: true
    }
  },
  emits: ['edit', 'remove'],
  setup(_, { emit }) {
    function edit(item: IStockItem) {
      emit('edit', item)
    }

    function remove(item: IStockItem) {
      emit('remove', item)
    }

    function resaleTotalCents(item: IStockItem) {
      return item.resalePriceCents * item.quantity
    }

    return {
      edit,
      formatCurrency,
      formatDateTime,
      remove,
      resaleTotalCents
    }
  }
})
</script>

<template>
  <section class="panel table-wrap">
    <table class="stock-table">
      <thead>
        <tr>
          <th>Produto</th>
          <th>Custo</th>
          <th>Revenda estimada (un.)</th>
          <th>Revenda estimada (total)</th>
          <th>Qtd.</th>
          <th>Usados</th>
          <th>Cadastrado em</th>
          <th>Ações</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in items" :key="item.id">
          <td>{{ item.name }}</td>
          <td>{{ formatCurrency(item.costCents) }}</td>
          <td>{{ formatCurrency(item.resalePriceCents) }}</td>
          <td>{{ formatCurrency(resaleTotalCents(item)) }}</td>
          <td>{{ item.quantity }}</td>
          <td>{{ item.usedQuantity }}</td>
          <td>{{ formatDateTime(item.createdAt) }}</td>
          <td>
            <div class="stock-table__actions">
              <IconActionButton title="Editar" @click="edit(item)">
                <Edit3 :size="16" />
              </IconActionButton>
              <IconActionButton title="Excluir" variant="danger" @click="remove(item)">
                <Trash2 :size="16" />
              </IconActionButton>
            </div>
          </td>
        </tr>
      </tbody>
    </table>

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
