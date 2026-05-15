import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import '@/assets/css/index.scss'
import '@/assets/css/element-reset.scss'

import App from './App.vue'
import router from './router'
import { i18n } from './intl'
import GIcon from './components/GIcon.vue'
import GEllipsis from './components/GEllipsis.vue'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(ElementPlus)
app.use(i18n)
app.component('GIcon', GIcon)
app.component('GEllipsis', GEllipsis)
app.mount('#app')
