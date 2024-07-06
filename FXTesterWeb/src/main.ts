/**
 * main.ts
 *
 * Bootstraps Vuetify and other plugins then mounts the App`
 */

// Styles
import 'material-design-icons-iconfont/dist/material-design-icons.css'
import '@mdi/font/css/materialdesignicons.css'
import "vuetify/dist/vuetify.min.css"

// Components
import App from './App.vue'
import "./css/app.css"

// Composables
import { createApp } from 'vue'
import { createVuetify } from 'vuetify'
import router from './router' // !!! ポイント !!!

const app = createApp(App)

const vuetify = createVuetify({
  theme: {
    defaultTheme: 'light'
  },
})
app.use(vuetify)
app.use(router)

app.mount('#app')
