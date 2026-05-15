import { createI18n } from 'vue-i18n'
import zh from './langs/zh/index'
import en from './langs/en/index'
import { storage } from '@/utils/storage'

const saved = storage.get('lang') || 'zh'

export const i18n = createI18n({
  legacy: false,
  locale: saved,
  fallbackLocale: 'zh',
  messages: { zh, en },
})
