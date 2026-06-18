<script lang="ts">
import { Trash2 } from '@lucide/vue'
import { defineComponent, type PropType } from 'vue'
import type { IUser } from '../../../../server/contracts/types'
import { maskPhone } from '../../../utils/masks'
import IconActionButton from '../../atoms/IconActionButton/Index.vue'
import EmptyState from '../../molecules/EmptyState/Index.vue'

export default defineComponent({
  name: 'ClientsList',
  components: {
    EmptyState,
    IconActionButton,
    Trash2
  },
  props: {
    clients: {
      type: Array as PropType<IUser[]>,
      required: true
    }
  },
  emits: ['open-detail', 'remove'],
  setup(_, { emit }) {
    function openDetail(client: IUser) {
      emit('open-detail', client)
    }

    function remove(client: IUser) {
      emit('remove', client)
    }

    function clientPhone(phone: string) {
      return phone ? maskPhone(phone) : '-'
    }

    return {
      clientPhone,
      openDetail,
      remove
    }
  }
})
</script>

<template>
  <section class="panel table-wrap">
    <table class="clients-list">
      <thead>
        <tr>
          <th>Cliente</th>
          <th>Telefone</th>
          <th class="clients-list__actions-heading">Ações</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="client in clients" :key="client.id">
          <td>
            <a class="clients-list__name" :href="`/clients/${client.id}`" @click.prevent="openDetail(client)">
              {{ client.name }}
            </a>
          </td>
          <td>{{ clientPhone(client.phone) }}</td>
          <td>
            <div class="clients-list__actions">
              <IconActionButton title="Remover" variant="danger" @click="remove(client)">
                <Trash2 :size="16" />
              </IconActionButton>
            </div>
          </td>
        </tr>
      </tbody>
    </table>

    <EmptyState
      v-if="!clients.length"
      title="Nenhum cliente encontrado"
      description="Clientes aparecem aqui automaticamente quando um recibo é criado."
    />
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
