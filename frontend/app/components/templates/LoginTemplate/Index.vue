<script lang="ts">
import { defineComponent } from "vue";
import LoginForm from "../../organisms/LoginForm/Index.vue";

export default defineComponent({
  name: "LoginTemplate",
  components: {
    LoginForm,
  },
});
</script>

<template>
  <LoginForm />
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
