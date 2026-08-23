export type AppTheme = 'dark' | 'light'

const STORAGE_KEY = 'empi-theme'

export function useTheme() {
  const theme = useState<AppTheme>('app-theme', () => 'dark')

  function apply(nextTheme: AppTheme) {
    theme.value = nextTheme

    if (import.meta.client) {
      document.documentElement.dataset.theme = nextTheme
      window.localStorage.setItem(STORAGE_KEY, nextTheme)
    }
  }

  function initialise() {
    if (!import.meta.client) return

    const stored = window.localStorage.getItem(STORAGE_KEY)
    const preferred = window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark'
    apply(stored === 'light' || stored === 'dark' ? stored : preferred)
  }

  function toggle() {
    apply(theme.value === 'dark' ? 'light' : 'dark')
  }

  return { theme: readonly(theme), initialise, toggle }
}
