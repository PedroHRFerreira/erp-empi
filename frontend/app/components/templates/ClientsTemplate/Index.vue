<script lang="ts">
import { computed, defineComponent } from 'vue'
import type { IUser } from '../../../../server/contracts/types'
import ClientsList from '../../organisms/ClientsList/Index.vue'
import PageHeader from '../../molecules/PageHeader/Index.vue'
import PaginationControls from '../../molecules/PaginationControls/Index.vue'

export default defineComponent({
  name: 'ClientsTemplate',
  components: {
    ClientsList,
    PageHeader,
    PaginationControls
  },
  setup() {
    const clients = useClientsStore()
    const pages = computed(() => Math.ceil(clients.total / clients.limit))
    const currentPage = computed(() => Math.floor(clients.offset / clients.limit) + 1)

    async function remove(client: IUser) {
      const confirmed = window.confirm(`Remover ${client.name} da listagem de clientes? Os recibos serão preservados.`)
      if (!confirmed) return

      await clients.remove(client.id)
    }

    function previousPage() {
      return clients.load(clients.offset - clients.limit)
    }

    function nextPage() {
      return clients.load(clients.offset + clients.limit)
    }

    return {
      clients,
      currentPage,
      nextPage,
      pages,
      previousPage,
      remove
    }
  }
})
</script>

<template>
  <section class="page">
    <PageHeader title="Clientes" subtitle="Clientes cadastrados automaticamente pelos recibos." />

    <div v-if="clients.loading" class="panel empty">Carregando clientes...</div>
    <p v-else-if="clients.error" class="form-error">{{ clients.error }}</p>
    <ClientsList v-else :clients="clients.clients" @remove="remove" />

    <PaginationControls :current-page="currentPage" :pages="pages" @next="nextPage" @previous="previousPage" />
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
