<script lang="ts">
import { defineComponent } from "vue";
import EmptyState from "../../molecules/EmptyState/Index.vue";

export default defineComponent({
  name: "MetricsTableSection",
  components: {
    EmptyState,
  },
  props: {
    title: {
      type: String,
      required: true,
    },
    to: {
      type: String,
      required: true,
    },
    empty: {
      type: Boolean,
      required: true,
    },
    emptyTitle: {
      type: String,
      required: true,
    },
    emptyDescription: {
      type: String,
      required: true,
    },
  },
});
</script>

<template>
  <section class="metrics-table-section panel table-wrap">
    <header class="metrics-table-section__header">
      <h2>{{ title }}</h2>
      <NuxtLink class="button button--secondary" :to="to">Ver mais</NuxtLink>
    </header>

    <slot />

    <EmptyState
      v-if="empty"
      :description="emptyDescription"
      :title="emptyTitle"
    />
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
