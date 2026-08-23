import { expect, test, type Page } from '@playwright/test'

const adminCpf = process.env.E2E_ADMIN_CPF || '52998224725'
const adminPassword = process.env.E2E_ADMIN_PASSWORD || 'change-me-admin-password'
const authBaseUrl = process.env.E2E_AUTH_BASE_URL

async function login(page: Page) {
  await page.goto(authBaseUrl ? `${authBaseUrl}/login` : '/login')
  await page.getByLabel('CPF').fill(adminCpf)
  await page.getByLabel('Senha').fill(adminPassword)
  await page.getByRole('button', { name: 'Entrar' }).click()
  await expect(page).toHaveURL(/\/$/)
}

async function waitForHydration(page: Page) {
  await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as HTMLElement & { __vue_app__?: unknown } | null)?.__vue_app__))
}

test('separa lançamentos operacionais das saídas realizadas', async ({ page }) => {
  await login(page)
  await page.goto('/expenses')
  await waitForHydration(page)

  await expect(page.getByRole('heading', { name: 'Gastos operacionais cadastrados' })).toBeVisible()
  await expect(page.getByText('Inclui parcelas pendentes.')).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Histórico de gastos pagos' })).toBeVisible()
})

test('mantém gasto pendente acessível e abre seu histórico', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop', 'Jornada mutável executada uma vez na base compartilhada')
  await login(page)
  await page.goto('/expenses')
  await waitForHydration(page)
  await expect(page.getByRole('heading', { name: 'Gastos operacionais cadastrados' })).toBeVisible()
  await page.getByRole('button', { name: 'Adicionar' }).click()
  await expect(page.getByRole('heading', { name: 'Adicionar gasto' })).toBeVisible()

  const description = `Gasto parcelado E2E ${Date.now()}`
  await page.getByLabel('Descrição').fill(description)
  await page.getByLabel('Categoria').selectOption('outros')
  await page.getByPlaceholder('R$ 200,00').fill('R$ 42,00')
  await page.getByRole('button', { name: 'Salvar gasto' }).click()

  const row = page.locator('.expenses-registry__row').filter({ hasText: description })
  await expect(row).toContainText('Pendente')
  await row.getByRole('link', { name: 'Ver histórico' }).click()
  await expect(page).toHaveURL(/\/expenses\/[^/]+\/history$/)
  await expect(page.getByRole('heading', { name: new RegExp(description) })).toBeVisible()
  await expect(page.getByText('Saldo pendente')).toBeVisible()
})
