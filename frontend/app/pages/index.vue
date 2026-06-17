<script lang="ts">
import { defineComponent } from 'vue'
import MetricsTemplate from '../components/templates/MetricsTemplate/Index.vue'

export default defineComponent({
  name: 'MetricsPage',
  components: {
    MetricsTemplate
  },
  async setup() {
    const metrics = useMetricsStore()
    await metrics.load()

    return {}
  }
})
</script>

<template>
  <MetricsTemplate />
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
