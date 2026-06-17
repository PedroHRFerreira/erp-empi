<script lang="ts">
import { defineComponent } from "vue";

export default defineComponent({
  name: "BrandMark",
  props: {
    initials: {
      type: String,
      default: "EA",
    },
    title: {
      type: String,
      required: true,
    },
    subtitle: {
      type: String,
      default: "",
    },
    to: {
      type: String,
      default: "",
    },
  },
});
</script>

<template>
  <NuxtLink v-if="to" class="brand-mark" :to="to">
    <span class="brand-mark__badge">{{ initials }}</span>
    <span class="brand-mark__content">
      <strong>{{ title }}</strong>
      <small v-if="subtitle">{{ subtitle }}</small>
    </span>
  </NuxtLink>

  <div v-else class="brand-mark">
    <span class="brand-mark__badge">{{ initials }}</span>
    <span class="brand-mark__content">
      <strong>{{ title }}</strong>
      <small v-if="subtitle">{{ subtitle }}</small>
    </span>
  </div>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
