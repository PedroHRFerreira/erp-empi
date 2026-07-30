<script lang="ts">
import { defineComponent } from 'vue'
import GoalsTemplate from '../components/templates/GoalsTemplate/Index.vue'

export default defineComponent({
  name: 'GoalsPage',
  components: { GoalsTemplate },
  async setup() {
    await useGoalsStore().load()
    return {}
  }
})
</script>

<template><GoalsTemplate /></template>
