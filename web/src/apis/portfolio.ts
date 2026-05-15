import type { Portfolio, PortfolioReq } from '@/types/api'
import { $http } from './axios'

export const listPortfolio = (): Promise<Portfolio[]> => $http.get('/portfolio') as any
export const addPortfolio = (req: PortfolioReq): Promise<void> => $http.post('/portfolio', req) as any
export const removePortfolio = (code: string): Promise<void> => $http.delete(`/portfolio/${code}`) as any
export const updatePortfolioNote = (code: string, req: PortfolioReq): Promise<void> =>
  $http.patch(`/portfolio/${code}`, req) as any
