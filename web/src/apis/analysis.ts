import type { AnalysisResult, AnalysisPrediction, PageResult } from '@/types/api'
import { $http } from './axios'

export interface AnalysisParams {
  actualOpen?: number
  actualHigh?: number
  actualLow?: number
  actualClose?: number
}

export const getAnalysis = (code: string, params?: AnalysisParams): Promise<AnalysisResult> => {
  const qs = new URLSearchParams()
  if (params?.actualOpen !== undefined) qs.set('actual_open', String(params.actualOpen))
  if (params?.actualHigh !== undefined) qs.set('actual_high', String(params.actualHigh))
  if (params?.actualLow !== undefined) qs.set('actual_low', String(params.actualLow))
  if (params?.actualClose !== undefined) qs.set('actual_close', String(params.actualClose))
  const q = qs.toString()
  return $http.get(`/analysis/${code}${q ? '?' + q : ''}`) as any
}

export interface PredictionsParams {
  from?: string
  to?: string
  page?: number
  limit?: number
}

export const getPredictions = (
  code: string,
  params?: PredictionsParams,
): Promise<PageResult<AnalysisPrediction>> => {
  const qs = new URLSearchParams()
  if (params?.from) qs.set('from', params.from)
  if (params?.to) qs.set('to', params.to)
  qs.set('page', String(params?.page ?? 1))
  qs.set('limit', String(params?.limit ?? 20))
  return $http.get(`/analysis/predictions/${code}?${qs.toString()}`) as any
}

export const recalcPredictions = (code?: string): Promise<{ updated: number }> => {
  const qs = new URLSearchParams()
  if (code) qs.set('code', code)
  const path = `/analysis/recalc${qs.toString() ? '?' + qs.toString() : ''}`
  return $http.post(path) as any
}
