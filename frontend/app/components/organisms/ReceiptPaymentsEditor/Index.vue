<script setup lang="ts">
import { computed, ref } from "vue";
import type { ReceiptForm } from "../../../stores/useReceiptsStore";
import { formatCurrency } from "../../../utils/format";
import {
  currencyMaskToCents,
  formatCentsAsCurrency,
  maskCurrency,
} from "../../../utils/masks";
import {
  calculateReceiptPayments,
  validateReceiptPayments,
  type ReceiptPaymentForm,
} from "../../../utils/receiptPayments";

const props = defineProps<{
  form: ReceiptForm;
  subtotalCents: number;
  fieldErrors: Record<string, string>;
}>();
const emit = defineEmits<{ "clear-field-error": [field: string] }>();
const auth = useAuthStore();
const methods: { value: ReceiptPaymentForm["paymentMethod"]; label: string }[] =
  [
    { value: "cash", label: "Dinheiro" },
    { value: "pix", label: "PIX" },
    { value: "debit_card", label: "Cartão de débito" },
    { value: "credit_card", label: "Cartão de crédito" },
  ];
const installmentOptions = Array.from({ length: 12 }, (_, index) => index + 1);
const customFees = ref(
  props.form.payments.map(
    (payment) => payment.cardFeePercent !== defaultFee(payment),
  ),
);
const divided = computed(() => props.form.payments.length === 2);
const calculation = computed(() =>
  calculateReceiptPayments(props.form.payments, props.subtotalCents),
);
const errors = computed(() => ({
  ...props.fieldErrors,
  ...validateReceiptPayments(props.form.payments, props.subtotalCents),
}));

function isCard(payment: ReceiptPaymentForm) {
  return (
    payment.paymentMethod === "credit_card" ||
    payment.paymentMethod === "debit_card"
  );
}

function defaultFee(payment: ReceiptPaymentForm) {
  if (!isCard(payment)) return 0;
  return payment.paymentMethod === "credit_card" && payment.installments > 1
    ? Number(
        auth.user?.installmentFeePercent ??
          props.form.installmentFeePercent ??
          0,
      )
    : Number(auth.user?.machineFeePercent ?? props.form.machineFeePercent ?? 0);
}

function clearErrors() {
  for (const key of Object.keys(props.fieldErrors)) {
    if (
      key === "payments" ||
      key.startsWith("payments.") ||
      ["paymentMethod", "cardFeePercent", "installments"].includes(key)
    ) {
      emit("clear-field-error", key);
    }
  }
}

function fieldError(index: number, field: string) {
  return (
    errors.value[`payments.${index}.${field}`] ||
    (index === 0 &&
    field === "baseAmountCents" &&
    divided.value &&
    props.form.payments[0]!.baseAmountCents >= props.subtotalCents
      ? errors.value["payments.1.baseAmountCents"]
      : "") ||
    (index === 0 ? props.fieldErrors[field] : "")
  );
}

function toggleDivision(event: Event) {
  const enabled = (event.target as HTMLInputElement).checked;
  const first = props.form.payments[0]!;
  if (enabled) {
    first.baseAmountCents = Math.floor(props.subtotalCents / 2);
    props.form.payments.push({
      paymentMethod: first.paymentMethod === "pix" ? "cash" : "pix",
      baseAmountCents: props.subtotalCents - first.baseAmountCents,
      cardFeePercent: 0,
      installments: 1,
    });
    customFees.value[1] = false;
  } else {
    props.form.payments.splice(1);
    first.baseAmountCents = props.subtotalCents;
    customFees.value.splice(1);
  }
  clearErrors();
}

function updateMethod(index: number, event: Event) {
  const payment = props.form.payments[index]!;
  payment.paymentMethod = (event.target as HTMLSelectElement)
    .value as ReceiptPaymentForm["paymentMethod"];
  payment.installments = 1;
  payment.cardFeePercent = defaultFee(payment);
  customFees.value[index] = false;
  clearErrors();
}

function updateBase(event: Event) {
  const input = event.target as HTMLInputElement;
  input.value = maskCurrency(input.value);
  props.form.payments[0]!.baseAmountCents = currencyMaskToCents(input.value);
  clearErrors();
}

function updateInstallments(index: number, event: Event) {
  const payment = props.form.payments[index]!;
  payment.installments = Number((event.target as HTMLSelectElement).value);
  if (!customFees.value[index]) payment.cardFeePercent = defaultFee(payment);
  clearErrors();
}

function updateFee(index: number, event: Event) {
  props.form.payments[index]!.cardFeePercent = Number(
    (event.target as HTMLInputElement).value,
  );
  customFees.value[index] = true;
  clearErrors();
}

function installmentSummary(amount: number, installments: number) {
  const regular = Math.floor(amount / installments);
  const last = amount - regular * (installments - 1);
  if (regular === last) return `${installments}x de ${formatCurrency(regular)}`;
  return `${installments - 1}x de ${formatCurrency(regular)} + última de ${formatCurrency(last)}`;
}
</script>

