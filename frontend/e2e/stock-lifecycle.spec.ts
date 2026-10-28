import AxeBuilder from '@axe-core/playwright'
import { expect, test, type Page, type TestInfo } from '@playwright/test'
import type {
  ICashEntry, IPaginated, IPayableInstallment, IRealizedExpense, IReceipt,
  IStockItem, IStockPurchase, IStockSupplierDetail
} from '../server/contracts/types'

test.setTimeout(120000)
test.use({ timezoneId: 'America/Sao_Paulo', colorScheme: 'dark', actionTimeout: 15000 })

test.beforeEach(async ({ baseURL }) => {
  // These tests create financial records; never allow an external application URL.
  expect(['127.0.0.1', 'localhost', '[::1]']).toContain(new URL(baseURL!).hostname)
})

async function login(page: Page) {
  await page.goto('/login')
  await mounted(page, '.login-form')
  await page.getByLabel('CPF').fill(process.env.E2E_ADMIN_CPF || '52998224725')
  await page.getByLabel('Senha').fill(process.env.E2E_ADMIN_PASSWORD || 'change-me-admin-password')
  await page.getByRole('button', { name: 'Entrar', exact: true }).click()
  await expect(page).toHaveURL(/\/$/)
}

async function mounted(page: Page, selector: string) {
  await page.waitForFunction((selector) => {
    const element = document.querySelector(selector) as HTMLElement & { __vueParentComponent?: { isMounted: boolean } } | null
    return element?.__vueParentComponent?.isMounted === true
  }, selector)
}

async function api<T>(page: Page, path: string, method = 'GET', data?: unknown): Promise<T> {
  const cookie = (await page.context().cookies()).find(({ name }) => name === 'auth._token.laravelJWT')
  expect(cookie).toBeDefined()
  const response = await page.request.fetch(`/api${path}`, {
    method, data, headers: { Authorization: decodeURIComponent(cookie!.value) }
  })
  expect(response.ok(), `${method} ${path}: ${response.status()} ${await response.text()}`).toBe(true)
  if (response.status() === 204) return undefined as T
  return response.json()
}

async function captureSave<T>(page: Page, name: string, path: string, method = 'POST', double = false): Promise<T> {
  const saved = page.waitForResponse((response) => new URL(response.url()).pathname === `/api${path}` && response.request().method() === method)
  const button = page.getByRole('button', { name, exact: true })
  if (double) {
    // Two events in the same browser task exercise the synchronous saving guard.
    await button.evaluate((element: HTMLButtonElement) => { element.click(); element.click() })
  } else await button.click()
  const response = await saved
  expect(response.ok(), await response.text()).toBe(true)
  return response.json()
}

async function screenshot(page: Page, info: TestInfo, name: string) {
  await mounted(page, '.page')
  await expect.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  const content = page.locator('main')
  expect(await content.evaluate((element) => element.scrollWidth <= element.clientWidth)).toBe(true)
  const previousScroll = await content.evaluate((element) => element.scrollTop)
  await content.evaluate((element) => { element.scrollTop = 0 })
  await page.screenshot({ path: `/tmp/empi-stock-${info.project.name}-${name}.png`, animations: 'disabled' })
  if (await content.evaluate((element) => element.scrollHeight > element.clientHeight + 10)) {
    await content.evaluate((element) => { element.scrollTop = element.scrollHeight })
    await page.screenshot({ path: `/tmp/empi-stock-${info.project.name}-${name}-bottom.png`, animations: 'disabled' })
  }
  await content.evaluate((element, scroll) => { element.scrollTop = scroll }, previousScroll)
}

async function assertAccessible(page: Page) {
  const result = await new AxeBuilder({ page }).withTags(['wcag2a', 'wcag2aa', 'wcag21aa']).analyze()
  const blocking = result.violations.filter(({ impact }) => impact === 'critical' || impact === 'serious')
  expect(blocking, blocking.map(({ id, help }) => `${id}: ${help}`).join('\n')).toEqual([])
}

async function openEntry(page: Page, id?: string) {
  await page.goto(id ? `/stock/${id}/edit` : '/stock/new')
  await mounted(page, '.stock-entry')
}

