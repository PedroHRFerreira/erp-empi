<script lang="ts">
import { defineComponent } from "vue";
import ReceiptsTemplate from "../components/templates/ReceiptsTemplate/Index.vue";

export default defineComponent({
  name: "ReceiptsPage",
  components: {
    ReceiptsTemplate,
  },
  async setup() {
    const receipts = useReceiptsStore();
    const stock = useStockStore();

    await Promise.all([receipts.load(), stock.load()]);

    return {};
  },
});
</script>

<template>
  <ReceiptsTemplate />
</template>
