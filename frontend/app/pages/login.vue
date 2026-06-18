<script lang="ts">
import { defineComponent } from "vue";
import LoginTemplate from "../components/templates/LoginTemplate/Index.vue";

export default defineComponent({
  name: "LoginPage",
  components: {
    LoginTemplate,
  },
});
</script>

<template>
  <LoginTemplate />
</template>
