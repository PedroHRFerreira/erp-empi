export interface IUser {
  id: string
  name: string
  cpf: string
  type: 'admin' | 'client'
  email: string
  phone: string
  markupPercent: number
  machineFeePercent: number
  address: string
  notes: string
  archivedAt?: string
  createdAt: string
  updatedAt: string
}

export interface IAuthTokens {
  accessToken: string
  refreshToken: string
}

export interface ILoginResponse {
  user: IUser
  tokens: IAuthTokens
}

export interface IPaginated<T> {
  data: T[]
  total: number
  limit: number
  offset: number
}

export interface IStockItem {
  id: string
  name: string
  description: string
  costCents: number
  markupPercent: number
  resalePriceCents: number
  quantity: number
  usedQuantity: number
  active: boolean
  createdAt: string
  updatedAt: string
}

export interface IReceiptItem {
  id: string
  stockItemId: string
  quantity: number
  unitCostCents: number
  unitResaleCents: number
  markupPercent: number
  stockItem?: IStockItem
}

export interface IReceipt {
  id: string
  userId: string
  user: IUser
  vehicleModel: string
  vehicleYear: number
  vehiclePlate: string
  services: string
  laborPriceCents: number
  productsTotalCents: number
  subtotalCents: number
  cardFeePercent: number
  cardFeeCents: number
  paymentMethod: 'credit_card' | 'debit_card' | 'pix' | 'cash'
  installments: number
  priceCents: number
  status: 'pending' | 'paid' | 'cancelled'
  notes: string
  paidAt?: string
  items: IReceiptItem[]
  createdAt: string
  updatedAt: string
}

export interface IClientDetail {
  client: IUser
  receipts: IReceipt[]
}

export interface IMetricsSummary {
  clientsTotal: number
  receiptsTotal: number
  receiptsPaid: number
  receiptsPending: number
  receiptsCancelled: number
  revenuePaidCents: number
  revenuePendingCents: number
  averageTicketPaidCents: number
  stockItemsTotal: number
  stockUnitsAvailableTotal: number
  stockUnitsUsedTotal: number
  lastReceipt: { id: string; clientName: string; priceCents: number; status: string; createdAt: string } | null
  lastStockItem: { id: string; name: string; quantity: number; usedQuantity: number; createdAt: string } | null
  topProducts: Array<{ id: string; name: string; usedQuantity: number }>
  lowStockProducts: Array<{ id: string; name: string; quantity: number; usedQuantity: number; createdAt: string }>
  recentClients: Array<{ id: string; name: string; cpf: string; receiptsCount: number; lastReceiptAt: string }>
  pendingReceipts: Array<{ id: string; clientName: string; priceCents: number; status: string; createdAt: string }>
  paidReceipts: Array<{ id: string; clientName: string; priceCents: number; status: string; createdAt: string }>
}
