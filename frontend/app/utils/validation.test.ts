import { describe, expect, it } from 'vitest'
import { isCpf, isPlate, onlyDigits } from './validation'

describe('validation helpers', () => {
  it('normalizes digits', () => {
    expect(onlyDigits('529.982.247-25')).toBe('52998224725')
  })

  it('validates cpf', () => {
    expect(isCpf('529.982.247-25')).toBe(true)
    expect(isCpf('111.111.111-11')).toBe(false)
  })

  it('validates old and mercosul plates', () => {
    expect(isPlate('ABC-1234')).toBe(true)
    expect(isPlate('ABC1D23')).toBe(true)
  })
})
