import { describe, it, expect, vi, beforeEach } from 'vitest'
import MockAdapter from 'axios-mock-adapter'
import axios from 'axios'

vi.mock('@/router', () => ({
  default: { push: vi.fn(), beforeEach: vi.fn() },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: vi.fn(() => ({ logout: vi.fn() })),
}))

vi.mock('@/stores/lang', () => ({
  useLangStore: vi.fn(() => ({ lang: 'zh' })),
}))

describe('auth api', () => {
  let mock: MockAdapter

  beforeEach(() => {
    mock = new MockAdapter(axios)
  })

  it('login returns user with camelCase fields', async () => {
    mock.onPost('/api/auth/login').reply(200, {
      code: 200,
      message: 'ok',
      data: { id: 1, username: 'alice', role: 'user', tushareToken: 'tk', createdAt: '2025-01-01', updatedAt: '2025-01-01' },
    })
    const { login } = await import('./auth')
    const user = await login({ username: 'alice', password: 'secret' })
    expect(user.username).toBe('alice')
    expect(user.tushareToken).toBe('tk')
  })
})
