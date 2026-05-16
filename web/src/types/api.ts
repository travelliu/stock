// keep in sync with pkg/models
// @see pkg/models/foo.go::Foo

export interface PageResult<T> {
  items: T[]
  total: number
  page: number
  limit: number
}

export interface User {
  id: number
  username: string
  role: string
  tushareToken: string
  disabled: boolean
  createdAt: string
  updatedAt: string
}

export interface Portfolio {
  id: number
  userId: number
  tsCode: string
  code: string
  name: string
  note: string
  addedAt: string
}

export interface PortfolioReq {
  code: string
  note?: string
}

export interface Stock {
  tsCode: string
  code: string
  name: string
  area: string
  industry: string
  market: string
  exchange: string
  listDate: string
  delisted: boolean
  updatedAt: string
}

export interface Spreads {
  oh: number
  ol: number
  hl: number
  oc: number
  hc: number
  lc: number
}

export interface DailyBar {
  tsCode: string
  tradeDate: string
  open: number
  high: number
  low: number
  close: number
  vol: number
  amount: number
  spreads: Spreads
}

export interface WindowInfo {
  id: string
  name: string
  day: number
}

export interface RecommendRangeResult {
  low: number
  high: number
  cumPct: number
}

export interface MeansAvgData {
  count: number
  avg: number
  median: number
  mean: number
  ewma: number
  stdDev: number
  avgRatio: number
  ewmaRatio: number
  distribution: DistBucket[] | null
  recommend: RecommendRangeResult | null
}

export interface DistBucket {
  lower: number
  upper: number
  count: number
  pct: number
}

export interface MeansData {
  spreadOH: MeansAvgData | null
  spreadOL: MeansAvgData | null
  spreadHL: MeansAvgData | null
  spreadHC: MeansAvgData | null
  spreadLC: MeansAvgData | null
  spreadOC: MeansAvgData | null
}

export interface PredictBreakdown {
  byMean: number
  byMedian: number
  byEwma: number
  byRatio: number
  reverseLow: number
  reverseHigh: number
  mean: number
}

export interface WindowPredict {
  high: PredictBreakdown
  low: PredictBreakdown
  close: PredictBreakdown
}

export interface WindowData {
  info: WindowInfo
  means: MeansData | null
  predict?: WindowPredict
}

export interface PredictRow {
  mean: number
}

export interface RefTable {
  high: PredictRow
  low: PredictRow
  close: PredictRow
}

export interface AnalysisResult {
  tsCode: string
  stockName: string
  windows: WindowData[]
  compositeMeans: Record<string, number>
  refTable?: RefTable
  openPrice?: number
  actualHigh?: number
  actualLow?: number
  actualClose?: number
}

export interface AnalysisPrediction {
  id: number
  tsCode: string
  tradeDate: string
  sampleCounts: Record<string, number> | null
  windowMeans: unknown
  compositeMeans: Record<string, number> | null
  openPrice: number
  predictHigh: number
  predictLow: number
  predictClose: number
  actualHigh: number
  actualLow: number
  actualClose: number
  createdAt: string
  updatedAt: string
}


export interface APIToken {
  id: number
  userId: number
  name: string
  tokenHash: string
  lastUsedAt: string | null
  expiresAt: string | null
  createdAt: string
}

export interface JobRun {
  id: number
  jobName: string
  startedAt: string
  finishedAt: string | null
  status: string
  message: string
}

export interface RealtimeQuote {
  tsCode: string
  name: string
  price: number
  prevClose: number
  open: number
  vol: number
  high: number
  low: number
  amount: number
  change: number
  changePct: number
  limitUp: number
  limitDown: number
  quoteTime: string
  updatedAt: string
}

export interface LoginReq {
  username: string
  password: string
}

export interface ChangePasswordReq {
  old: string
  new: string
}

export interface SetTushareTokenReq {
  token: string
}

export interface IssueTokenReq {
  name: string
  expiresAt?: string
}

export interface IssueTokenResp {
  token: string
  metadata: APIToken
}

export interface CreateUserReq {
  username: string
  password: string
  role: string
  tushareToken?: string
}

export interface PatchUserReq {
  role?: string
  disabled?: boolean
  tushareToken?: string
}
