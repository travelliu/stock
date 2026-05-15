import type { ChangePasswordReq, SetTushareTokenReq, IssueTokenReq, IssueTokenResp, APIToken } from '@/types/api'
import { $http } from './axios'

export const changePassword = (req: ChangePasswordReq): Promise<void> => $http.post('/me/password', req) as any
export const setTushareToken = (req: SetTushareTokenReq): Promise<void> => $http.patch('/me/tushare_token', req) as any
export const listTokens = (): Promise<APIToken[]> => $http.get('/me/tokens') as any
export const issueToken = (req: IssueTokenReq): Promise<IssueTokenResp> => $http.post('/me/tokens', req) as any
export const revokeToken = (id: number): Promise<void> => $http.delete(`/me/tokens/${id}`) as any
