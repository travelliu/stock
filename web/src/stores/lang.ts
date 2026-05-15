import { defineStore } from 'pinia'
import { ref } from 'vue'
import { storage } from '@/utils/storage'
import type { Lang } from '@/intl/lang'

export const useLangStore = defineStore('lang', () => {
  const lang = ref<Lang>((storage.get('lang') as Lang) || 'zh')

  function setLang(v: Lang) {
    lang.value = v
    storage.set('lang', v)
  }

  return { lang, setLang }
})
