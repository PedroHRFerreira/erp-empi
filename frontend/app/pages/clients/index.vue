<script lang="ts">
import { defineComponent } from "vue";
import ClientsTemplate from "../../components/templates/ClientsTemplate/Index.vue";

export default defineComponent({
  name: "ClientsPage",
  components: {
    ClientsTemplate,
  },
  async setup() {
    const clients = useClientsStore();
    await clients.load();

    return {};
  },
});
</script>

<template>
  <ClientsTemplate />
</template>
