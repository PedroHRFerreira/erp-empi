<script lang="ts">
import { defineComponent, watch } from "vue";
import ClientDetailTemplate from "../../components/templates/ClientDetailTemplate/Index.vue";

export default defineComponent({
  name: "ClientDetailPage",
  components: {
    ClientDetailTemplate,
  },
  setup() {
    const route = useRoute();
    const clients = useClientsStore();

    watch(
      () => route.params.id,
      (id) => {
        if (!id) {
          return;
        }

        void clients.loadDetail(String(id));
      },
      { immediate: true },
    );

    return {};
  },
});
</script>

<template>
  <ClientDetailTemplate />
</template>
