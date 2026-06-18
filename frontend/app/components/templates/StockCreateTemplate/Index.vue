<script lang="ts">
import { ArrowLeft } from '@lucide/vue'
import { computed, defineComponent, reactive, ref } from 'vue'
import type { StockForm as IStockForm } from '../../../stores/useStockStore'
import { currencyMaskToCents } from '../../../utils/masks'
import PageHeader from '../../molecules/PageHeader/Index.vue'
import StockForm from '../../organisms/StockForm/Index.vue'

export default defineComponent({
  name: 'StockCreateTemplate',
  components: {
    ArrowLeft,
    PageHeader,
    StockForm
  },
  setup() {
    const auth = useAuthStore()
    const router = useRouter()
    const stock = useStockStore()
    const costInput = ref('')
    const defaultMarkupPercent = computed(() => {
      const markupPercent = Number(auth.user?.markupPercent ?? 10)
      return Number.isFinite(markupPercent) ? markupPercent : 10
    })
    const form = reactive<IStockForm>({
      name: '',
      description: '',
      costCents: 0,
      markupPercent: defaultMarkupPercent.value,
      quantity: 0
    })

    stock.error = ''
    stock.fieldErrors = {}

    async function save() {
      form.costCents = currencyMaskToCents(costInput.value)
      const result = await stock.save({ ...form })

      if (result.status === 'success') {
        await router.push('/stock')
      }
    }

    return {
      costInput,
      form,
      save,
      stock
    }
  }
})
</script>

<template>
  <section class="page">
    <PageHeader title="Adicionar produto" subtitle="Cadastre itens de estoque para uso e revenda em recibos.">
      <template #actions>
        <NuxtLink class="button button--secondary" to="/stock">
          <ArrowLeft :size="18" />
          Estoque
        </NuxtLink>
      </template>
    </PageHeader>

    <StockForm
      :cost-input="costInput"
      :error="stock.error"
      :field-errors="stock.fieldErrors"
      :form="form"
      @clear-field-error="stock.clearFieldError"
      @save="save"
      @update:cost-input="(value) => (costInput = value)"
    />
  </section>
</template>
