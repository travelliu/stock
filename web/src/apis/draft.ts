import type { IntradayDraft } from '@/types/api'
import { $http } from './axios'

export const getDraftToday = (tsCode: string, tradeDate?: string): Promise<IntradayDraft> =>
  $http.get('/drafts/today', { params: { ts_code: tsCode, trade_date: tradeDate } }) as any
export const upsertDraft = (body: Record<string, unknown>): Promise<IntradayDraft> =>
  $http.put('/drafts', body) as any
export const deleteDraft = (id: number): Promise<void> => $http.delete(`/drafts/${id}`) as any
