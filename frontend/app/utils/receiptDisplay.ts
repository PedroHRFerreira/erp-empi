import type { IReceipt } from '../../server/contracts/types'

export const QUICK_RECEIPT_CLIENT_LABEL = 'Recibo rápido'
export const QUICK_RECEIPT_VEHICLE_LABEL = 'Sem veículo'
export const EMPTY_RECEIPT_VALUE = '-'

export function receiptClientName(receipt: Pick<IReceipt, 'user'>) {
  return receipt.user?.name?.trim() || QUICK_RECEIPT_CLIENT_LABEL
}

export function receiptClientPhone(receipt: Pick<IReceipt, 'user'>) {
  return receipt.user?.phone || ''
}

export function receiptVehicleName(receipt: Pick<IReceipt, 'vehicleModel' | 'vehicleYear'>) {
  return [receipt.vehicleModel, receipt.vehicleYear || ''].filter(Boolean).join(' ').trim() || QUICK_RECEIPT_VEHICLE_LABEL
}

export function receiptVehiclePlate(receipt: Pick<IReceipt, 'vehiclePlate'>) {
  return receipt.vehiclePlate?.trim() || EMPTY_RECEIPT_VALUE
}

export function receiptVehicleLine(receipt: Pick<IReceipt, 'vehicleModel' | 'vehicleYear' | 'vehiclePlate'>) {
  const vehicle = receiptVehicleName(receipt)
  const plate = receiptVehiclePlate(receipt)

  if (plate === EMPTY_RECEIPT_VALUE) {
    return vehicle
  }
  return `${vehicle} - ${plate}`
}
