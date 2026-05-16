// fmtQuoteTime normalises Tencent's quoteTime field to "HH:MM".
// Handles: "20260515161500" (14-digit compact), "HH:MM:SS", "YYYY-MM-DD HH:MM:SS".
export function fmtQuoteTime(t: string | undefined): string {
  if (!t) return ''
  if (/^\d{14}$/.test(t)) return t.slice(8, 10) + ':' + t.slice(10, 12)
  if (t.length >= 16 && t.includes(' ')) return t.slice(11, 16)
  return t.slice(0, 5)
}

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