<template>
  <section class="receipt-payments" aria-label="Pagamento do recibo">
    <h3>Pagamento</h3>
    <label class="receipt-payments__toggle">
      <input
        type="checkbox"
        :checked="divided"
        aria-describedby="receipt-payments-hint"
        @change="toggleDivision"
      />
      <span>Dividir em duas formas</span>
    </label>
    <p v-if="errors.payments" class="field__error" role="alert">
      {{ errors.payments }}
    </p>

    <section
      v-for="(payment, index) in form.payments"
      :key="index"
      class="receipt-payments__part"
      :aria-label="`Pagamento ${index + 1}`"
    >
      <h4 v-if="divided">Pagamento {{ index + 1 }}</h4>
      <div class="receipt-step__grid receipt-step__grid--two">
        <label
          class="field"
          :class="{ 'field--error': fieldError(index, 'paymentMethod') }"
        >
          <span>{{
            divided ? `Forma de pagamento ${index + 1}` : "Forma de pagamento"
          }}</span>
          <select
            :value="payment.paymentMethod"
            required
            :aria-invalid="Boolean(fieldError(index, 'paymentMethod'))"
            :aria-describedby="
              fieldError(index, 'paymentMethod')
                ? `payment-${index}-method-error`
                : undefined
            "
            @change="updateMethod(index, $event)"
          >
            <option
              v-for="method in methods"
              :key="method.value"
              :value="method.value"
              :disabled="
                form.payments.some(
                  (other, otherIndex) =>
                    otherIndex !== index &&
                    other.paymentMethod === method.value,
                )
              "
            >
              {{ method.label }}
            </option>
          </select>
          <small
            v-if="fieldError(index, 'paymentMethod')"
            :id="`payment-${index}-method-error`"
            class="field__error"
            >{{ fieldError(index, "paymentMethod") }}</small
          >
        </label>

        <label
          class="field"
          :class="{ 'field--error': fieldError(index, 'baseAmountCents') }"
        >
          <span>{{
            divided && index === 1
              ? "Valor restante (sem taxas)"
              : "Valor base (sem taxas)"
          }}</span>
          <input
            :value="
              formatCentsAsCurrency(
                calculation.payments[index]?.baseAmountCents || 0,
              )
            "
            :readonly="!divided || index === 1"
            inputmode="decimal"
            :aria-invalid="Boolean(fieldError(index, 'baseAmountCents'))"
            :aria-describedby="
              fieldError(index, 'baseAmountCents')
                ? `payment-${index}-base-error`
                : undefined
            "
            @input="updateBase"
          />
          <small
            v-if="fieldError(index, 'baseAmountCents')"
            :id="`payment-${index}-base-error`"
            class="field__error"
            >{{ fieldError(index, "baseAmountCents") }}</small
          >
        </label>

        <label
          v-if="payment.paymentMethod === 'credit_card'"
          class="field"
          :class="{ 'field--error': fieldError(index, 'installments') }"
        >
          <span>Parcelas</span>
          <select
            :value="payment.installments"
            :aria-invalid="Boolean(fieldError(index, 'installments'))"
            :aria-describedby="
              fieldError(index, 'installments')
                ? `payment-${index}-installments-error`
                : undefined
            "
            @change="updateInstallments(index, $event)"
          >
            <option
              v-for="installment in installmentOptions"
              :key="installment"
              :value="installment"
            >
              {{ installment
              }}{{ installment === 1 ? " parcela" : " parcelas" }}
            </option>
          </select>
          <small
            v-if="fieldError(index, 'installments')"
            :id="`payment-${index}-installments-error`"
            class="field__error"
            >{{ fieldError(index, "installments") }}</small
          >
        </label>

        <label
          v-if="isCard(payment)"
          class="field"
          :class="{ 'field--error': fieldError(index, 'cardFeePercent') }"
        >
          <span
            >{{
              payment.paymentMethod === "credit_card" &&
              payment.installments > 1
                ? "Juros de parcelamento"
                : "Juros da maquininha"
            }}
            (%)</span
          >
          <input
            :value="payment.cardFeePercent"
            min="0"
            step="0.01"
            type="number"
            :aria-invalid="Boolean(fieldError(index, 'cardFeePercent'))"
            :aria-describedby="
              fieldError(index, 'cardFeePercent')
                ? `payment-${index}-fee-error`
                : undefined
            "
            @input="updateFee(index, $event)"
          />
          <small
            v-if="fieldError(index, 'cardFeePercent')"
            :id="`payment-${index}-fee-error`"
            class="field__error"
            >{{ fieldError(index, "cardFeePercent") }}</small
          >
        </label>
      </div>

      <section
        class="receipt-mini-summary"
        :aria-label="`Valores do pagamento ${index + 1}`"
      >
        <div>
          <span>Valor base</span
          ><strong>{{
            formatCurrency(calculation.payments[index]?.baseAmountCents || 0)
          }}</strong>
        </div>
        <div v-if="isCard(payment)">
          <span>Taxa ({{ payment.cardFeePercent }}%)</span
          ><strong>{{
            formatCurrency(calculation.payments[index]?.cardFeeCents || 0)
          }}</strong>
        </div>
        <div>
          <span>Total desta forma</span
          ><strong>{{
            formatCurrency(calculation.payments[index]?.amountCents || 0)
          }}</strong>
        </div>
        <div
          v-if="payment.paymentMethod === 'credit_card'"
          class="receipt-summary__installment"
        >
          <span>Parcelamento</span
          ><strong>{{
            installmentSummary(
              calculation.payments[index]?.amountCents || 0,
              payment.installments,
            )
          }}</strong>
        </div>
      </section>
    </section>
    <p id="receipt-payments-hint" class="receipt-payments__hint">
      As formas serão confirmadas juntas na quitação do recibo.
    </p>
  </section>
</template>