async function sellAndReceive(page: Page, item: IStockItem, quantity: number, tag: string) {
  await page.goto('/receipts/new?quick=1')
  await mounted(page, '.receipts-form')
  await page.getByLabel('Valor da mão de obra').fill('R$ 10,00')
  await page.getByLabel('Descrição do serviço').fill(tag)
  await page.getByRole('button', { name: 'Avançar', exact: true }).click()
  await expect(page.getByRole('heading', { name: 'Produtos usados', exact: true })).toBeVisible()
  await page.getByRole('combobox', { name: 'Produto usado', exact: true }).selectOption(item.id)
  // Reject insufficient stock visibly, then complete the real sale.
  await page.getByLabel('Quantidade', { exact: true }).fill(String(item.quantity + 1))
  await page.getByRole('button', { name: 'Adicionar item', exact: true }).click()
  await expect(page.getByText(`Quantidade máxima disponível: ${item.quantity}.`, { exact: true })).toBeVisible()
  await page.getByLabel('Quantidade', { exact: true }).fill(String(quantity))
  await page.getByRole('button', { name: 'Adicionar item', exact: true }).click()
  await page.getByRole('button', { name: 'Avançar', exact: true }).click()
  await page.getByRole('button', { name: 'Avançar', exact: true }).click()
  await page.getByRole('combobox', { name: 'Forma de pagamento', exact: true }).selectOption('pix')
  const receipt = await captureSave<IReceipt>(page, 'Salvar recibo', '/receipts')
  await expect(page).toHaveURL(/\/receipts$/)
  await mounted(page, '.receipts-template__filters')
  await page.getByLabel('Cliente, telefone, placa ou serviço').fill(tag)
  const filtered = page.waitForResponse((response) => new URL(response.url()).pathname === '/api/receipts' && new URL(response.url()).searchParams.get('q') === tag)
  await page.getByRole('button', { name: 'Aplicar filtros', exact: true }).click()
  await filtered
  await expect(page.locator('.receipts-list__row')).toHaveCount(1)
  const paid = page.waitForResponse((response) => response.url().includes(`/receipts/${receipt.id}/pay`) && response.request().method() === 'POST')
  await page.getByRole('button', { name: 'Marcar como pago', exact: true }).click()
  await paid
  await expect(page.locator('.receipts-list__row').getByText('Pago', { exact: true })).toBeVisible()
  expect((await api<IReceipt>(page, `/receipts/${receipt.id}`)).status).toBe('paid')
  return receipt
}

