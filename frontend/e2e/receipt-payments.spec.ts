import AxeBuilder from '@axe-core/playwright'
import { expect, test, type Page } from '@playwright/test'
import type { ICashBalances, ICashEntry, ICashSession, IMetricsSummary, IReceipt } from '../server/contracts/types'

const adminCpf = process.env.E2E_ADMIN_CPF || '52998224725'
const adminPassword = process.env.E2E_ADMIN_PASSWORD || 'change-me-admin-password'

async function goto(page: Page, path: string) {
  await page.goto(path)
  await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as HTMLElement & { __vue_app__?: unknown } | null)?.__vue_app__))
}

async function login(page: Page) {
  await goto(page, '/login')
  await page.getByLabel('CPF').fill(adminCpf)
  await page.getByLabel('Senha').fill(adminPassword)
  await page.getByRole('button', { name: 'Entrar', exact: true }).click()
  await expect(page).toHaveURL(/\/$/)
}

// Read the real authenticated proxy, without mocking receipts, metrics or cash.
async function api<T>(page: Page, path: string, method = 'GET'): Promise<T> {
  const cookie = (await page.context().cookies()).find(({ name }) => name === 'auth._token.laravelJWT')
  expect(cookie, 'A sessão deve fornecer autorização às consultas do caixa').toBeDefined()
  const response = await page.request.fetch(`/api${path}`, {
    method,
    headers: { Authorization: decodeURIComponent(cookie!.value) }
  })
  expect(response.ok(), `${method} ${path}: ${response.status()}`).toBe(true)
  const text = await response.text()
  return (text ? JSON.parse(text) : null) as T
}

async function startQuickReceipt(page: Page, description: string) {
  await goto(page, '/receipts')
  await page.getByRole('button', { name: 'Recibo rápido', exact: true }).click()
  await expect(page).toHaveURL(/\/receipts\/new\?quick=1$/)
  await page.getByLabel('Valor da mão de obra').fill('R$ 120,00')
  await page.getByLabel('Desconto', { exact: true }).fill('R$ 20,00')
  await page.getByLabel('Descrição do serviço').fill(description)
  await advanceToPayments(page)
}

async function advanceToPayments(page: Page) {
  for (let step = 0; step < 6; step++) {
    if (await page.getByRole('heading', { name: 'Finalizar recibo', exact: true }).isVisible()) return
    await page.getByRole('button', { name: 'Avançar', exact: true }).click()
  }
  await expect(page.getByRole('heading', { name: 'Finalizar recibo', exact: true })).toBeVisible()
}

async function save(page: Page, editing = false): Promise<IReceipt> {
  const method = editing ? 'PUT' : 'POST'
  const responsePromise = page.waitForResponse((response) =>
    response.request().method() === method && /\/api\/receipts(?:\/[^/]+)?$/.test(new URL(response.url()).pathname))
  await page.getByRole('button', { name: editing ? 'Salvar alterações' : 'Salvar recibo', exact: true }).click()
  const response = await responsePromise
  expect(response.ok()).toBe(true)
  const receipt = await response.json() as IReceipt
  await expect(page).toHaveURL(/\/receipts$/)
  return receipt
}

async function markPaid(page: Page, id: string) {
  await goto(page, `/receipts/${id}/edit`)
  await expect(page.getByRole('heading', { name: 'Editar recibo', exact: true })).toBeVisible()
  await goto(page, '/receipts')
  const receipts = await api<{ data: IReceipt[] }>(page, '/receipts?limit=100&offset=0')
  const index = receipts.data.findIndex((receipt) => receipt.id === id)
  expect(index).toBeGreaterThanOrEqual(0)
  const target = page.locator('.receipts-list__row').nth(index)
  await expect(target).toContainText('Pendente')
  const responsePromise = page.waitForResponse((response) => response.request().method() === 'POST'
    && new URL(response.url()).pathname === `/api/receipts/${id}/pay`)
  await target.getByRole('button', { name: 'Marcar como pago', exact: true }).click()
  expect((await responsePromise).ok()).toBe(true)
  await expect(target).toContainText('Pago')
}

async function receiptEntries(page: Page, id: string) {
  return (await api<ICashEntry[]>(page, '/cash/daily-entries'))
    .filter((entry) => entry.kind === 'receipt_payment' && entry.referenceId === id)
}

