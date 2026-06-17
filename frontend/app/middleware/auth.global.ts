export default defineNuxtRouteMiddleware(async (to) => {
  if (to.path === '/login') {
    setPageLayout('auth')
  }

  const auth = useAuthStore()
  await auth.bootstrap()

  if (to.path === '/login' && Boolean(auth.user)) {
    return navigateTo('/')
  }

  if (to.path !== '/login' && !auth.user) {
    return navigateTo('/login')
  }
})
