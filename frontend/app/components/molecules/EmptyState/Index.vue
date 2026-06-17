<script lang="ts">
import { Inbox } from "@lucide/vue";
import { defineComponent } from "vue";

export default defineComponent({
  name: "EmptyState",
  components: {
    Inbox,
  },
  props: {
    title: {
      type: String,
      required: true,
    },
    description: {
      type: String,
      default: "",
    },
  },
});
</script>

<template>
  <div class="empty-state">
    <div class="empty-state__icon" aria-hidden="true">
      <Inbox :size="22" />
    </div>
    <strong>{{ title }}</strong>
    <p v-if="description">{{ description }}</p>
    <slot />
  </div>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
