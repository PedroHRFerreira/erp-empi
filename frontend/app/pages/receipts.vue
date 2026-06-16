<script setup lang="ts">
import { CheckCircle2, Copy, MessageCircle, Printer, Plus } from '@lucide/vue'
import type { IReceipt, IStockItem } from '../../server/contracts/types'
import { centsFromCurrency, formatCurrency } from '../utils/format'

const receipts = useReceiptsStore()
const stock = useStockStore()
const showForm = ref(false)
const selectedStockId = ref('')
const selectedQuantity = ref(1)
const priceInput = ref('')
const form = reactive({
  client: { name: '', cpf: '', phone: '', email: '' },
  vehicleModel: '',
  vehicleYear: new Date().getFullYear(),
  vehiclePlate: '',
  services: '',
  priceCents: 0,
  notes: '',
  items: [] as Array<{ stockItemId: string; quantity: number }>
})

await Promise.all([receipts.load(), stock.load()])

const pages = computed(() => Math.ceil(receipts.total / receipts.limit))
const currentPage = computed(() => Math.floor(receipts.offset / receipts.limit) + 1)

function stockName(id: string) {
  return stock.items.find((item) => item.id === id)?.name || 'Produto'
}

function addItem() {
  if (!selectedStockId.value || selectedQuantity.value <= 0) return
  form.items.push({ stockItemId: selectedStockId.value, quantity: selectedQuantity.value })
  selectedStockId.value = ''
  selectedQuantity.value = 1
}

function removeItem(index: number) {
  form.items.splice(index, 1)
}

async function createReceipt() {
  form.priceCents = centsFromCurrency(priceInput.value)
  await receipts.create(form)
  Object.assign(form, {
    client: { name: '', cpf: '', phone: '', email: '' },
    vehicleModel: '',
    vehicleYear: new Date().getFullYear(),
    vehiclePlate: '',
    services: '',
    priceCents: 0,
    notes: '',
    items: []
  })
  priceInput.value = ''
  showForm.value = false
}

function printReceipt(receipt: IReceipt) {
  const html = `
    <main style="font-family:Arial;padding:32px">
      <h1>Recibo EMPI Autocenter</h1>
      <p><strong>Cliente:</strong> ${receipt.user.name}</p>
      <p><strong>Veiculo:</strong> ${receipt.vehicleModel} ${receipt.vehicleYear} - ${receipt.vehiclePlate}</p>
      <p><strong>Servicos:</strong> ${receipt.services}</p>
      <p><strong>Valor:</strong> ${formatCurrency(receipt.priceCents)}</p>
      <p><strong>Status:</strong> ${receipt.status}</p>
    </main>`
  const popup = window.open('', '_blank', 'noopener,noreferrer')
  if (!popup) return
  popup.document.write(html)
  popup.document.close()
  popup.print()
}
</script>

