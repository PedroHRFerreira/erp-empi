export interface IUser {
  id: string;
  name: string;
  cpf: string;
  type: "admin" | "client";
  email: string;
  phone: string;
  markupPercent: number;
  machineFeePercent: number;
  installmentFeePercent: number;
  address: string;
  notes: string;
  archivedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface IAuthTokens {
  accessToken: string;
  expiresIn: number;
}

export interface ILoginResponse {
  user: IUser;
  tokens: IAuthTokens;
}

export interface IPaginated<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface IStockItem {
  id: string;
  name: string;
  description: string;
  costCents: number;
  markupPercent: number;
  resalePriceCents: number;
  quantity: number;
  usedQuantity: number;
  active: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface IReceiptItem {
  id: string;
  stockItemId: string;
  quantity: number;
  unitCostCents: number;
  unitResaleCents: number;
  markupPercent: number;
  stockItem?: IStockItem;
}

export interface IReceipt {
  id: string;
  userId?: string | null;
  user?: IUser | null;
  vehicleModel: string;
  vehicleYear: number;
  vehiclePlate: string;
  services: string;
  laborPriceCents: number;
  discountCents: number;
  productsTotalCents: number;
  subtotalCents: number;
  cardFeePercent: number;
  cardFeeCents: number;
  paymentMethod: "credit_card" | "debit_card" | "pix" | "cash";
  installments: number;
  priceCents: number;
  status: "pending" | "paid" | "cancelled";
  notes: string;
  paidAt?: string;
  items: IReceiptItem[];
  expenses?: IExpense[];
  createdAt: string;
  updatedAt: string;
}

export interface IClientDetail {
  client: IUser;
  receipts: IReceipt[];
}

export type FinancialHealthStatus = "red" | "yellow" | "green";

export interface IExpense {
  id: string;
  receiptId?: string | null;
  receipt?: IReceipt;
  description: string;
  category: string;
  amountCents: number;
  spentAt: string;
  notes: string;
  affectsProfit?: boolean;
  installments?: IPayableInstallment[];
  archivedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface IExpenseForm {
  id?: string;
  receiptId?: string | null;
  description: string;
  category: string;
  amountCents: number;
  spentAt: string;
  notes: string;
  paymentMethod?: IReceipt["paymentMethod"];
  installments: Array<{
    amountCents: number;
    dueDate: string;
    plannedMethod: PayableMethod;
  }>;
}

export type RealizedExpenseOrigin = "all" | "operational" | "stock";

export interface IRealizedExpense {
  id: string;
  origin: Exclude<RealizedExpenseOrigin, "all">;
  description: string;
  category: string;
  amountCents: number;
  occurredAt: string;
  paymentMethod?: PaymentMethod;
  supplierName?: string;
  installmentNumber?: number;
  expenseId?: string;
  stockPurchaseId?: string;
  installmentId?: string;
  editable: boolean;
}

export interface ICashEntry {
  id: string;
  direction: "in" | "out";
  kind: "receipt_payment" | "expense" | "stock_payment" | "adjustment";
  paymentMethod: IReceipt["paymentMethod"];
  amountCents: number;
  description: string;
  reason: string;
  occurredAt: string;
}

export interface ICashSession {
  id: string;
  businessDate: string;
  status: "open" | "closed";
  openingCashCents: number;
  closingCashCents?: number;
  cashDifferenceCents?: number;
  closingNotes: string;
	  expectedCashCents: number;
  entries?: ICashEntry[];
}

export interface IPayableInstallment {
  id: string;
  stockPurchaseId?: string;
  expenseId?: string;
  number: number;
  amountCents: number;
  dueDate: string;
  status: "pending" | "paid" | "cancelled";
  plannedMethod: PayableMethod;
  paymentMethod?: PaymentMethod;
  paidAt?: string;
  cashEntryId?: string;
  stockPurchase?: IStockPurchase;
  expense?: IExpense;
}

export type PaymentMethod = "credit_card" | "debit_card" | "pix" | "cash" | "legacy";
export type PayableMethod = PaymentMethod | "boleto";

export interface IStockPurchaseItem {
  id: string;
  stockItemId: string;
  quantity: number;
  unitCostCents: number;
  subtotalCents: number;
  stockItem?: IStockItem;
}

export interface IStockPurchase {
  id: string;
  supplierName: string;
  totalCents: number;
  status: "confirmed" | "cancelled";
  purchasedAt: string;
  cancelledAt?: string;
  items: IStockPurchaseItem[];
  installments: IPayableInstallment[];
}

export interface IPayableAlert {
  installmentId: string;
  kind: "overdue" | "due_today" | "due_tomorrow" | "early_payment";
  supplierName: string;
  number: number;
  amountCents: number;
  dueDate: string;
  plannedMethod: PayableMethod;
}

export interface IExpenseCategorySummary {
  category: string;
  amountCents: number;
  count: number;
}

export interface IReceiptCostSummary {
  receiptId: string;
  clientName: string;
  vehicleModel: string;
  vehiclePlate: string;
  serviceExpensesCents: number;
  productCostCents: number;
  totalCostCents: number;
}

export interface IFinancialSummary {
  startDate: string;
  endDate: string;
  paidReceiptsCount: number;
  expensesCount: number;
  revenuePaidCents: number;
  productCostCents: number;
  cardFeesCents: number;
  grossProfitCents: number;
  operationalExpensesCents: number;
  stockExpensesCents: number;
  totalRealizedExpensesCents: number;
  stockPaymentsCount: number;
  operationalProfitCents: number;
  netProfitCents: number;
  netMarginPercent: number;
  healthStatus: FinancialHealthStatus;
  expensesByCategory: IExpenseCategorySummary[];
  receiptCosts: IReceiptCostSummary[];
}

export interface IMetricsSummary {
  clientsTotal: number;
  receiptsTotal: number;
  receiptsPaid: number;
  receiptsPending: number;
  receiptsCancelled: number;
  revenuePaidCents: number;
  revenuePendingCents: number;
  discountsGrantedCents: number;
  receiptsActiveTotalCents: number;
  averageTicketPaidCents: number;
  stockItemsTotal: number;
  stockUnitsAvailableTotal: number;
  stockUnitsUsedTotal: number;
  lastReceipt: {
    id: string;
    clientName: string;
    priceCents: number;
    status: string;
    createdAt: string;
  } | null;
  lastStockItem: {
    id: string;
    name: string;
    quantity: number;
    usedQuantity: number;
    createdAt: string;
  } | null;
  topProducts: Array<{ id: string; name: string; usedQuantity: number }>;
  lowStockProducts: Array<{
    id: string;
    name: string;
    quantity: number;
    usedQuantity: number;
    createdAt: string;
  }>;
  recentClients: Array<{
    id: string;
    name: string;
    receiptsCount: number;
    lastReceiptAt: string;
  }>;
  pendingReceipts: Array<{
    id: string;
    clientName: string;
    priceCents: number;
    status: string;
    createdAt: string;
  }>;
  paidReceipts: Array<{
    id: string;
    clientName: string;
    priceCents: number;
    status: string;
    createdAt: string;
  }>;
}

export interface IGoalMetrics {
  revenueCents: number;
  laborCents: number;
  productsCents: number;
  clients: number;
  appointments: number;
  netProfitCents: number;
  cardFeesCents: number;
  expensesCents: number;
  productCostCents: number;
}

export interface IMonthlyGoal {
  month: string;
  revenueTargetCents: number;
  laborTargetCents: number;
  productsTargetCents: number;
  clientsTarget: number;
  netProfitTargetCents: number;
}

export interface IGoalsSummary {
  month: string;
  periodStart: string;
  periodEnd: string;
  saved: boolean;
  targets: IMonthlyGoal;
  previous: IGoalMetrics;
  actual: IGoalMetrics;
  projection: Pick<
    IGoalMetrics,
    | "revenueCents"
    | "laborCents"
    | "productsCents"
    | "clients"
    | "netProfitCents"
  >;
  requirements: {
    averageTicketCents: number;
    laborPerAppointmentCents: number;
    productsPerAppointmentCents: number;
  };
  pendingOpportunityCents: number;
  pricingRecommendations: Array<{
    stockItemId: string;
    name: string;
    previousQuantity: number;
    suggestedQuantity: number;
    unitCostCents: number;
    markupPercent: number;
    currentPriceCents: number;
    minimumPriceCents: number;
    belowMinimum: boolean;
  }>;
  tips: Array<{ kind: string; title: string; description: string }>;
}
