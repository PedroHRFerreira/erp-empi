export interface IUser {
  id: string
  name: string
  cpf: string
  type: 'admin' | 'client'
  email: string
  phone: string
  markupPercent: number
  address: string
  notes: string
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
  priceCents: number
  status: 'pending' | 'paid' | 'cancelled'
  notes: string
  paidAt?: string
  items: IReceiptItem[]
  createdAt: string
  updatedAt: string
}

export interface IMetricsSummary {
  clientsTotal: number
  receiptsTotal: number
  receiptsPaid: number
  receiptsPending: number
  revenuePaidCents: number
  topProducts: Array<{ id: string; name: string; usedQuantity: number }>
  recentClients: Array<{ id: string; name: string; cpf: string; receiptsCount: number; lastReceiptAt: string }>
  pendingReceipts: Array<{ id: string; clientName: string; priceCents: number; status: string; createdAt: string }>
  paidReceipts: Array<{ id: string; clientName: string; priceCents: number; status: string; createdAt: string }>
}
