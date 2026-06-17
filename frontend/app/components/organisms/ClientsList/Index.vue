<script lang="ts">
import { Trash2 } from '@lucide/vue'
import { defineComponent, type PropType } from 'vue'
import type { IUser } from '../../../../server/contracts/types'
import { formatCpf } from '../../../utils/format'
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
  emits: ['remove'],
  setup(_, { emit }) {
    function remove(client: IUser) {
      emit('remove', client)
    }

    function clientPhone(phone: string) {
      return phone ? maskPhone(phone) : '-'
    }

    return {
      clientPhone,
      formatCpf,
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
          <th>CPF</th>
          <th>Telefone</th>
          <th>E-mail</th>
          <th class="clients-list__actions-heading">Ações</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="client in clients" :key="client.id">
          <td>
            <NuxtLink class="clients-list__name" :to="`/clients/${client.id}`">
              {{ client.name }}
            </NuxtLink>
          </td>
          <td>{{ client.cpf ? formatCpf(client.cpf) : '-' }}</td>
          <td>{{ clientPhone(client.phone) }}</td>
          <td>{{ client.email || '-' }}</td>
          <td class="clients-list__actions-cell">
            <IconActionButton title="Remover" variant="danger" @click="remove(client)">
              <Trash2 :size="16" />
            </IconActionButton>
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
