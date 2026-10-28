import AxeBuilder from '@axe-core/playwright'
import { expect, test, type Page } from '@playwright/test'
import type { IReceipt, IPaginated } from '../server/contracts/types'

test.setTimeout(60000)

test.beforeEach(async ({ baseURL }) => {
  expect(['127.0.0.1', 'localhost', '[::1]']).toContain(new URL(baseURL!).hostname)
})

async function login(page: Page) {
  await page.goto('/login')
  await page.waitForFunction(() => {
    const form = document.querySelector('.login-form') as HTMLElement & { __vueParentComponent?: { isMounted: boolean } } | null
    return form?.__vueParentComponent?.isMounted === true
  })
  await page.getByLabel('CPF').fill(process.env.E2E_ADMIN_CPF || '52998224725')
  await page.getByLabel('Senha').fill(process.env.E2E_ADMIN_PASSWORD || 'change-me-admin-password')
  await page.getByRole('button', { name: 'Entrar', exact: true }).click()
  await expect(page).toHaveURL(/\/$/)
}

async function api<T>(page: Page, path: string, method = 'GET', data?: unknown): Promise<T> {
  const cookie = (await page.context().cookies()).find(({ name }) => name === 'auth._token.laravelJWT')
  expect(cookie).toBeDefined()
  const response = await page.request.fetch(`/api${path}`, {
    method, data, headers: { Authorization: decodeURIComponent(cookie!.value) }
  })
  expect(response.ok(), `${method} ${path}: ${response.status()} ${await response.text()}`).toBe(true)
  return response.json()
}

async function createReceipt(page: Page, services: string, options: { normal?: boolean; mixed?: boolean } = {}) {
  return api<IReceipt>(page, '/receipts', 'POST', {
    quick: !options.normal,
    client: options.normal ? { name: `Cliente ${services}`, phone: '11987654321' } : { name: '', phone: '' },
    vehicleModel: options.normal ? 'Fiat Uno' : '', vehicleYear: options.normal ? 2020 : 0,
    vehiclePlate: options.normal ? 'ABC1D23' : '', services, laborPriceCents: 10000,
    discountCents: 0, priceCents: 10000, cardFeePercent: 0, installments: 1,
    paymentMethod: options.mixed ? 'mixed' : 'pix',
    payments: options.mixed ? [
      { paymentMethod: 'pix', baseAmountCents: 6000, cardFeePercent: 0, installments: 1 },
      { paymentMethod: 'credit_card', baseAmountCents: 4000, cardFeePercent: 0, installments: 1 }
    ] : [{ paymentMethod: 'pix', baseAmountCents: 10000, cardFeePercent: 0, installments: 1 }],
    notes: '', items: [], serviceExpenses: []
  })
}

async function apply(page: Page) {
  const query = (await page.getByLabel(queryLabel).inputValue()).trim()
  const response = page.waitForResponse((response) => response.request().method() === 'GET'
    && new URL(response.url()).pathname === '/api/receipts'
    && new URL(response.url()).searchParams.get('q') === query)
  await page.getByRole('button', { name: 'Aplicar filtros', exact: true }).click()
  const result = await response
  expect(result.ok()).toBe(true)
  const body = await result.json() as IPaginated<IReceipt>
  await expect(page.getByRole('status', { name: 'Carregando recibos...' })).toBeHidden()
  await expect(page.locator('.receipts-template__total')).toContainText(`${body.total} recibo`)
  return body
}

const queryLabel = 'Cliente, telefone, placa ou serviço'

async function openReceipts(page: Page) {
  await page.goto('/receipts')
  await waitForReceiptsReady(page)
}

async function waitForReceiptsReady(page: Page) {
  await page.waitForURL(/\/receipts$/)
  await page.waitForFunction(() => {
    const root = document.querySelector('.receipts-template__filters')?.closest('.page') as HTMLElement & { __vueParentComponent?: { isMounted: boolean } } | null
    return root?.__vueParentComponent?.isMounted === true
  })
  await expect(page.getByRole('button', { name: 'Aplicar filtros', exact: true })).toBeEnabled()
}

async function navigateWithSidebar(page: Page, name: string) {
  const link = page.getByRole('link', { name, exact: true })
  if (!await link.isVisible()) await page.getByRole('button', { name: 'Abrir menu principal', exact: true }).click()
  await link.click()
}

