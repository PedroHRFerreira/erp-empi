const withProtocol = (value?: string) => {
  if (!value) {
    return undefined
  }

  return /^https?:\/\//.test(value) ? value : `https://${value}`
}

const withApiPath = (value?: string) => {
  if (!value) {
    return undefined
  }

  return value.endsWith('/api') ? value : `${value.replace(/\/$/, '')}/api`
}

const apiBase = withProtocol(process.env.NUXT_API_BASE || process.env.NUXT_API_HOST) || 'http://localhost:8080'
const publicApiBaseUrl =
  process.env.NUXT_PUBLIC_API_BASE_URL ||
  withApiPath(withProtocol(process.env.NUXT_PUBLIC_API_HOST || process.env.NUXT_API_HOST)) ||
  '/api'

export default defineNuxtConfig({
  srcDir: 'app',
  modules: ['@pinia/nuxt'],
  css: ['~/assets/scss/main.scss'],
  runtimeConfig: {
    apiBase,
    public: {
      appName: process.env.NUXT_PUBLIC_APP_NAME || 'EMPI ERP',
      apiBaseUrl: publicApiBaseUrl
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
