<script setup lang="ts">
import { Download, Edit3, FileDown, Plus, Trash2 } from '@lucide/vue'
import type { IStockItem } from '../../server/contracts/types'
import { centsFromCurrency, formatCurrency } from '../utils/format'

const stock = useStockStore()
await stock.load()

const showForm = ref(false)
const costInput = ref('')
const form = reactive({
  id: '',
  name: '',
  description: '',
  costCents: 0,
  markupPercent: 10,
  quantity: 0
})

const pages = computed(() => Math.ceil(stock.total / stock.limit))
const currentPage = computed(() => Math.floor(stock.offset / stock.limit) + 1)

function startCreate() {
  Object.assign(form, { id: '', name: '', description: '', costCents: 0, markupPercent: 10, quantity: 0 })
  costInput.value = ''
  showForm.value = true
}

function startEdit(item: IStockItem) {
  Object.assign(form, {
    id: item.id,
    name: item.name,
    description: item.description,
    costCents: item.costCents,
    markupPercent: item.markupPercent,
    quantity: item.quantity
  })
  costInput.value = String(item.costCents)
  showForm.value = true
}

async function save() {
  form.costCents = centsFromCurrency(costInput.value)
  await stock.save({ ...form, id: form.id || undefined })
  showForm.value = false
}

function printStock() {
  window.print()
}
</script>

<template>
  <section class="page">
    <header class="page__header">
      <div>
        <h1 class="page__title">Estoque</h1>
        <p class="page__subtitle">Controle produtos, markup de revenda e uso em recibos.</p>
      </div>
      <div class="toolbar">
        <button class="button button--secondary" type="button" @click="stock.exportCsv">
          <Download :size="18" />
          Excel
        </button>
        <button class="button button--secondary" type="button" @click="printStock">
          <FileDown :size="18" />
          PDF
        </button>
        <button class="button button--primary" type="button" @click="startCreate">
          <Plus :size="18" />
          Adicionar
        </button>
      </div>
    </header>

    <form v-if="showForm" class="stock-form panel" @submit.prevent="save">
      <label class="field">
        <span>Produto</span>
        <input v-model="form.name" required placeholder="Oleo, filtro, pastilha..." />
      </label>
      <label class="field">
        <span>Custo em centavos</span>
        <input v-model="costInput" required inputmode="numeric" placeholder="12000 para R$ 120,00" />
      </label>
      <label class="field">
        <span>Markup (%)</span>
        <input v-model.number="form.markupPercent" required type="number" min="0" step="0.1" />
      </label>
      <label class="field">
        <span>Quantidade</span>
        <input v-model.number="form.quantity" required type="number" min="0" />
      </label>
      <label class="field stock-form__wide">
        <span>Descricao</span>
        <textarea v-model="form.description" placeholder="Detalhes do produto" />
      </label>
      <p v-if="stock.error" class="form-error stock-form__wide">{{ stock.error }}</p>
      <button class="button button--primary stock-form__wide" type="submit">Salvar produto</button>
    </form>

    <section class="panel table-wrap">
      <table>
        <thead>
          <tr>
            <th>Produto</th>
            <th>Custo</th>
            <th>Revenda estimada</th>
            <th>Qtd.</th>
            <th>Usados</th>
            <th>Acoes</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in stock.items" :key="item.id">
            <td>{{ item.name }}</td>
            <td>{{ formatCurrency(item.costCents) }}</td>
            <td>{{ formatCurrency(item.resalePriceCents) }}</td>
            <td>{{ item.quantity }}</td>
            <td>{{ item.usedQuantity }}</td>
            <td>
              <div class="actions">
                <button class="button button--secondary" title="Editar" type="button" @click="startEdit(item)">
                  <Edit3 :size="16" />
                </button>
                <button class="button button--danger" title="Excluir" type="button" @click="stock.remove(item.id)">
                  <Trash2 :size="16" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="!stock.items.length" class="empty">Nenhum produto cadastrado.</div>
    </section>

    <footer class="pagination">
      <button class="button button--secondary" :disabled="currentPage <= 1" @click="stock.load(stock.offset - stock.limit)">
        Anterior
      </button>
      <span>Pagina {{ currentPage }} de {{ pages || 1 }}</span>
      <button class="button button--secondary" :disabled="currentPage >= pages" @click="stock.load(stock.offset + stock.limit)">
        Proxima
      </button>
    </footer>
  </section>
</template>

<style scoped lang="scss">
.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.stock-form {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  padding: 18px;
}

.stock-form__wide {
  grid-column: 1 / -1;
}

.form-error {
  margin: 0;
  color: var(--status-danger);
  font-weight: 700;
}

.actions {
  display: flex;
  gap: 8px;
}

.actions .button {
  width: 38px;
  min-height: 38px;
  padding: 0;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  align-items: center;
}

@media (max-width: 980px) {
  .stock-form {
    grid-template-columns: 1fr;
  }
}
</style>
