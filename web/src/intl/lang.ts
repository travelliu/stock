export const Langs = {
  zh: 'zh',
  en: 'en',
} as const

export type Lang = typeof Langs[keyof typeof Langs]

export const ElementLangs: Record<Lang, string> = {
  zh: 'zh-CN',
  en: 'en',
}
