<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'PaginationControls',
  props: {
    currentPage: {
      type: Number,
      required: true
    },
    pages: {
      type: Number,
      required: true
    }
  },
  emits: ['previous', 'next'],
  setup(_, { emit }) {
    function previous() {
      emit('previous')
    }

    function next() {
      emit('next')
    }

    return {
      previous,
      next
    }
  }
})
</script>

<template>
  <footer class="pagination-controls">
    <button class="button button--secondary" :disabled="currentPage <= 1" type="button" @click="previous">
      Anterior
    </button>
    <span>Página {{ currentPage }} de {{ pages || 1 }}</span>
    <button class="button button--secondary" :disabled="currentPage >= pages" type="button" @click="next">
      Próxima
    </button>
  </footer>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
