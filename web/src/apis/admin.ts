import type { User, JobRun, CreateUserReq, PatchUserReq } from '@/types/api'
import { $http } from './axios'

export const listUsers = (): Promise<User[]> => $http.get('/admin/users') as any
export const createUser = (req: CreateUserReq): Promise<User> => $http.post('/admin/users', req) as any
export const patchUser = (id: number, req: PatchUserReq): Promise<void> => $http.patch(`/admin/users/${id}`, req) as any
export const deleteUser = (id: number): Promise<void> => $http.delete(`/admin/users/${id}`) as any
export const syncStocklist = (): Promise<void> => $http.post('/admin/stocks/sync') as any
export const syncBars = (): Promise<void> => $http.post('/admin/bars/sync') as any
export const jobStatus = (job: string): Promise<JobRun> => $http.get('/admin/sync/status', { params: { job } }) as any
