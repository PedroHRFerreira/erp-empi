<script lang="ts">
import { defineComponent } from "vue";
import StockTemplate from "../components/templates/StockTemplate/Index.vue";

export default defineComponent({
  name: "StockPage",
  components: {
    StockTemplate,
  },
  async setup() {
    const stock = useStockStore();
    await stock.load();

    return {};
  },
});
</script>

<template>
  <StockTemplate />
</template>
