import { describe, expect, it } from 'vitest'
import { currencyMaskToCents, formatCentsAsCurrency, maskCpf, maskCurrency, maskPhone, maskVehiclePlate } from './masks'

describe('mask helpers', () => {
  it('masks cpf', () => {
    expect(maskCpf('52998224725')).toBe('529.982.247-25')
    expect(maskCpf('52998224725999')).toBe('529.982.247-25')
  })

  it('masks phone', () => {
    expect(maskPhone('33987351922')).toBe('(33) 98735-1922')
    expect(maskPhone('3333334444')).toBe('(33) 3333-4444')
  })

  it('masks vehicle plate', () => {
    expect(maskVehiclePlate('abc-1d23')).toBe('ABC1D23')
    expect(maskVehiclePlate('abc-1234')).toBe('ABC1234')
  })

  it('masks currency and extracts cents', () => {
    expect(currencyMaskToCents('R$ 1.234,56')).toBe(123456)
    expect(maskCurrency('123456')).toBe('R$ 1.234,56')
    expect(formatCentsAsCurrency(12000)).toBe('R$ 120,00')
  })
})
