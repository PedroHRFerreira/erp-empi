<script lang="ts">
import { defineComponent } from 'vue'
import ClientDetailTemplate from '../../components/templates/ClientDetailTemplate/Index.vue'

export default defineComponent({
  name: 'ClientDetailPage',
  components: {
    ClientDetailTemplate
  },
  async setup() {
    const route = useRoute()
    const clients = useClientsStore()
    await clients.loadDetail(String(route.params.id))

    return {}
  }
})
</script>

<template>
  <ClientDetailTemplate />
</template>

<style scoped lang="scss">
@use "../styles.module.scss";
</style>
