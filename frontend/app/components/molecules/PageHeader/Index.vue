<script lang="ts">
import { defineComponent } from "vue";

export default defineComponent({
  name: "PageHeader",
  props: {
    title: {
      type: String,
      required: true,
    },
    subtitle: {
      type: String,
      default: "",
    },
  },
});
</script>

<template>
  <header class="page__header">
    <div>
      <h1 class="page__title">{{ title }}</h1>
      <p v-if="subtitle" class="page__subtitle">{{ subtitle }}</p>
    </div>

    <slot name="actions" />
  </header>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
