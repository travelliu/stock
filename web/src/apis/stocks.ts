import type { Stock, DailyBar, PageResult, RealtimeAndAnalysis, ConceptBlocks, FundFlow } from '@/types/api'
import { $http } from './axios'

export const searchStocks = (q: string, limit = 20): Promise<Stock[]> =>
  $http.get('/stocks', { params: { q, limit } }) as any

export const getStock = (code: string): Promise<Stock> => $http.get(`/stocks/${code}`) as any

export const queryBars = (
  code: string,
  params?: { from?: string; to?: string; page?: number; limit?: number },
): Promise<PageResult<DailyBar>> => $http.get(`/stocks/${code}/bars`, { params }) as any

export const getQuote = (code: string): Promise<RealtimeAndAnalysis> =>
  $http.get(`/stocks/${code}/quote`) as any

export const getConceptBlocks = (code: string): Promise<ConceptBlocks> =>
  $http.get(`/stocks/${code}/concepts`) as any

export const getFundFlow = (code: string): Promise<FundFlow> =>
  $http.get(`/stocks/${code}/fund-flow`) as any
