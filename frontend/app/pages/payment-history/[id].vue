<script setup lang="ts">
import {
  AlertTriangle,
  ArrowLeft,
  CalendarDays,
  CheckCircle2,
  Clock3,
  Package,
  ReceiptText,
  RotateCcw,
  WalletCards,
} from "@lucide/vue";
import { computed, onMounted, ref } from "vue";
import type { IPayableInstallment } from "../../../server/contracts/types";
import PageHeader from "../../components/molecules/PageHeader/Index.vue";
import { formatCurrency, formatDate } from "../../utils/format";

const route = useRoute();
const purchases = usePurchasesStore();
const systemFeedback = useSystemFeedback();
const loading = ref(true);
const loadError = ref("");
const revoking = ref(false);

const installmentId = computed(() => String(route.params.id || ""));
const history = computed(() => purchases.paymentHistory);
const current = computed(() => history.value?.installment);
const origin = computed(
  () =>
    history.value?.expense?.description ||
    history.value?.purchase?.supplierName ||
    "Origem não informada",
);
const originKind = computed(() =>
  history.value?.expense ? "Gasto operacional" : "Compra de estoque",
);
const paidInstallments = computed(
  () =>
    history.value?.installments.filter((row) => row.status === "paid") || [],
);
const remainingInstallments = computed(
  () =>
    history.value?.installments.filter((row) => row.status === "pending") || [],
);
const totalCents = computed(
  () =>
    history.value?.installments.reduce(
      (sum, row) => sum + row.amountCents,
      0,
    ) || 0,
);
const paidCents = computed(() =>
  paidInstallments.value.reduce((sum, row) => sum + row.amountCents, 0),
);
const remainingCents = computed(() =>
  remainingInstallments.value.reduce((sum, row) => sum + row.amountCents, 0),
);
const canRevoke = computed(
  () => current.value?.status === "paid" && !current.value.paymentRevokedAt,
);

async function loadPage() {
  loading.value = true;
  loadError.value = "";
  purchases.error = "";
  purchases.paymentHistory = null;
  if (
    !installmentId.value ||
    !(await purchases.loadPaymentHistory(installmentId.value))
  ) {
    loadError.value =
      purchases.error ||
      "Não foi possível carregar o histórico deste pagamento.";
  }
  loading.value = false;
}

function methodLabel(method?: string) {
  if (!method) return "Não informada";
  return (
    (
      {
        boleto: "Boleto",
        cash: "Dinheiro",
        pix: "PIX",
        debit_card: "Débito",
        credit_card: "Crédito",
        legacy: "Legado",
      } as Record<string, string>
    )[method] || method
  );
}

function statusLabel(row: IPayableInstallment) {
  return (
    { pending: "Pendente", paid: "Pago", cancelled: "Cancelado" } as const
  )[row.status];
}

async function requestRevocation() {
  if (!current.value || revoking.value) return;
  const confirmed = await systemFeedback.confirm({
    title: "Revogar este pagamento?",
    message: "O pagamento voltará para pendente, permitindo editar ou remover os itens vinculados.",
    warning: "Esta revogação só pode ser realizada uma única vez. Depois que a conta for paga novamente, ela não poderá ser revogada outra vez.",
    tone: "danger",
    confirmLabel: "Revogar pagamento",
    cancelLabel: "Manter pagamento",
  });
  if (!confirmed) return;
  revoking.value = true;
  const success = await purchases.revokePayment(current.value.id);
  revoking.value = false;
  if (!success) {
    systemFeedback.error("Não foi possível revogar o pagamento", purchases.error);
    return;
  }
  systemFeedback.success("Pagamento revogado", "A parcela voltou para pendente e não poderá ser revogada novamente após uma nova quitação.");
}

onMounted(loadPage);
</script>

