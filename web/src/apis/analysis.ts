import type { AnalysisResult, AnalysisPrediction } from '@/types/api'
import { $http } from './axios'

export interface AnalysisParams {
  actualOpen?: number
  actualHigh?: number
  actualLow?: number
  actualClose?: number
  withDraft?: boolean
}

export const getAnalysis = (tsCode: string, params?: AnalysisParams): Promise<AnalysisResult> => {
  const qs = new URLSearchParams()
  if (params?.actualOpen !== undefined) qs.set('actual_open', String(params.actualOpen))
  if (params?.actualHigh !== undefined) qs.set('actual_high', String(params.actualHigh))
  if (params?.actualLow !== undefined) qs.set('actual_low', String(params.actualLow))
  if (params?.actualClose !== undefined) qs.set('actual_close', String(params.actualClose))
  qs.set('with_draft', String(params?.withDraft ?? true))
  return $http.get(`/analysis/${tsCode}?${qs.toString()}`) as any
}

export interface PredictionsParams {
  from?: string
  to?: string
  limit?: number
}

export const getPredictions = (tsCode: string, params?: PredictionsParams): Promise<AnalysisPrediction[]> => {
  const qs = new URLSearchParams()
  if (params?.from) qs.set('from', params.from)
  if (params?.to) qs.set('to', params.to)
  qs.set('limit', String(params?.limit ?? 30))
  return $http.get(`/analysis/predictions/${tsCode}?${qs.toString()}`) as any
}

export const recalcPredictions = (tsCode?: string): Promise<{ updated: number }> => {
  const qs = new URLSearchParams()
  if (tsCode) qs.set('ts_code', tsCode)
  const path = `/analysis/recalc${qs.toString() ? '?' + qs.toString() : ''}`
  return $http.post(path) as any
}
