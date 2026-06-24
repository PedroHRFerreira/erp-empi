<script lang="ts">
import { defineComponent, type PropType } from 'vue'

type ReceiptWizardProgressStep = {
  key: string
  title: string
}

export default defineComponent({
  name: 'ReceiptWizardProgress',
  props: {
    activeIndex: {
      type: Number,
      required: true
    },
    steps: {
      type: Array as unknown as PropType<readonly ReceiptWizardProgressStep[]>,
      required: true
    }
  }
})
</script>

<template>
  <nav class="receipt-wizard-progress" aria-label="Etapas do recibo">
    <ol>
      <li
        v-for="(step, index) in steps"
        :key="step.key"
        class="receipt-wizard-progress__step"
        :class="{
          'receipt-wizard-progress__step--active': index === activeIndex,
          'receipt-wizard-progress__step--done': index < activeIndex
        }"
        :aria-current="index === activeIndex ? 'step' : undefined"
      >
        <span>{{ index + 1 }}</span>
        <strong>{{ step.title }}</strong>
      </li>
    </ol>
  </nav>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