test('cria, quita, vende, repõe e vende novamente sem duplicar cadastro ou pagamentos', async ({ page }, info) => {
  await login(page)
  const tag = `Ciclo ${info.project.name} ${Date.now()}`
  const supplierA = `Fornecedor A ${tag}`
  const supplierB = `Fornecedor B ${tag}`
  await page.goto('/stock')
  await expect(page.getByRole('link', { name: /Ver fornecedores/ })).toBeVisible()
  await expect(page.getByRole('link', { name: /Ver produtos/ })).toBeVisible()
  await screenshot(page, info, 'home')
  await assertAccessible(page)

  await openEntry(page)
  await page.getByLabel('Produto', { exact: true }).fill(tag)
  await page.getByLabel('Fornecedor', { exact: true }).fill(supplierA)
  await page.getByLabel('Quantidade a adicionar').fill('2')
  await page.getByLabel('Custo unitário', { exact: true }).fill('R$ 91,00')
  await page.getByLabel('Margem de revenda (%)', { exact: true }).fill('40')
  await page.getByRole('combobox', { name: 'Forma prevista', exact: true }).selectOption('pix')
  await screenshot(page, info, 'create')
  await assertAccessible(page)
  let writes = 0
  const track = (request: import('@playwright/test').Request) => {
    if (new URL(request.url()).pathname === '/api/stock' && request.method() === 'POST') writes++
  }
  page.on('request', track)
  let item = await captureSave<IStockItem>(page, 'Cadastrar produto e confirmar entrada', '/stock', 'POST', true)
  await expect(page).toHaveURL(/\/stock\/products$/)
  page.off('request', track)
  expect(writes).toBe(1)
  expect(item.quantity).toBe(2)
  expect(item.resalePriceCents).toBe(12740)
  const row = page.locator('tr:visible, .stock-card:visible').filter({ hasText: tag })
  await expect(row).toHaveCount(1)
  await expect(row.getByRole('button', { name: 'Editar', exact: true })).toHaveCount(1)
  await expect(row.getByRole('button', { name: /Adicionar.*estoque/ })).toHaveCount(0)
  const initial = (await api<IStockPurchase[]>(page, '/stock/purchases')).filter((purchase) => purchase.items.some(({ stockItemId }) => stockItemId === item.id))
  expect(initial).toHaveLength(1)
  expect(initial[0]!.totalCents).toBe(18200)
  const installment = initial[0]!.installments[0]!
  expect(installment.status).toBe('pending')

  await page.goto('/payables')
  await page.getByRole('searchbox').fill(supplierA)
  const payable = page.locator('tr:visible, .payable-card:visible').filter({ hasText: supplierA })
  await expect(payable).toHaveCount(1)
  await payable.getByRole('combobox').selectOption('pix')
  await payable.getByRole('button', { name: 'Registrar pagamento', exact: true }).click()
  await captureSave(page, 'Confirmar pagamento', `/payables/${installment.id}/pay`)
  await expect(payable).toHaveCount(0)
  const paidPurchase = await api<IStockPurchase>(page, `/stock/purchases/${initial[0]!.id}`)
  expect(paidPurchase.installments[0]!.status).toBe('paid')
  const ledger = await api<ICashEntry[]>(page, '/cash/daily-entries')
  const outgoing = ledger.filter(({ id }) => id === paidPurchase.installments[0]!.cashEntryId)
  expect(outgoing).toHaveLength(1)
  expect(outgoing[0]).toMatchObject({ amountCents: -18200, direction: 'out', kind: 'stock_payment', paymentMethod: 'pix' })
  const realized = await api<IPaginated<IRealizedExpense>>(page, '/financial/expenses?origin=stock&limit=100&offset=0')
  expect(realized.data.filter(({ installmentId }) => installmentId === installment.id)).toHaveLength(1)

  await sellAndReceive(page, item, 2, `${tag} venda1`)
  item = await api<IStockItem>(page, `/stock/${item.id}`)
  expect(item.quantity).toBe(0)
  expect(item.usedQuantity).toBe(2)
  // The autocomplete must include products with no available stock.
  await openEntry(page)
  await page.getByLabel('Fornecedor', { exact: true }).fill(supplierA)
  await page.getByLabel('Produto', { exact: true }).fill(tag)
  await expect(page.getByRole('option').filter({ hasText: tag })).toBeVisible()
  await screenshot(page, info, 'zero-stock-autocomplete')
  await page.getByLabel('Produto', { exact: true }).press('ArrowDown')
  await page.getByLabel('Produto', { exact: true }).press('Enter')
  await expect(page.getByText('Adicionar quantidade ao produto existente', { exact: true })).toBeVisible()
  expect((await api<IStockItem>(page, `/stock/${item.id}`)).quantity).toBe(0)

  await openEntry(page, item.id)
  await expect(page.getByText(`Saldo atual de ${tag}`, { exact: true })).toBeVisible()
  await page.getByRole('checkbox').check()
  await page.getByLabel('Fornecedor', { exact: true }).fill(supplierA)
  await page.getByLabel('Quantidade a adicionar').fill('3')
  await page.getByLabel('Custo unitário da entrada', { exact: true }).fill('R$ 91,00')
  await page.getByRole('combobox', { name: 'Forma prevista', exact: true }).selectOption('pix')
  await screenshot(page, info, 'edit-restock')
  await captureSave(page, 'Salvar e confirmar entrada', `/stock/${item.id}`, 'PUT')
  await expect(page).toHaveURL(/\/stock\/products$/)
  item = await api<IStockItem>(page, `/stock/${item.id}`)
  expect(item.quantity).toBe(3)
  await sellAndReceive(page, item, 1, `${tag} venda2`)
  item = await api<IStockItem>(page, `/stock/${item.id}`)
  expect(item.quantity).toBe(2)
  expect(item.usedQuantity).toBe(3)

  // A second supplier feeds the same product rather than creating a second balance.
  await openEntry(page, item.id)
  await page.getByRole('checkbox').check()
  await page.getByLabel('Fornecedor', { exact: true }).fill(supplierB)
  await page.getByLabel('Quantidade a adicionar').fill('2')
  await page.getByLabel('Custo unitário da entrada', { exact: true }).fill('R$ 101,00')
  await captureSave(page, 'Salvar e confirmar entrada', `/stock/${item.id}`, 'PUT')
  await expect(page).toHaveURL(/\/stock\/products$/)
  item = await api<IStockItem>(page, `/stock/${item.id}`)
  expect(item.quantity).toBe(4)
  expect(item.costCents).toBe(9600)
  const purchases = (await api<IStockPurchase[]>(page, '/stock/purchases')).filter((purchase) => purchase.items.some(({ stockItemId }) => stockItemId === item.id))
  expect(purchases).toHaveLength(3)
  expect(purchases.flatMap(({ installments }) => installments).filter(({ status }) => status === 'pending')).toHaveLength(2)

  for (const [supplier, count, bought, paid] of [[supplierA, 2, 45500, 18200], [supplierB, 1, 20200, 0]] as const) {
    const detail = await api<IStockSupplierDetail>(page, `/stock/suppliers/detail?supplierName=${encodeURIComponent(supplier)}`)
    expect(detail).toMatchObject({ purchaseCount: count, productCount: 1, totalCents: bought, paidCents: paid, pendingCents: bought - paid })
    expect(detail.products[0]).toMatchObject({ id: item.id, quantity: 4 })
    await page.goto(`/stock/suppliers/detail?supplierName=${encodeURIComponent(supplier)}`)
    await expect(page.getByRole('heading', { name: supplier, exact: true })).toBeVisible()
    await expect(page.locator('.supplier-purchase')).toHaveCount(count)
    await screenshot(page, info, supplier === supplierA ? 'supplier-paid-history' : 'supplier-other-history')
    await assertAccessible(page)
  }
  await page.getByRole('link', { name: `Adicionar quantidade a ${tag}`, exact: true }).click()
  await expect(page.getByLabel('Fornecedor', { exact: true })).toHaveValue(supplierB)
  await expect(page.getByLabel('Fornecedor', { exact: true })).toBeDisabled()
  await expect(page.getByLabel('Produto', { exact: true })).toBeDisabled()
  await page.getByLabel('Quantidade a adicionar').fill('1')
  await captureSave(page, 'Confirmar entrada no estoque', '/stock')
  await expect(page).toHaveURL(/\/stock\/suppliers\/detail/)
  expect((await api<IStockItem>(page, `/stock/${item.id}`)).quantity).toBe(5)

  const beforeMetadata = (await api<IStockPurchase[]>(page, '/stock/purchases')).length
  const pendingBefore = (await api<IPayableInstallment[]>(page, '/payables')).length
  await openEntry(page, item.id)
  await expect(page.getByRole('checkbox')).not.toBeChecked()
  await page.getByLabel('Descrição', { exact: true }).fill('Descrição corrigida sem entrada')
  await captureSave(page, 'Salvar produto', `/stock/${item.id}`, 'PUT')
  await expect(page).toHaveURL(/\/stock\/products$/)
  expect((await api<IStockPurchase[]>(page, '/stock/purchases')).length).toBe(beforeMetadata)
  expect((await api<IPayableInstallment[]>(page, '/payables')).length).toBe(pendingBefore)
  expect(await api<IStockItem>(page, `/stock/${item.id}`)).toMatchObject({ quantity: 5, description: 'Descrição corrigida sem entrada' })
  await screenshot(page, info, 'products-final')
})

