<script lang="ts">
import { defineComponent } from "vue";
import ProfileTemplate from "../components/templates/ProfileTemplate/Index.vue";

export default defineComponent({
  name: "ProfilePage",
  components: {
    ProfileTemplate,
  },
});
</script>

<template>
  <ProfileTemplate />
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
