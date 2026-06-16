export default defineNuxtConfig({
  srcDir: 'app',
  modules: ['@pinia/nuxt'],
  css: ['~/assets/scss/main.scss'],
  runtimeConfig: {
    apiBase: process.env.NUXT_API_BASE || 'http://localhost:8080',
    public: {
      appName: process.env.NUXT_PUBLIC_APP_NAME || 'EMPI ERP',
      apiBaseUrl: process.env.NUXT_PUBLIC_API_BASE_URL || '/api'
    }
  },
  typescript: {
    strict: true,
    typeCheck: false
  },
  routeRules: {
    '/**': {
      headers: {
        'X-Frame-Options': 'DENY',
        'X-Content-Type-Options': 'nosniff',
        'Referrer-Policy': 'strict-origin-when-cross-origin',
        'Permissions-Policy': 'camera=(), microphone=(), geolocation=()'
      }
    }
  },
  compatibilityDate: '2026-06-16'
})
