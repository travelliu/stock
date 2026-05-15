import type { User, LoginReq } from '@/types/api'
import { $http } from './axios'

export const login = (req: LoginReq): Promise<User> => $http.post('/auth/login', req) as any
export const logout = (): Promise<void> => $http.post('/auth/logout') as any
export const me = (): Promise<User> => $http.get('/auth/me') as any
