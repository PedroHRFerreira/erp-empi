<script lang="ts">
import { Download, FileDown, Plus } from '@lucide/vue'
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'StockToolbar',
  components: {
    Download,
    FileDown,
    Plus
  },
  emits: ['add', 'export', 'print'],
  setup(_, { emit }) {
    function add() {
      emit('add')
    }

    function exportCsv() {
      emit('export')
    }

    function print() {
      emit('print')
    }

    return {
      add,
      exportCsv,
      print
    }
  }
})
</script>

<template>
  <div class="stock-toolbar">
    <button class="button button--secondary" type="button" @click="exportCsv">
      <Download :size="18" />
      Excel
    </button>
    <button class="button button--secondary" type="button" @click="print">
      <FileDown :size="18" />
      PDF
    </button>
    <button class="button button--primary" type="button" @click="add">
      <Plus :size="18" />
      Adicionar
    </button>
  </div>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