test('combina busca, datas, situação, pagamento e tipo com dados reais', async ({ page }, testInfo) => {
  await login(page)
  const tag = `filtros-${testInfo.project.name}-${Date.now()}`
  const normal = await createReceipt(page, `${tag}-normal`, { normal: true })
  const mixed = await createReceipt(page, `${tag}-misto`, { mixed: true })
  const cancelled = await createReceipt(page, `${tag}-cancelado`)
  await api(page, `/receipts/${cancelled.id}/cancel`, 'POST')
  await openReceipts(page)
  await page.getByLabel(queryLabel).fill(tag)
  expect((await apply(page)).total).toBe(3)

  // The same search covers service, client, normalized phone, and vehicle plate.
  for (const query of [`Cliente ${tag}-normal`, '11987654321', 'ABC1D23']) {
    await page.getByLabel(queryLabel).fill(query)
    expect((await apply(page)).data.some(({ id }) => id === normal.id)).toBe(true)
  }
  await page.getByLabel(queryLabel).fill(tag)
  await page.getByLabel('Forma de pagamento', { exact: true }).selectOption('credit_card')
  expect((await apply(page)).data.map(({ id }) => id)).toEqual([mixed.id])
  await page.getByLabel('Forma de pagamento', { exact: true }).selectOption('mixed')
  expect((await apply(page)).data.map(({ id }) => id)).toEqual([mixed.id])
  await page.getByLabel('Forma de pagamento', { exact: true }).selectOption('pix')
  expect((await apply(page)).total).toBe(3)

  await page.getByLabel('Situação', { exact: true }).selectOption('cancelled')
  expect((await apply(page)).data.map(({ id }) => id)).toEqual([cancelled.id])
  await page.getByLabel('Situação', { exact: true }).selectOption('pending')
  await page.getByLabel('Tipo de recibo', { exact: true }).selectOption('normal')
  expect((await apply(page)).data.map(({ id }) => id)).toEqual([normal.id])
  await page.getByLabel('Tipo de recibo', { exact: true }).selectOption('quick')
  const today = new Intl.DateTimeFormat('en-CA', { timeZone: 'America/Sao_Paulo', year: 'numeric', month: '2-digit', day: '2-digit' }).format(new Date())
  await page.getByLabel('Data inicial de criação').fill(today)
  await page.getByLabel('Data final de criação').fill(today)
  expect((await apply(page)).data.map(({ id }) => id)).toEqual([mixed.id])

  const requests: string[] = []
  page.on('request', (request) => { if (new URL(request.url()).pathname === '/api/receipts') requests.push(request.url()) })
  await page.getByLabel('Data inicial de criação').fill('2099-01-02')
  await page.getByLabel('Data final de criação').fill('2099-01-01')
  await page.getByRole('button', { name: 'Aplicar filtros', exact: true }).click()
  await expect(page.getByRole('alert')).toContainText('A data final deve ser igual ou posterior à data inicial.')
  expect(requests).toHaveLength(0)
  await page.getByLabel('Data inicial de criação').fill('2099-01-01')
  expect((await apply(page)).total).toBe(0)
  await expect(page.getByText('Nenhum recibo encontrado', { exact: true })).toBeVisible()

  const result = await new AxeBuilder({ page }).include('form[aria-label="Filtros de recibos"]').withTags(['wcag2a', 'wcag2aa', 'wcag21aa']).analyze()
  expect(result.violations.filter(({ impact }) => impact === 'critical' || impact === 'serious')).toEqual([])
  expect(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth + 1)).toBe(false)
  await page.screenshot({ path: `/tmp/receipt-filters-${testInfo.project.name}.png`, fullPage: true })
  await page.getByRole('button', { name: 'Limpar filtros', exact: true }).click()
  await expect(page.getByLabel(queryLabel)).toHaveValue('')
  await expect(page.getByLabel('Situação', { exact: true })).toHaveValue('')
  await expect(page.getByLabel('Data inicial de criação')).toHaveValue('')
})

test('preserva filtros na paginação e quitação e isola a recuperação', async ({ page }, testInfo) => {
  await login(page)
  const tag = `paginacao-${testInfo.project.name}-${Date.now()}`
  for (let index = 0; index < 11; index++) await createReceipt(page, `${tag}-${index}`)
  await openReceipts(page)
  await page.getByLabel(queryLabel).fill(tag)
  await page.getByLabel('Situação', { exact: true }).selectOption('pending')
  expect((await apply(page)).total).toBe(11)
  await expect(page.locator('.receipts-list__row')).toHaveCount(10)
  await page.getByRole('button', { name: 'Próxima', exact: true }).click()
  await expect(page.getByText('Página 2 de 2', { exact: true })).toBeVisible()
  await expect(page.locator('.receipts-list__row')).toHaveCount(1)
  await page.getByRole('button', { name: 'Marcar como pago', exact: true }).click()
  await expect(page.locator('.receipts-template__total')).toContainText('10 recibos encontrados')
  await expect(page.getByLabel(queryLabel)).toHaveValue(tag)
  await expect(page.getByLabel('Situação', { exact: true })).toHaveValue('pending')
  await expect(page.getByText('Página 1 de 1', { exact: true })).toBeVisible()
  await expect(page.locator('.receipts-list__row')).toHaveCount(10)

  // Client-side navigation keeps the Pinia instances alive while visiting recovery.
  await navigateWithSidebar(page, 'Recuperação')
  await expect(page.getByRole('heading', { name: 'Recuperação', exact: true })).toBeVisible()
  await navigateWithSidebar(page, 'Recibos')
  await waitForReceiptsReady(page)
  await expect(page.getByLabel(queryLabel)).toHaveValue(tag)
  await expect(page.getByLabel('Situação', { exact: true })).toHaveValue('pending')
  await expect(page.locator('.receipts-template__total')).toContainText('10 recibos encontrados')
  // Move the pointer into the content before selecting: the tablet sidebar expands on hover.
  await page.getByLabel('Situação', { exact: true }).hover()
  await page.getByLabel('Situação', { exact: true }).focus()
  await page.getByLabel('Situação', { exact: true }).selectOption('paid')
  expect((await apply(page)).total).toBe(1)
  await expect(page.getByText('Página 1 de 1', { exact: true })).toBeVisible()
})
