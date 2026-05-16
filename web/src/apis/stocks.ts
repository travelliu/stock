import type { Stock, DailyBar, PageResult, RealtimeQuote } from '@/types/api'
import { $http } from './axios'

export const searchStocks = (q: string, limit = 20): Promise<Stock[]> =>
  $http.get('/stocks', { params: { q, limit } }) as any

export const getStock = (code: string): Promise<Stock> => $http.get(`/stocks/${code}`) as any

export const queryBars = (
  code: string,
  params?: { from?: string; to?: string; page?: number; limit?: number },
): Promise<PageResult<DailyBar>> => $http.get(`/bars/${code}`, { params }) as any

export const getQuote = (code: string): Promise<RealtimeQuote> =>
  $http.get(`/quotes/${code}`) as any