<template>
  <section class="page payment-history-page">
    <NuxtLink class="back-link" to="/payables"
      ><ArrowLeft :size="17" /> Voltar para contas a pagar</NuxtLink
    >
    <PageHeader
      title="Histórico do pagamento"
      subtitle="Consulte a origem, as parcelas e os registros desta conta antes de realizar uma revogação."
    />

    <div v-if="loadError" class="alert-box" role="alert">
      <AlertTriangle :size="19" />
      <span>{{ loadError }}</span>
      <button
        class="button button--secondary"
        type="button"
        :disabled="loading"
        @click="loadPage"
      >
        Tentar novamente
      </button>
    </div>

    <section v-if="loading" class="panel empty-state" aria-live="polite">
      <WalletCards :size="32" /><strong>Carregando histórico...</strong
      ><span>Aguarde enquanto buscamos os dados do pagamento.</span>
    </section>

    <template v-else-if="history && current">
      <section class="summary-grid" aria-label="Resumo do pagamento">
        <article class="summary-card">
          <span>Valor total</span
          ><strong>{{ formatCurrency(totalCents) }}</strong
          ><small>{{ history.installments.length }} parcela(s)</small>
        </article>
        <article class="summary-card summary-card--paid">
          <span>Total pago</span><strong>{{ formatCurrency(paidCents) }}</strong
          ><small>{{ paidInstallments.length }} quitada(s)</small>
        </article>
        <article class="summary-card">
          <span>Saldo restante</span
          ><strong>{{ formatCurrency(remainingCents) }}</strong
          ><small>{{ remainingInstallments.length }} pendente(s)</small>
        </article>
      </section>

      <section class="panel origin-panel">
        <header>
          <div class="section-icon">
            <Package v-if="history.purchase" :size="22" /><ReceiptText
              v-else
              :size="22"
            />
          </div>
          <div>
            <span class="eyebrow">Origem da conta</span>
            <h2>{{ origin }}</h2>
            <p>{{ originKind }}</p>
          </div>
        </header>
        <dl class="detail-grid">
          <div v-if="history.purchase">
            <dt>Data da compra</dt>
            <dd>{{ formatDate(history.purchase.purchasedAt) }}</dd>
          </div>
          <div v-if="history.expense">
            <dt>Data do gasto</dt>
            <dd>{{ formatDate(history.expense.spentAt) }}</dd>
          </div>
          <div v-if="history.expense">
            <dt>Categoria</dt>
            <dd>{{ history.expense.category || "Não informada" }}</dd>
          </div>
          <div>
            <dt>Parcela consultada</dt>
            <dd>{{ current.number }} de {{ history.installments.length }}</dd>
          </div>
          <div>
            <dt>Vencimento</dt>
            <dd>{{ formatDate(current.dueDate) }}</dd>
          </div>
          <div>
            <dt>Forma prevista</dt>
            <dd>{{ methodLabel(current.plannedMethod) }}</dd>
          </div>
        </dl>
        <div v-if="history.purchase?.items?.length" class="items-list">
          <h3>Itens adicionados ao estoque</h3>
          <ul>
            <li v-for="item in history.purchase.items" :key="item.id">
              <span
                >{{ item.stockItem?.name || "Produto" }} ·
                {{ item.quantity }} un.</span
              ><strong>{{ formatCurrency(item.subtotalCents) }}</strong>
            </li>
          </ul>
        </div>
      </section>

      <section class="panel installments-panel">
        <header>
          <div>
            <span class="eyebrow">Linha do tempo</span>
            <h2>Parcelas e pagamentos</h2>
          </div>
          <span class="count">{{ history.installments.length }}</span>
        </header>
        <div class="installment-list">
          <article
            v-for="row in history.installments"
            :key="row.id"
            :class="{ 'installment--current': row.id === current.id }"
            class="installment"
          >
            <div class="timeline-icon">
              <CheckCircle2 v-if="row.status === 'paid'" :size="20" /><Clock3
                v-else
                :size="20"
              />
            </div>
            <div class="installment__main">
              <strong>Parcela {{ row.number }}</strong
              ><span>Vencimento em {{ formatDate(row.dueDate) }}</span
              ><small v-if="row.id === current.id">Pagamento selecionado</small>
            </div>
            <dl>
              <div>
                <dt>Situação</dt>
                <dd :class="`status status--${row.status}`">
                  {{ statusLabel(row) }}
                </dd>
              </div>
              <div>
                <dt>Valor</dt>
                <dd class="amount">{{ formatCurrency(row.amountCents) }}</dd>
              </div>
              <div>
                <dt>Pagamento</dt>
                <dd>{{ row.paidAt ? formatDate(row.paidAt) : "—" }}</dd>
              </div>
              <div>
                <dt>Meio</dt>
                <dd>
                  {{ methodLabel(row.paymentMethod || row.plannedMethod) }}
                </dd>
              </div>
            </dl>
          </article>
        </div>
      </section>

      <section class="panel revocation-panel">
        <div>
          <span class="eyebrow">Correção de pagamento</span>
          <h2>Revogar pagamento da parcela {{ current.number }}</h2>
          <p v-if="canRevoke">
            Use esta ação somente se a quitação foi registrada por engano. A
            parcela voltará ao estado pendente para permitir a edição ou remoção
            da origem.
          </p>
          <p v-else-if="current.paymentRevokedAt">
            Este pagamento já utilizou sua única revogação. Se for quitado
            novamente, a ação não ficará disponível.
          </p>
          <p v-else>Somente pagamentos quitados podem ser revogados.</p>
        </div>
        <button
          v-if="canRevoke"
          class="button button--danger revoke-button"
          type="button"
          :disabled="revoking"
          @click="requestRevocation"
        >
          <RotateCcw :size="17" /> Revogar pagamento
        </button>
      </section>

    </template>
  </section>
</template>

<style scoped lang="scss">
.payment-history-page {
  min-width: 0;
  gap: 20px;
}
.back-link {
  display: inline-flex;
  align-items: center;
  align-self: start;
  gap: 7px;
  color: var(--watt-text-muted);
  font-weight: 700;
  text-decoration: none;
}
.back-link:hover {
  color: var(--watt-data);
}
.summary-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}
.summary-card {
  display: grid;
  gap: 7px;
  padding: 20px;
  border: 1px solid var(--watt-border);
  border-radius: 16px;
  background: var(--watt-surface);
}
.summary-card > span,
.eyebrow,
dt {
  color: var(--watt-text-muted);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}