async function expectNoOverflow(page: Page) {
  const overflow = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 1)
  expect(overflow, 'A divisão de pagamentos deve caber na largura disponível').toBe(false)
}

test('salva, edita e quita PIX + crédito parcelado com taxa somente no cartão', async ({ page }, testInfo) => {
  await login(page)
  const description = `Divisão crédito ${testInfo.project.name} ${Date.now()}`
  const metricsBefore = await api<IMetricsSummary>(page, '/metrics/summary')
  await startQuickReceipt(page, description)
  await page.getByRole('checkbox', { name: 'Dividir em duas formas' }).check()
  await page.getByRole('combobox', { name: 'Forma de pagamento 2', exact: true }).selectOption('credit_card')
  await page.getByRole('combobox', { name: 'Forma de pagamento 1', exact: true }).selectOption('pix')
  await page.getByRole('textbox', { name: /^Valor base \(sem taxas\)/ }).fill('R$ 60,00')
  await page.getByRole('combobox', { name: 'Parcelas', exact: true }).selectOption('4')
  await page.getByLabel('Juros de parcelamento (%)', { exact: true }).fill('5')
  await expect(page.getByRole('textbox', { name: /^Valor restante \(sem taxas\)/ })).toHaveValue(/40,00/)
  await expect(page.getByRole('region', { name: 'Valores do pagamento 2', exact: true })).toContainText('4x de R$ 10,50')
  await expectNoOverflow(page)
  await page.getByRole('region', { name: 'Pagamento do recibo', exact: true }).screenshot({ path: `/tmp/receipt-split-credit-${testInfo.project.name}.png` })

  const receipt = await save(page)
  expect(receipt.status).toBe('pending')
  expect(receipt.paymentMethod).toBe('mixed')
  expect(receipt.priceCents).toBe(10200)
  expect(receipt.cardFeeCents).toBe(200)
  expect(receipt.payments).toEqual([
    expect.objectContaining({ paymentMethod: 'pix', baseAmountCents: 6000, amountCents: 6000, installments: 1 }),
    expect.objectContaining({ paymentMethod: 'credit_card', baseAmountCents: 4000, cardFeePercent: 5, cardFeeCents: 200, amountCents: 4200, installments: 4 })
  ])
  expect(await receiptEntries(page, receipt.id)).toHaveLength(0)
  await goto(page, `/receipts/${receipt.id}/edit`)
  await advanceToPayments(page)
  await expect(page.getByRole('checkbox', { name: 'Dividir em duas formas' })).toBeChecked()
  await expect(page.getByRole('textbox', { name: /^Valor base \(sem taxas\)/ })).toHaveValue(/60,00/)
  await expect(page.getByRole('combobox', { name: 'Parcelas', exact: true })).toHaveValue('4')
  await expect(page.getByLabel('Juros de parcelamento (%)', { exact: true })).toHaveValue('5')
  await page.getByLabel('Descrição do recibo').fill(`Editado ${description}`)
  const edited = await save(page, true)
  expect(edited.payments).toEqual([
    expect.objectContaining({ paymentMethod: 'pix', baseAmountCents: 6000, amountCents: 6000, installments: 1 }),
    expect.objectContaining({ paymentMethod: 'credit_card', baseAmountCents: 4000, cardFeePercent: 5, cardFeeCents: 200, amountCents: 4200, installments: 4 })
  ])

  const balancesBefore = await api<ICashBalances>(page, '/cash/balances')
  await markPaid(page, receipt.id)
  const entries = await receiptEntries(page, receipt.id)
  expect(entries).toHaveLength(2)
  expect(entries).toEqual(expect.arrayContaining([
    expect.objectContaining({ paymentMethod: 'pix', amountCents: 6000, direction: 'in' }),
    expect.objectContaining({ paymentMethod: 'credit_card', amountCents: 4200, direction: 'in' })
  ]))
  const balancesAfter = await api<ICashBalances>(page, '/cash/balances')
  expect(balancesAfter.pixCents - balancesBefore.pixCents).toBe(6000)
  expect(balancesAfter.creditCardCents - balancesBefore.creditCardCents).toBe(4200)
  const metricsAfter = await api<IMetricsSummary>(page, '/metrics/summary')
  expect(metricsAfter.receiptsTotal - metricsBefore.receiptsTotal).toBe(1)
  expect(metricsAfter.receiptsPaid - metricsBefore.receiptsPaid).toBe(1)
  expect(metricsAfter.revenuePaidCents - metricsBefore.revenuePaidCents).toBe(10200)
  await api(page, `/receipts/${receipt.id}/pay`, 'POST')
  expect(await receiptEntries(page, receipt.id)).toHaveLength(2)
})

