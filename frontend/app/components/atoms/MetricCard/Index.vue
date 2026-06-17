<script lang="ts">
import { defineComponent, type PropType } from "vue";

export default defineComponent({
  name: "MetricCard",
  props: {
    icon: {
      type: [Object, Function] as PropType<unknown>,
      required: true,
    },
    label: {
      type: String,
      required: true,
    },
    value: {
      type: [String, Number],
      required: true,
    },
  },
});
</script>

<template>
  <article class="metric-card panel">
    <component :is="icon" :size="22" />
    <span>{{ label }}</span>
    <strong>{{ value }}</strong>
  </article>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
