import type { Stock, DailyBar } from '@/types/api'
import { $http } from './axios'

export const searchStocks = (q: string, limit = 20): Promise<Stock[]> =>
  $http.get('/stocks', { params: { q, limit } }) as any
export const getStock = (tsCode: string): Promise<Stock> => $http.get(`/stocks/${tsCode}`) as any
export const queryBars = (tsCode: string, from?: string, to?: string): Promise<DailyBar[]> =>
  $http.get(`/bars/${tsCode}`, { params: { from, to } }) as any
