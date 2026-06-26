<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'IconActionButton',
  props: {
    title: {
      type: String,
      required: true
    },
    variant: {
      type: String,
      default: 'secondary'
    }
  },
  emits: ['click'],
  setup(_, { emit }) {
    function handleClick() {
      emit('click')
    }

    return {
      handleClick
    }
  }
})
</script>

<template>
  <span class="icon-action-button__wrap">
    <button class="button icon-action-button" :aria-label="title" :class="`button--${variant}`" type="button" @click="handleClick">
      <slot />
    </button>
    <span class="icon-action-button__popper" role="tooltip">{{ title }}</span>
  </span>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
