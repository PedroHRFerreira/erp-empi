export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuthStore()
  await auth.bootstrap()

  if (to.path === '/login' && Boolean(auth.user)) {
    return navigateTo('/')
  }

  if (to.path !== '/login' && !auth.user) {
    return navigateTo('/login')
  }
})
