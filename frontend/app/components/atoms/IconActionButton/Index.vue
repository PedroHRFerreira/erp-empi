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
  <button class="button icon-action-button" :class="`button--${variant}`" :title="title" type="button" @click="handleClick">
    <slot />
  </button>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
