import { expect, test, type Page } from '@playwright/test'

test.setTimeout(60000)
test.beforeEach(async ({ baseURL }) => {
  expect(['localhost', '127.0.0.1', '[::1]']).toContain(new URL(baseURL!).hostname)
})

async function mounted(page: Page, selector: string) {
  await page.waitForFunction((selector) => {
    const el = document.querySelector(selector) as HTMLElement & { __vueParentComponent?: { isMounted: boolean } } | null
    return el?.__vueParentComponent?.isMounted === true
  }, selector)
}

test('filtra produtos em todo o catálogo, preserva paginação e permite limpar', async ({ page }, info) => {
  await page.goto('/login')
  await mounted(page, '.login-form')
  await page.getByLabel('CPF').fill(process.env.E2E_ADMIN_CPF || '52998224725')
  await page.getByLabel('Senha').fill(process.env.E2E_ADMIN_PASSWORD || 'change-me-admin-password')
  await page.getByRole('button', { name: 'Entrar', exact: true }).click()
  await expect(page).toHaveURL(/\/$/)
  const cookie = (await page.context().cookies()).find(({ name }) => name === 'auth._token.laravelJWT')!
  const tag = `FiltroProduto-${info.project.name}-${Date.now()}`
  const day = new Intl.DateTimeFormat('en-CA', { timeZone: 'America/Sao_Paulo', year: 'numeric', month: '2-digit', day: '2-digit' }).format(new Date())
  for (let index = 0; index < 12; index++) {
    const response = await page.request.post('/api/stock', {
      headers: { Authorization: decodeURIComponent(cookie.value) },
      data: { name: `${tag}-${index}`, description: '', costCents: 1000, markupPercent: 20, quantity: 0,
        entry: { supplierName: tag, quantity: 1, unitCostCents: 1000,
          installments: [{ amountCents: 1000, dueDate: day, plannedMethod: 'pix' }] } }
    })
    expect(response.ok(), await response.text()).toBe(true)
  }
  await page.goto('/stock/products')
  await mounted(page, '.stock-filters')
  const search = page.getByRole('searchbox', { name: 'Buscar produto' })
  await search.fill(tag)
  await page.getByRole('button', { name: 'Filtrar', exact: true }).click()
  await expect(page.getByText('12 produtos encontrados', { exact: true })).toBeVisible()
  await expect(page.getByText('Página 1 de 2', { exact: true })).toBeVisible()
  await page.getByRole('button', { name: 'Próxima', exact: true }).click()
  await expect(page.getByText('Página 2 de 2', { exact: true })).toBeVisible()
  await expect(search).toHaveValue(tag)
  await expect(page.locator('.stock-table tbody tr')).toHaveCount(2)
  // The oldest result is outside the unfiltered first page and matches exactly.
  await search.fill(`${tag}-0`)
  await search.press('Enter')
  await expect(page.getByText('1 produto encontrado', { exact: true })).toBeVisible()
  await expect(page.getByText('Página 1 de 1', { exact: true })).toBeVisible()
  await expect(page.locator('.stock-table tbody tr')).toContainText(`${tag}-0`)
  await page.screenshot({ path: `/tmp/empi-product-filter-${info.project.name}.png`, fullPage: true })
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
  await search.fill(`sem-resultado-${tag}`)
  await search.press('Enter')
  await expect(page.getByText('Nenhum produto encontrado', { exact: true })).toBeVisible()
  await page.getByRole('button', { name: 'Limpar filtro', exact: true }).click()
  await expect(search).toHaveValue('')
  await expect(page.getByText('Nenhum produto encontrado', { exact: true })).toBeHidden()
})
