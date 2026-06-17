<script lang="ts">
import { defineComponent } from "vue";
import AppSidebar from "../../organisms/AppSidebar/Index.vue";

export default defineComponent({
  name: "DefaultLayoutTemplate",
  components: {
    AppSidebar,
  },
});
</script>

<template>
  <div class="default-layout-template">
    <AppSidebar />

    <main class="default-layout-template__content">
      <slot />
    </main>
  </div>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
