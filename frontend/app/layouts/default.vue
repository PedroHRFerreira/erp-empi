<script lang="ts">
import { defineComponent } from "vue";
import DefaultLayoutTemplate from "../components/templates/DefaultLayoutTemplate/Index.vue";

export default defineComponent({
  name: "DefaultLayout",
  components: {
    DefaultLayoutTemplate,
  },
});
</script>

<template>
  <DefaultLayoutTemplate>
    <slot />
  </DefaultLayoutTemplate>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
