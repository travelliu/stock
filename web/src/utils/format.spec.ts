import { describe, it, expect } from 'vitest'
import { fmtPrice, fmtPct, priceClass } from './format'

describe('format', () => {
  it('fmtPrice', () => {
    expect(fmtPrice(10.5)).toBe('10.50')
    expect(fmtPrice(0)).toBe('0.00')
  })

  it('fmtPct', () => {
    expect(fmtPct(11, 10)).toBe('+10.00%')
    expect(fmtPct(9, 10)).toBe('-10.00%')
    expect(fmtPct(10, 0)).toBe('0.00%')
  })

  it('priceClass', () => {
    expect(priceClass(11, 10)).toBe('g-up')
    expect(priceClass(9, 10)).toBe('g-down')
    expect(priceClass(10, 10)).toBe('g-flat')
  })
})