test('busca fora da primeira página, permite duplicata explícita e rejeita entradas inválidas sem gravar', async ({ page }, info) => {
  await login(page)
  const tag = `Busca ${info.project.name} ${Date.now()}`
  const supplier = `Fornecedor ${tag}`
  const item = await api<IStockItem>(page, '/stock', 'POST', { name: tag, description: '', costCents: 1000, markupPercent: 20, quantity: 1 })
  // Push the target beyond the 10-row listing, without altering production data.
  for (let index = 0; index < 11; index++) {
    await api(page, '/stock', 'POST', { name: `Posterior ${tag} ${index}`, description: '', costCents: 1000, markupPercent: 20, quantity: 0 })
  }
  const firstPage = await api<IPaginated<IStockItem>>(page, '/stock?limit=10&offset=0')
  expect(firstPage.data.some(({ id }) => id === item.id)).toBe(false)
  await openEntry(page)
  await page.getByLabel('Fornecedor', { exact: true }).fill(supplier)
  await page.getByLabel('Produto', { exact: true }).fill(tag)
  await expect(page.getByRole('option').filter({ hasText: tag }).filter({ hasNotText: 'Posterior' })).toBeVisible()
  await page.getByRole('option').filter({ hasText: tag }).filter({ hasNotText: 'Posterior' }).getByRole('button').click()
  await expect(page.getByLabel('Produto', { exact: true })).toBeDisabled()
  await page.getByLabel('Quantidade a adicionar').fill('0')
  await page.getByRole('button', { name: 'Confirmar entrada no estoque', exact: true }).click()
  await expect(page.getByRole('alert')).toContainText('Informe uma quantidade inteira maior que zero.')
  await screenshot(page, info, 'validation-error')
  await page.getByLabel('Quantidade a adicionar').fill('2')
  await page.getByLabel('Valor', { exact: true }).fill('R$ 1,00')
  await page.getByRole('button', { name: 'Confirmar entrada no estoque', exact: true }).click()
  await expect(page.getByRole('alert')).toContainText('A soma das parcelas deve ser igual ao total da entrada.')
  await page.getByLabel('Número de parcelas', { exact: true }).fill('49')
  await page.getByRole('button', { name: 'Confirmar entrada no estoque', exact: true }).click()
  await expect(page.getByRole('alert')).toContainText('Informe de 1 a 48 parcelas.')
  expect((await api<IStockItem>(page, `/stock/${item.id}`)).quantity).toBe(1)
  expect((await api<IStockPurchase[]>(page, '/stock/purchases')).some(({ items }) => items.some(({ stockItemId }) => stockItemId === item.id))).toBe(false)
  await assertAccessible(page)
  await sellAndReceive(page, item, 1, `${tag} venda fora da primeira página`)
  expect((await api<IStockItem>(page, `/stock/${item.id}`)).quantity).toBe(0)

  await openEntry(page)
  await page.getByLabel('Fornecedor', { exact: true }).fill(supplier)
  await page.getByLabel('Produto', { exact: true }).fill(tag)
  await expect(page.getByRole('option').filter({ hasText: tag }).first()).toBeVisible()
  await page.getByRole('button', { name: 'Criar outro produto', exact: true }).click()
  await expect(page.getByLabel('Produto', { exact: true })).toBeEnabled()
  await page.getByLabel('Custo unitário', { exact: true }).fill('R$ 15,00')
  await page.getByLabel('Quantidade a adicionar').fill('1')
  const another = await captureSave<IStockItem>(page, 'Cadastrar produto e confirmar entrada', '/stock')
  expect(another.name).toBe(tag)
  expect(another.id).not.toBe(item.id)
  expect(another.quantity).toBe(1)
  expect((await api<IStockItem>(page, `/stock/${item.id}`)).quantity).toBe(0)
  await api(page, `/stock/${another.id}`, 'DELETE')
  await page.goto(`/stock/suppliers/detail?supplierName=${encodeURIComponent(supplier)}`)
  const archived = page.locator('.supplier-product').filter({ hasText: tag })
  await expect(archived.getByText('Arquivado', { exact: true })).toBeVisible()
  await expect(archived.getByRole('link', { name: /Adicionar quantidade/ })).toHaveCount(0)
  await expect(page.getByText('Cancelada', { exact: true }).first()).toBeVisible()
  await screenshot(page, info, 'supplier-archived-history')
})
