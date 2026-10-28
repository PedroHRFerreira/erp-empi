<script setup lang="ts">
import { ArrowLeft, ChevronRight, Search, Truck } from '@lucide/vue'
import PageHeader from '../../molecules/PageHeader/Index.vue'
import { formatCurrency } from '../../../utils/format'

const stock = useStockStore()
const query = ref('')
const failed = ref(false)
async function load() {
  failed.value = !await stock.loadSuppliers()
}
await load()

const visibleSuppliers = computed(() => {
  const term = query.value.trim().toLocaleLowerCase('pt-BR')
  return stock.suppliers.filter((supplier) => supplier.supplierName.toLocaleLowerCase('pt-BR').includes(term))
})
</script>

<template>
  <section class="page suppliers-page">
    <PageHeader title="Fornecedores" subtitle="Consulte os produtos, entradas e pagamentos de cada fornecedor.">
      <template #actions>
        <NuxtLink class="button button--secondary" to="/stock"><ArrowLeft :size="18" /> Estoque</NuxtLink>
      </template>
    </PageHeader>

    <label class="panel supplier-search">
      <span>Buscar fornecedor</span>
      <span class="supplier-search__control"><Search :size="18" /><input v-model="query" type="search" placeholder="Nome do fornecedor"></span>
    </label>

    <section v-if="failed" class="panel supplier-state" role="alert">
      <strong>Não foi possível carregar os fornecedores</strong>
      <span>{{ stock.error }}</span>
      <button class="button button--secondary" type="button" :disabled="stock.loading" @click="load">Tentar novamente</button>
    </section>
    <section v-else-if="stock.loading" class="panel supplier-state" role="status"><strong>Carregando fornecedores...</strong></section>
    <section v-else-if="!stock.suppliers.length" class="panel supplier-state">
      <Truck :size="28" /><strong>Nenhum fornecedor registrado</strong>
      <span>Os fornecedores aparecem aqui quando você registra uma entrada de produto.</span>
      <NuxtLink class="button button--primary" to="/stock/products">Ver produtos</NuxtLink>
    </section>
    <section v-else-if="!visibleSuppliers.length" class="panel supplier-state">
      <Search :size="28" /><strong>Nenhum fornecedor encontrado</strong>
      <button class="button button--secondary" type="button" @click="query = ''">Limpar busca</button>
    </section>
    <section v-else class="supplier-grid" aria-label="Fornecedores cadastrados">
      <NuxtLink v-for="supplier in visibleSuppliers" :key="supplier.supplierName" class="panel supplier-card" :to="{ path: '/stock/suppliers/detail', query: { supplierName: supplier.supplierName } }">
        <header><Truck :size="22" /><h2>{{ supplier.supplierName }}</h2><ChevronRight :size="20" /></header>
        <p>{{ supplier.productCount }} {{ supplier.productCount === 1 ? 'produto' : 'produtos' }} · {{ supplier.purchaseCount }} {{ supplier.purchaseCount === 1 ? 'entrada' : 'entradas' }}</p>
        <dl>
          <div><dt>Total comprado</dt><dd>{{ formatCurrency(supplier.totalCents) }}</dd></div>
          <div><dt>Pago</dt><dd>{{ formatCurrency(supplier.paidCents) }}</dd></div>
          <div><dt>Pendente</dt><dd>{{ formatCurrency(supplier.pendingCents) }}</dd></div>
        </dl>
        <span class="supplier-card__action">Ver produtos e histórico</span>
      </NuxtLink>
    </section>
  </section>
</template>

<style scoped lang="scss">
@use 'styles.module.scss';
</style>
