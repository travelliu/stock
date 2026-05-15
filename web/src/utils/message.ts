import { ElMessage } from 'element-plus'

const recent = new Map<string, number>()

export function wMessage(type: 'success' | 'error' | 'warning' | 'info', message: string) {
  const key = `${type}:${message}`
  const last = recent.get(key)
  const now = Date.now()
  if (last && now - last < 3000) {
    return
  }
  recent.set(key, now)
  ElMessage[type](message)
}
