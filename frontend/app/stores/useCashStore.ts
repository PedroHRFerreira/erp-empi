import { defineStore } from "pinia";
import type {
  ICashBalances,
  ICashEntry,
  ICashSession,
  IReceipt,
} from "../../server/contracts/types";

type PaymentMethod = IReceipt["paymentMethod"];

export const useCashStore = defineStore("cash", {
  state: () => ({
    session: null as ICashSession | null,
    sessions: [] as ICashSession[],
    dailyEntries: [] as ICashEntry[],
    balances: {
      pixCents: 0,
      debitCardCents: 0,
      creditCardCents: 0,
    } as ICashBalances,
    loading: false,
    error: "",
  }),
  actions: {
    async load(forceRefresh = false) {
      if (forceRefresh) {
        invalidateApiCache([
          "/cash/current",
          "/cash/sessions",
          "/cash/daily-entries",
          "/cash/balances",
        ]);
      }
      this.loading = true;
      const [current, sessions, dailyEntries, balances] = await Promise.all([
        useApiFetch<ICashSession | null>("/cash/current"),
        useApiFetch<ICashSession[]>("/cash/sessions"),
        useApiFetch<ICashEntry[]>("/cash/daily-entries"),
        useApiFetch<ICashBalances>("/cash/balances"),
      ]);
      this.loading = false;
      if (
        current.status.value === "error" ||
        sessions.status.value === "error" ||
        dailyEntries.status.value === "error" ||
        balances.status.value === "error"
      ) {
        this.error = "Não foi possível carregar o caixa.";
        return false;
      }
      this.session = current.data.value || null;
      this.sessions = Array.isArray(sessions.data.value) ? sessions.data.value : [];
      this.dailyEntries = Array.isArray(dailyEntries.data.value) ? dailyEntries.data.value : [];
      this.balances = balances.data.value || {
        pixCents: 0,
        debitCardCents: 0,
        creditCardCents: 0,
      };
      this.error = "";
      return true;
    },
    async open(openingCashCents: number) {
      const { status } = await useApiFetch("/cash/open", {
        method: "POST",
        body: { openingCashCents },
      });
      if (status.value === "error") {
        this.error =
          "Não foi possível abrir o caixa. Feche o dia anterior antes de abrir um novo.";
        return false;
      }
      await this.load(true);
      return true;
    },
    async close(closingCashCents: number, closingNotes: string) {
      const { status } = await useApiFetch("/cash/close", {
        method: "POST",
        body: { closingCashCents, closingNotes },
      });
      if (status.value === "error") {
        this.error =
          "Informe uma justificativa caso haja divergência e tente novamente.";
        return false;
      }
      await this.load(true);
      return true;
    },
    async addAdjustment(input: {
      direction: "in" | "out";
      paymentMethod: PaymentMethod;
      amountCents: number;
      description: string;
      reason: string;
    }) {
      const { status } = await useApiFetch("/cash/adjustments", {
        method: "POST",
        body: input,
      });
      if (status.value === "error") {
        this.error = input.paymentMethod === "cash"
          ? "Preencha os dados e mantenha o caixa aberto para ajustar Dinheiro."
          : "Preencha todos os dados do ajuste.";
        return false;
      }
      await this.load(true);
      return true;
    },
  },
});