<template>
  <section class="page">
    <header class="page__header">
      <div>
        <h1 class="page__title">Recibos</h1>
        <p class="page__subtitle">Crie orcamentos, acompanhe pagamentos e baixe produtos do estoque.</p>
      </div>
      <button class="button button--primary" type="button" @click="showForm = !showForm">
        <Plus :size="18" />
        Adicionar
      </button>
    </header>

    <form v-if="showForm" class="receipt-form panel" @submit.prevent="createReceipt">
      <label class="field">
        <span>Cliente</span>
        <input v-model="form.client.name" required placeholder="Nome do cliente" />
      </label>
      <label class="field">
        <span>CPF</span>
        <input v-model="form.client.cpf" required inputmode="numeric" placeholder="000.000.000-00" />
      </label>
      <label class="field">
        <span>Telefone</span>
        <input v-model="form.client.phone" inputmode="numeric" placeholder="33900000000" />
      </label>
      <label class="field">
        <span>E-mail</span>
        <input v-model="form.client.email" type="email" placeholder="cliente@email.com" />
      </label>
      <label class="field">
        <span>Veiculo</span>
        <input v-model="form.vehicleModel" required placeholder="Modelo" />
      </label>
      <label class="field">
        <span>Ano</span>
        <input v-model.number="form.vehicleYear" required type="number" min="1950" />
      </label>
      <label class="field">
        <span>Placa</span>
        <input v-model="form.vehiclePlate" required placeholder="ABC1D23" />
      </label>
      <label class="field">
        <span>Valor</span>
        <input v-model="priceInput" required inputmode="decimal" placeholder="35000 para R$ 350,00" />
      </label>
      <label class="field receipt-form__wide">
        <span>Servicos</span>
        <textarea v-model="form.services" required placeholder="Troca de oleo, diagnostico eletrico..." />
      </label>

      <div class="receipt-form__wide receipt-items">
        <div class="receipt-items__row">
          <label class="field">
            <span>Produto usado</span>
            <select v-model="selectedStockId">
              <option value="">Selecione</option>
              <option v-for="item in stock.items" :key="item.id" :value="item.id">
                {{ item.name }} - disponivel {{ item.quantity }}
              </option>
            </select>
          </label>
          <label class="field">
            <span>Quantidade</span>
            <input v-model.number="selectedQuantity" type="number" min="1" />
          </label>
          <button class="button button--secondary" type="button" @click="addItem">Adicionar item</button>
        </div>
        <ul v-if="form.items.length">
          <li v-for="(item, index) in form.items" :key="`${item.stockItemId}-${index}`">
            {{ stockName(item.stockItemId) }} x {{ item.quantity }}
            <button type="button" @click="removeItem(index)">remover</button>
          </li>
        </ul>
      </div>

      <label class="field receipt-form__wide">
        <span>Observacoes</span>
        <textarea v-model="form.notes" placeholder="Informacoes adicionais" />
      </label>

      <p v-if="receipts.error" class="form-error receipt-form__wide">{{ receipts.error }}</p>
      <button class="button button--primary receipt-form__wide" type="submit">Salvar recibo</button>
    </form>

    <section class="panel table-wrap">
      <table>
        <thead>
          <tr>
            <th>Cliente</th>
            <th>Veiculo</th>
            <th>Valor</th>
            <th>Status</th>
            <th>Acoes</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="receipt in receipts.receipts" :key="receipt.id">
            <td>{{ receipt.user.name }}</td>
            <td>{{ receipt.vehicleModel }} / {{ receipt.vehiclePlate }}</td>
            <td>{{ formatCurrency(receipt.priceCents) }}</td>
            <td>
              <span class="badge" :class="receipt.status === 'paid' ? 'badge--paid' : 'badge--pending'">
                {{ receipt.status === 'paid' ? 'Pago' : 'Pendente' }}
              </span>
            </td>
            <td>
              <div class="actions">
                <button class="button button--secondary" title="WhatsApp" type="button" @click="receipts.shareWhatsApp(receipt)">
                  <MessageCircle :size="16" />
                </button>
                <button class="button button--secondary" title="Copiar para Instagram" type="button" @click="receipts.copyInstagramText(receipt)">
                  <Copy :size="16" />
                </button>
                <button class="button button--secondary" title="PDF / imprimir" type="button" @click="printReceipt(receipt)">
                  <Printer :size="16" />
                </button>
                <button
                  v-if="receipt.status !== 'paid'"
                  class="button button--primary"
                  title="Efetuou pagamento"
                  type="button"
                  @click="receipts.markPaid(receipt.id)"
                >
                  <CheckCircle2 :size="16" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="!receipts.receipts.length" class="empty">Nenhum recibo encontrado.</div>
    </section>

    <footer class="pagination">
      <button class="button button--secondary" :disabled="currentPage <= 1" @click="receipts.load(receipts.offset - receipts.limit)">
        Anterior
      </button>
      <span>Pagina {{ currentPage }} de {{ pages || 1 }}</span>
      <button
        class="button button--secondary"
        :disabled="currentPage >= pages"
        @click="receipts.load(receipts.offset + receipts.limit)"
      >
        Proxima
      </button>
    </footer>
  </section>
</template>

<style scoped lang="scss">
.receipt-form {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  padding: 18px;
}

.receipt-form__wide {
  grid-column: 1 / -1;
}

.receipt-items {
  display: grid;
  gap: 10px;
}

.receipt-items__row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 140px auto;
  align-items: end;
  gap: 12px;
}

.receipt-items ul {
  margin: 0;
  padding-left: 18px;
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
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
}

@media (max-width: 980px) {
  .receipt-form,
  .receipt-items__row {
    grid-template-columns: 1fr;
  }
}
</style>
