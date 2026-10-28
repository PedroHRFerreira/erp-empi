<script setup lang="ts">
import { ArrowLeft, ArrowRight, Package, Truck } from '@lucide/vue'
import { computed, ref, watch } from 'vue'
import type { IStockItem } from '../../../../server/contracts/types'
import PageHeader from '../../molecules/PageHeader/Index.vue'
import PaginationControls from '../../molecules/PaginationControls/Index.vue'
import StockTable from '../../organisms/StockTable/Index.vue'
import StockToolbar from '../../organisms/StockToolbar/Index.vue'
import { printStockReport } from '../../../utils/print'

const props = withDefaults(defineProps<{ view?: 'home' | 'products' }>(), { view: 'home' })
const router = useRouter()
const stock = useStockStore()
const purchases = usePurchasesStore()
const feedback = useSystemFeedback()
const pages = computed(() => Math.ceil(stock.total / stock.limit))
const currentPage = computed(() => Math.floor(stock.offset / stock.limit) + 1)
const searchInput = ref(stock.searchQuery)
async function clearSearch() {
  searchInput.value = ''
  await stock.applySearch('')
}
watch(() => props.view, (view) => { if (view === 'products') purchases.loadPurchases() }, { immediate: true })
function startEdit(item: IStockItem) { return router.push(`/stock/${item.id}/edit`) }
async function removeProduct(item: IStockItem) {
  const related = purchases.purchases.filter((purchase) => purchase.status === 'confirmed' && purchase.items.some((entry) => entry.stockItemId === item.id))
  if (related.some((purchase) => purchase.installments.some((installment) => installment.status === 'paid'))) {
    feedback.warning('Produto não removido', 'Este produto possui uma parcela paga. O histórico financeiro precisa ser preservado.')
    return
  }
  if (related.some((purchase) => purchase.items.length > 1)) {
    feedback.warning('Produto não removido', 'Este produto pertence a uma compra antiga com outros itens e não pode ser removido isoladamente.')
    return
  }
  if (!await feedback.confirm({ title: 'Remover produto?', message: `Remover ${item.name}, suas entradas e todas as parcelas pendentes?`, tone: 'danger', confirmLabel: 'Remover produto' })) return
  const result = await stock.remove(item.id)
  if (result.status === 'error') { feedback.error('Não foi possível remover o produto', stock.error); return }
  await purchases.loadPurchases()
  feedback.success('Produto removido', 'As contas pendentes vinculadas foram canceladas.')
}
</script>

<template>
  <section class="page stock-page">
    <PageHeader :title="view === 'home' ? 'Estoque' : 'Produtos'" :subtitle="view === 'home' ? 'Escolha como acompanhar seu estoque.' : 'Todos os produtos e seus saldos, reunindo as entradas de todos os fornecedores.'">
      <template #actions>
        <NuxtLink v-if="view === 'home'" class="button button--primary" to="/stock/new">Adicionar produto</NuxtLink>
        <StockToolbar v-else @add="router.push('/stock/new')" @export="stock.exportCsv" @print="printStockReport(stock.items)" />
      </template>
    </PageHeader>
    <div v-if="view === 'home'" class="stock-access">
      <NuxtLink class="panel stock-access__card" to="/stock/suppliers">
        <span class="stock-access__icon"><Truck :size="28" /></span>
        <div><h2>Fornecedores</h2><p>Consulte produtos, compras e pagamentos de cada fornecedor e adicione mais quantidade.</p></div>
        <span class="stock-access__link">Ver fornecedores <ArrowRight :size="18" /></span>
      </NuxtLink>
      <NuxtLink class="panel stock-access__card" to="/stock/products">
        <span class="stock-access__icon"><Package :size="28" /></span>
        <div><h2>Produtos</h2><p>Acompanhe o saldo de todos os produtos, edite seus dados e registre novas entradas.</p></div>
        <span class="stock-access__link">Ver produtos <ArrowRight :size="18" /></span>
      </NuxtLink>
    </div>
    <template v-else>
      <NuxtLink class="stock-back" to="/stock"><ArrowLeft :size="17" /> Estoque</NuxtLink>
      <form class="panel stock-filters" aria-label="Filtros de produtos" @submit.prevent="stock.applySearch(searchInput)">
        <fieldset :disabled="stock.loading">
          <label class="field"><span>Buscar produto</span><input v-model="searchInput" type="search" placeholder="Digite o nome do produto" /></label>
          <div class="stock-filters__actions">
            <button class="button button--secondary" type="button" @click="clearSearch">Limpar filtro</button>
            <button class="button button--primary" type="submit">Filtrar</button>
          </div>
        </fieldset>
      </form>
      <p v-if="stock.error || purchases.error" class="alert-box" role="alert">{{ stock.error || purchases.error }}</p>
      <p v-if="stock.loading" class="stock-loading" role="status">Carregando produtos...</p>
      <div v-else-if="stock.searchQuery && !stock.items.length && !stock.error" class="panel stock-filter-empty" role="status"><strong>Nenhum produto encontrado</strong><p>Tente outro nome ou limpe o filtro.</p></div>
      <StockTable v-else :items="stock.items" :purchases="purchases.purchases" @edit="startEdit" @remove="removeProduct" />
      <p v-if="!stock.loading" class="stock-loading" role="status">{{ stock.total }} {{ stock.total === 1 ? 'produto encontrado' : 'produtos encontrados' }}</p>
      <PaginationControls :current-page="currentPage" :pages="pages" @next="stock.load(stock.offset + stock.limit)" @previous="stock.load(stock.offset - stock.limit)" />
    </template>
  </section>
</template>

<style scoped lang="scss">
@use "styles.module.scss";
</style>