test('quita PIX + dinheiro e acrescenta somente dinheiro à gaveta', async ({ page }, testInfo) => {
  await login(page)
  await goto(page, '/cash')
  if (!await api<ICashSession | null>(page, '/cash/current')) {
    const open = page.getByRole('button', { name: 'Abrir caixa físico', exact: true })
    await expect(open).toBeVisible()
    await page.getByLabel('Dinheiro inicial').fill('R$ 100,00')
    const opened = page.waitForResponse((response) => new URL(response.url()).pathname === '/api/cash/open' && response.request().method() === 'POST')
    await open.click()
    expect((await opened).ok()).toBe(true)
    await expect(open).toBeHidden()
  }
  await expect(page.getByRole('heading', { name: 'Resumo do dia', exact: true })).toBeVisible()
  await startQuickReceipt(page, `Divisão dinheiro ${testInfo.project.name} ${Date.now()}`)
  await page.getByRole('checkbox', { name: 'Dividir em duas formas' }).check()
  await page.getByRole('textbox', { name: /^Valor base \(sem taxas\)/ }).fill('R$ 40,00')
  const receipt = await save(page)
  const drawerBefore = await api<ICashSession>(page, '/cash/current')
  const balancesBefore = await api<ICashBalances>(page, '/cash/balances')
  await markPaid(page, receipt.id)
  const drawerAfter = await api<ICashSession>(page, '/cash/current')
  expect(drawerAfter.expectedCashCents - drawerBefore.expectedCashCents).toBe(4000)
  const balancesAfter = await api<ICashBalances>(page, '/cash/balances')
  expect(balancesAfter.pixCents - balancesBefore.pixCents).toBe(6000)
  const entries = await receiptEntries(page, receipt.id)
  expect(entries).toHaveLength(2)
  expect(entries).toEqual(expect.arrayContaining([
    expect.objectContaining({ paymentMethod: 'cash', amountCents: 4000 }),
    expect.objectContaining({ paymentMethod: 'pix', amountCents: 6000 })
  ]))
  await goto(page, '/cash')
  await expect(page.getByText('Dinheiro esperado na gaveta', { exact: true })).toBeVisible()
  await expectNoOverflow(page)
})

test('bloqueia divisão sem saldo e mantém formulário acessível em todas as telas', async ({ page }, testInfo) => {
  await login(page)
  await startQuickReceipt(page, `Divisão inválida ${testInfo.project.name} ${Date.now()}`)
  await page.getByRole('checkbox', { name: 'Dividir em duas formas' }).check()
  const secondMethod = page.getByRole('combobox', { name: 'Forma de pagamento 2', exact: true })
  await expect(secondMethod.locator('option[value="cash"]')).toHaveAttribute('disabled', '')
  await page.getByRole('textbox', { name: /^Valor base \(sem taxas\)/ }).fill('R$ 100,00')
  const requests: string[] = []
  page.on('request', (request) => {
    if (request.method() === 'POST' && new URL(request.url()).pathname === '/api/receipts') requests.push(request.url())
  })
  await page.getByRole('button', { name: 'Salvar recibo', exact: true }).click()
  await expect(page.getByRole('textbox', { name: /^Valor restante \(sem taxas\)/ })).toHaveAttribute('aria-invalid', 'true')
  await expect(page.getByText('Cada forma de pagamento deve ter um valor maior que zero e dentro do subtotal.').first()).toBeVisible()
  expect(requests).toHaveLength(0)
  const accessibility = await new AxeBuilder({ page }).include('.receipts-form').withTags(['wcag2a', 'wcag2aa', 'wcag21aa']).analyze()
  const blocking = accessibility.violations.filter(({ impact }) => impact === 'critical' || impact === 'serious')
  expect(blocking, blocking.map(({ id, help }) => `${id}: ${help}`).join('\n')).toEqual([])
  await expectNoOverflow(page)
  await page.screenshot({ path: `/tmp/receipt-split-invalid-${testInfo.project.name}.png`, fullPage: true })
})
