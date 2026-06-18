<script lang="ts">
import { defineComponent } from 'vue'
import ExpensesTemplate from '../components/templates/ExpensesTemplate/Index.vue'

export default defineComponent({
  name: 'ExpensesPage',
  components: {
    ExpensesTemplate
  },
  async setup() {
    const expenses = useExpensesStore()
    await expenses.load()

    return {}
  }
})
</script>

<template>
  <ExpensesTemplate />
</template>
