import type { Portfolio, PortfolioReq } from '@/types/api'
import { $http } from './axios'

export const listPortfolio = (): Promise<Portfolio[]> => $http.get('/portfolio') as any
export const addPortfolio = (req: PortfolioReq): Promise<void> => $http.post('/portfolio', req) as any
export const removePortfolio = (tsCode: string): Promise<void> => $http.delete(`/portfolio/${tsCode}`) as any
export const updatePortfolioNote = (tsCode: string, req: PortfolioReq): Promise<void> =>
  $http.patch(`/portfolio/${tsCode}`, req) as any
