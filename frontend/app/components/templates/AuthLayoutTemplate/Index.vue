<script lang="ts">
import { defineComponent } from "vue";

export default defineComponent({
  name: "AuthLayoutTemplate",
});
</script>

<template>
  <main class="auth-layout-template">
    <slot />
  </main>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
