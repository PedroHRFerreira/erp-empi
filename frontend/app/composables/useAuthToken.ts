import type { CookieRef } from 'nuxt/app'
import type { IToken } from '~/types/auth'

type AuthCookieValue = string | null

const authCookieName = 'auth._token.laravelJWT'

export const useAuthToken = () => {
  const accessToken = useState<string | null>('auth-access-token', () => null)
  const tokenType = useState<string | null>('auth-token-type', () => null)

  function setTokenCookie(token: IToken): void {
    const cookie: CookieRef<AuthCookieValue> = useCookie<AuthCookieValue>(authCookieName, {
      path: '/',
      default: (): AuthCookieValue => null,
      expires: new Date(Date.now() + token.expires_in * 1000),
      maxAge: token.expires_in,
      sameSite: 'lax',
      watch: 'shallow'
    })

    accessToken.value = token.access_token
    tokenType.value = token.token_type
    cookie.value = `${token.token_type} ${token.access_token}`
  }

  function removeTokenCookie(): void {
    const cookie: CookieRef<AuthCookieValue> = useCookie<AuthCookieValue>(authCookieName, {
      path: '/',
      default: (): AuthCookieValue => null,
      expires: new Date(0),
      maxAge: 0,
      sameSite: 'lax',
      watch: 'shallow'
    })

    accessToken.value = null
    tokenType.value = null
    cookie.value = null
  }

  function getAuthorizationHeader(): string | null {
    const cookie: CookieRef<AuthCookieValue> = useCookie<AuthCookieValue>(authCookieName, {
      path: '/',
      default: (): AuthCookieValue => null,
      watch: 'shallow'
    })

    return cookie.value
  }

  return {
    accessToken,
    setTokenCookie,
    removeTokenCookie,
    getAuthorizationHeader
  }
}