.summary-card strong {
  font:
    700 clamp(22px, 2vw, 30px) "Fira Code",
    monospace;
}
.summary-card small,
.origin-panel p,
.installment__main span,
.revocation-panel p {
  color: var(--watt-text-muted);
}
.summary-card--paid {
  border-left: 4px solid var(--watt-success);
}
.summary-card--paid strong {
  color: var(--watt-success);
}
.origin-panel,
.installments-panel,
.revocation-panel {
  padding: 20px;
}
.origin-panel > header,
.installments-panel > header,
.revocation-panel {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}
.origin-panel > header {
  justify-content: flex-start;
}
.origin-panel h2,
.installments-panel h2,
.revocation-panel h2 {
  margin: 3px 0;
}
.origin-panel p,
.revocation-panel p {
  margin: 4px 0 0;
  line-height: 1.5;
}
.section-icon {
  display: grid;
  width: 46px;
  height: 46px;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 12px;
  color: var(--watt-data);
  background: var(--braip-theme-surface-muted);
}
.detail-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
  margin: 20px 0 0;
  padding-top: 20px;
  border-top: 1px solid var(--watt-border);
}
.detail-grid div,
.installment dl div {
  display: grid;
  gap: 5px;
}
.detail-grid dd,
.installment dd {
  margin: 0;
}
.items-list {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid var(--watt-border);
}
.items-list h3 {
  margin: 0 0 12px;
}
.items-list ul {
  display: grid;
  gap: 10px;
  margin: 0;
  padding: 0;
  list-style: none;
}
.items-list li {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 11px 12px;
  border-radius: 10px;
  background: var(--watt-surface-raised);
}
.count {
  display: grid;
  min-width: 34px;
  height: 34px;
  place-items: center;
  border-radius: 9px;
  background: var(--watt-surface-raised);
}
.installment-list {
  display: grid;
  margin-top: 18px;
}
.installment {
  display: grid;
  grid-template-columns: auto minmax(150px, 1fr) minmax(420px, 2fr);
  align-items: center;
  gap: 14px;
  padding: 16px 8px;
  border-top: 1px solid var(--watt-border);
}
.installment--current {
  margin-inline: -8px;
  padding-inline: 16px;
  border-radius: 12px;
  background: var(--watt-surface-raised);
}
.timeline-icon {
  display: grid;
  width: 40px;
  height: 40px;
  place-items: center;
  border-radius: 50%;
  color: var(--watt-text-muted);
  background: var(--watt-surface-raised);
}
.installment:has(.status--paid) .timeline-icon {
  color: var(--watt-success);
}
.installment__main {
  display: grid;
  gap: 4px;
}
.installment__main small {
  color: var(--watt-data);
  font-weight: 700;
}
.installment dl {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
  margin: 0;
}
.amount {
  font:
    700 14px "Fira Code",
    monospace;
}
.status {
  justify-self: start;
  padding: 5px 8px;
  border-radius: 8px;
  background: var(--watt-surface-raised);
  font-size: 12px;
}
.status--paid {
  color: var(--watt-success);
}
.status--cancelled {
  color: var(--watt-alert);
}
.revocation-panel > div {
  max-width: 760px;
}
.revoke-button {
  flex: 0 0 auto;
}
.alert-box,
.success-box {
  display: flex;
  align-items: center;
  gap: 9px;
  margin: 0;
  padding: 14px 16px;
  border-left: 4px solid var(--watt-alert);
  color: var(--watt-alert);
  background: var(--watt-alert-background);
}
.alert-box span {
  flex: 1;
}
.success-box {
  border-left-color: var(--watt-success);
  color: var(--watt-success);
  background: color-mix(in srgb, var(--watt-success) 10%, var(--watt-surface));
}
.empty-state {
  display: grid;
  min-height: 260px;
  place-items: center;
  align-content: center;
  gap: 8px;
  color: var(--watt-text-muted);
  text-align: center;
}
@media (max-width: 900px) {
  .installment {
    grid-template-columns: auto 1fr;
  }
  .installment dl {
    grid-column: 2;
    grid-template-columns: 1fr 1fr;
  }
  .detail-grid {
    grid-template-columns: 1fr 1fr;
  }
  .revocation-panel {
    align-items: flex-start;
    flex-direction: column;
  }
}
@media (max-width: 650px) {
  .summary-grid,
  .detail-grid {
    grid-template-columns: 1fr;
  }
  .origin-panel,
  .installments-panel,
  .revocation-panel {
    padding: 16px;
  }
  .installment {
    grid-template-columns: auto 1fr;
    padding-inline: 0;
  }
  .installment--current {
    margin-inline: 0;
    padding-inline: 10px;
  }
  .installment dl {
    grid-column: 1/-1;
    grid-template-columns: 1fr 1fr;
  }
  .revoke-button {
    width: 100%;
  }
  .items-list li {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
