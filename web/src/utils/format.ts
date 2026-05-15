export function fmtPrice(n: number): string {
  return n.toFixed(2)
}

export function fmtPct(cur: number, prev: number): string {
  if (prev === 0) return '0.00%'
  const v = ((cur - prev) / prev) * 100
  return `${v >= 0 ? '+' : ''}${v.toFixed(2)}%`
}

export function priceClass(cur: number, prev: number): string {
  if (cur > prev) return 'g-up'
  if (cur < prev) return 'g-down'
  return 'g-flat'
}
