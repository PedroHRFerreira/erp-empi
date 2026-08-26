<script lang="ts">
import { computed, defineComponent, reactive, ref } from 'vue'
import type { IStockItem } from '../../../../server/contracts/types'
import type { StockForm as IStockForm } from '../../../stores/useStockStore'
import { currencyMaskToCents, formatCentsAsCurrency } from '../../../utils/masks'
import PageHeader from '../../molecules/PageHeader/Index.vue'
import PaginationControls from '../../molecules/PaginationControls/Index.vue'
import StockForm from '../../organisms/StockForm/Index.vue'
import StockTable from '../../organisms/StockTable/Index.vue'
import StockToolbar from '../../organisms/StockToolbar/Index.vue'
import { printStockReport } from '../../../utils/print'

export default defineComponent({
  name: 'StockTemplate',
  components: {
    PageHeader,
    PaginationControls,
    StockForm,
    StockTable,
    StockToolbar,
  },
  setup() {
    const auth = useAuthStore()
    const router = useRouter()
    const stock = useStockStore()
    const purchases = usePurchasesStore()
    const feedback = useSystemFeedback()
    const showForm = ref(false)
    const costInput = ref('')
    const defaultMarkupPercent = computed(() => {
      const markupPercent = Number(auth.user?.markupPercent ?? 10)
      return Number.isFinite(markupPercent) ? markupPercent : 10
    })
    const form = reactive<IStockForm>({
      id: '',
      name: '',
      description: '',
      costCents: 0,
      markupPercent: defaultMarkupPercent.value,
      quantity: 0
    })
    const pages = computed(() => Math.ceil(stock.total / stock.limit))
    const currentPage = computed(() => Math.floor(stock.offset / stock.limit) + 1)

    purchases.loadPurchases()

    function startCreate() {
      return router.push('/stock/new')
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
      costInput.value = formatCentsAsCurrency(item.costCents)
      stock.error = ''
      stock.fieldErrors = {}
      showForm.value = true
    }

    function addStock(item: IStockItem) {
      return router.push(`/stock/purchases/new?productId=${item.id}`)
    }

    async function removeProduct(item: IStockItem) {
      const related = purchases.purchases.filter((purchase) =>
        purchase.status === 'confirmed' && purchase.items.some((entry) => entry.stockItemId === item.id)
      )
      if (related.some((purchase) => purchase.installments.some((installment) => installment.status === 'paid'))) {
        feedback.warning('Produto não removido', 'Este produto possui uma parcela paga. O histórico financeiro precisa ser preservado.')
        return
      }
      if (related.some((purchase) => purchase.items.length > 1)) {
        feedback.warning('Produto não removido', 'Este produto pertence a uma compra antiga com outros itens e não pode ser removido isoladamente.')
        return
      }
      if (!await feedback.confirm({ title: 'Remover produto?', message: `Remover ${item.name}, suas entradas e todas as parcelas pendentes?`, tone: 'danger', confirmLabel: 'Remover produto' })) return
      for (const purchase of related) {
        if (!await purchases.cancelPurchase(purchase.id)) {
          feedback.error('Não foi possível remover o produto', purchases.error)
          return
        }
      }
      const result = await stock.remove(item.id)
      if (result.status === 'error') feedback.error('Não foi possível remover o produto', stock.error)
    }

    async function save() {
      form.costCents = currencyMaskToCents(costInput.value)

      const result = await stock.save({ ...form, id: form.id || undefined })

      if (result.status === 'success') {
        showForm.value = false
      }
    }

    function printStock() {
      printStockReport(stock.items)
    }

    function previousPage() {
      return stock.load(stock.offset - stock.limit)
    }

    function nextPage() {
      return stock.load(stock.offset + stock.limit)
    }

    return {
      costInput,
      currentPage,
      form,
      addStock,
      nextPage,
      pages,
      previousPage,
      printStock,
      purchases,
      removeProduct,
      save,
      showForm,
      startCreate,
      startEdit,
      stock
    }
  }
})
</script>

<template>
  <section class="page">
    <PageHeader title="Estoque" subtitle="Cadastre produtos, registre entradas e acompanhe fornecedores e pagamentos.">
      <template #actions>
        <StockToolbar @add="startCreate" @export="stock.exportCsv" @print="printStock" />
      </template>
    </PageHeader>

    <StockForm
      v-if="showForm"
      :cost-input="costInput"
      :error="stock.error"
      :field-errors="stock.fieldErrors"
      :form="form"
      @clear-field-error="stock.clearFieldError"
      @save="save"
      @update:cost-input="(value) => (costInput = value)"
    />

    <p v-if="stock.error || purchases.error" class="alert-box" role="alert">{{ stock.error || purchases.error }}</p>
    <StockTable :items="stock.items" :purchases="purchases.purchases" @add-stock="addStock" @edit="startEdit" @remove="removeProduct" />

    <PaginationControls :current-page="currentPage" :pages="pages" @next="nextPage" @previous="previousPage" />
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
