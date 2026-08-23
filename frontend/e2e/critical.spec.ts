import AxeBuilder from '@axe-core/playwright'
import { expect, test, type Page } from '@playwright/test'

const adminCpf = process.env.E2E_ADMIN_CPF || '52998224725'
const adminPassword = process.env.E2E_ADMIN_PASSWORD || 'change-me-admin-password'

async function login(page: Page) {
  await page.goto('/login')
  await page.getByLabel('CPF').fill(adminCpf)
  await page.getByLabel('Senha').fill(adminPassword)
  await page.getByRole('button', { name: 'Entrar' }).click()
  await expect(page).toHaveURL(/\/$/)
  await expect(page.getByRole('heading', { name: 'Métricas' })).toBeVisible()
}

async function expectNoCriticalAccessibilityViolations(page: Page) {
  const result = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'])
    .analyze()

  const blocking = result.violations.filter(({ impact }) => impact === 'critical' || impact === 'serious')
  expect(blocking, blocking.map(({ id, help }) => `${id}: ${help}`).join('\n')).toEqual([])
}

async function ensureCashIsOpen(page: Page) {
  await page.goto('/cash')
  const openButton = page.getByRole('button', { name: 'Abrir caixa' })
  if (await openButton.isVisible()) {
    await page.getByLabel('Dinheiro inicial').fill('R$ 100,00')
    await openButton.click()
  }
  await expect(page.getByRole('heading', { name: 'Resumo do dia' })).toBeVisible()
}

test('rejeita credenciais inválidas com feedback acessível', async ({ page }) => {
  await page.goto('/login')
  await page.getByLabel('CPF').fill('52998224725')
  await page.getByLabel('Senha').fill('senha-incorreta')
  await page.getByRole('button', { name: 'Entrar' }).click()
  await expect(page.getByText('CPF ou senha inválidos.').first()).toBeVisible()
  await expectNoCriticalAccessibilityViolations(page)
})

test('alterna e persiste os temas WattVision', async ({ page }) => {
  await page.goto('/login')
  await page.evaluate(() => window.localStorage.setItem('empi-theme', 'dark'))
  await page.reload()
  await page.getByRole('button', { name: 'Ativar tema claro' }).click()
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'light')

  await page.reload()
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'light')
  await page.getByRole('button', { name: 'Ativar tema escuro' }).click()
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark')
  await expectNoCriticalAccessibilityViolations(page)
})

test('mantém o seletor de tema fora das ações de recibos', async ({ page }) => {
  await login(page)
  await page.goto('/receipts')

  const themeToggle = page.getByRole('button', { name: /Ativar tema (claro|escuro)/ })
  await expect(themeToggle).toBeVisible()

  const themeBox = await themeToggle.boundingBox()
  expect(themeBox).not.toBeNull()

  for (const name of ['Recibo rápido', 'Adicionar']) {
    const action = page.getByRole('button', { name, exact: true })
    await expect(action).toBeVisible()
    const actionBox = await action.boundingBox()
    expect(actionBox).not.toBeNull()

    const overlaps = !(
      themeBox!.x + themeBox!.width <= actionBox!.x
      || actionBox!.x + actionBox!.width <= themeBox!.x
      || themeBox!.y + themeBox!.height <= actionBox!.y
      || actionBox!.y + actionBox!.height <= themeBox!.y
    )
    expect(overlaps, `O seletor de tema está sobrepondo o botão ${name}`).toBe(false)
  }
})

test('percorre os módulos autenticados sem erros de console ou API', async ({ page }) => {
  const runtimeErrors: string[] = []
  let activeRoute = '/login'
  page.on('console', (message) => {
    if (message.type() === 'error') runtimeErrors.push(`${activeRoute} console: ${message.text()}`)
  })
  page.on('pageerror', (error) => runtimeErrors.push(`${activeRoute} page: ${error.message}`))
  page.on('response', (response) => {
    if (response.status() >= 500) runtimeErrors.push(`${response.status()}: ${response.url()}`)
  })

  await login(page)
  const routes = [
    ['/', 'Métricas'],
    ['/goals', 'Metas'],
    ['/receipts', 'Recibos'],
    ['/recovery', 'Recuperação'],
    ['/clients', 'Clientes'],
    ['/expenses', 'Gastos'],
    ['/cash', 'Caixa'],
    ['/payables', 'Contas a pagar'],
    ['/stock', 'Estoque'],
    ['/stock/purchases', 'Compras de estoque'],
    ['/profile', 'Perfil']
  ] as const

  for (const [route, heading] of routes) {
    activeRoute = route
    await page.goto(route)
    await expect(page.getByRole('heading', { name: heading, exact: true }).first()).toBeVisible()
    await expectNoCriticalAccessibilityViolations(page)
  }

  const creationRoutes = [
    ['/receipts/new', 'Adicionar recibo'],
    ['/stock/new', 'Adicionar produto ao estoque']
  ] as const

  for (const [route, heading] of creationRoutes) {
    activeRoute = route
    await page.goto(route)
    await expect(page.getByRole('heading', { name: heading, exact: true }).first()).toBeVisible()
    await expectNoCriticalAccessibilityViolations(page)
  }

  expect(runtimeErrors).toEqual([])
})

test('orquestra o drawer de notificações financeiras com teclado e navegação', async ({ page }) => {
  await login(page)

  const trigger = page.locator('.app-sidebar__notification:visible').first()
  await expect(trigger).toBeVisible()
  await expect(trigger).toHaveAttribute('aria-expanded', 'false')

  await trigger.click()
  const dialog = page.getByRole('dialog', { name: 'Notificações' })
  await expect(dialog).toBeVisible()
  await expect(trigger).toHaveAttribute('aria-expanded', 'true')
  await expect(page.getByRole('button', { name: 'Fechar notificações' }).last()).toBeFocused()

  await page.keyboard.press('Escape')
  await expect(dialog).toBeHidden()
  await expect(trigger).toBeFocused()

  await trigger.click()
  await page.locator('.financial-drawer__backdrop').click({ position: { x: 8, y: 8 } })
  await expect(dialog).toBeHidden()

  await trigger.click()
  await page.getByRole('link', { name: 'Ver contas a pagar' }).click()
  await expect(page).toHaveURL(/\/payables$/)
  await expect(dialog).toBeHidden()
  await expect(page.getByRole('heading', { name: 'Contas a pagar', exact: true })).toBeVisible()
})

test('cria produto e valida o fluxo principal de estoque', async ({ page }, testInfo) => {
  await login(page)
  await page.goto('/stock/new')
  const productName = `Produto E2E ${testInfo.project.name} ${Date.now()}`
  await page.getByLabel('Produto').fill(productName)
  await page.getByLabel('Fornecedor').fill(`Fornecedor ${testInfo.project.name}`)
  await page.getByLabel('Quantidade comprada').fill('8')
  await page.getByLabel('Custo unitário').fill('R$ 25,00')
  await page.getByLabel('Margem de revenda (%)').fill('30')
  await page.getByLabel('Número de parcelas').fill('3')
  await expect(page.getByLabel('Valor')).toHaveCount(3)
  const installmentDates = page.getByLabel('Vencimento', { exact: true })
  const preservedLastDueDate = await installmentDates.nth(2).inputValue()
  await page.getByLabel('Forma prevista').nth(2).selectOption('pix')
  await page.getByRole('button', { name: 'Remover 2ª parcela' }).click()
  await expect(page.getByLabel('Valor')).toHaveCount(2)
  await expect(page.getByLabel('Valor').nth(0)).toHaveValue(/R\$\s100,00/)
  await expect(page.getByLabel('Valor').nth(1)).toHaveValue(/R\$\s100,00/)
  await expect(installmentDates.nth(1)).toHaveValue(preservedLastDueDate)
  await expect(page.getByLabel('Forma prevista').nth(1)).toHaveValue('pix')
  await page.getByLabel('Quantidade comprada').fill('7')
  await expect(page.getByLabel('Valor').nth(0)).toHaveValue(/R\$\s87,50/)
  await expect(page.getByLabel('Valor').nth(1)).toHaveValue(/R\$\s87,50/)
  await expect(installmentDates.nth(1)).toHaveValue(preservedLastDueDate)
  await expect(page.getByLabel('Forma prevista').nth(1)).toHaveValue('pix')
  await page.getByLabel('Quantidade comprada').fill('8')
  await expect(page.getByLabel('Valor').nth(0)).toHaveValue(/R\$\s100,00/)
  await expect(page.getByLabel('Valor').nth(1)).toHaveValue(/R\$\s100,00/)
  await expect(page.getByRole('button', { name: 'Remover 1ª parcela' })).toBeEnabled()
  await page.getByRole('button', { name: 'Remover 1ª parcela' }).click()
  await expect(page.getByRole('button', { name: 'Remover 1ª parcela' })).toBeDisabled()
  await expect(page.getByLabel('Valor')).toHaveValue(/R\$\s200,00/)
  await page.getByRole('button', { name: 'Cadastrar produto e confirmar entrada' }).click()
  await expect(page).toHaveURL(/\/stock$/)
  await expect(page.getByText(productName, { exact: true }).filter({ visible: true })).toBeVisible()
})

test('cria recibo rápido pela jornada completa do wizard', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop', 'Jornada mutável executada uma vez na base compartilhada')
  await login(page)
  await page.goto('/receipts')
  await page.getByRole('button', { name: 'Recibo rápido' }).click()
  await page.getByLabel('Valor da mão de obra').fill('R$ 120,00')
  await page.getByLabel('Descrição do serviço').fill('Diagnóstico eletrônico E2E')
  await page.getByRole('button', { name: 'Avançar' }).click()
  await page.getByRole('button', { name: 'Avançar' }).click()
  await page.getByRole('button', { name: 'Avançar' }).click()
  await page.getByRole('button', { name: 'Salvar recibo' }).click()
  await expect(page).toHaveURL(/\/receipts$/)
  await expect(page.getByText('Recibo rápido').last()).toBeVisible()
  await expect(page.getByText('R$ 120,00').last()).toBeVisible()
})

test('registra gasto e atualiza o resumo financeiro', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop', 'Jornada mutável executada uma vez na base compartilhada')
  await login(page)
  await ensureCashIsOpen(page)
  await page.goto('/expenses')
  await page.getByRole('button', { name: 'Adicionar' }).click()
  const description = `Gasto E2E ${Date.now()}`
  await page.getByLabel('Descrição').fill(description)
  await page.getByLabel('Categoria').selectOption('outros')
  await page.getByPlaceholder('R$ 200,00').fill('R$ 35,00')
  await page.getByRole('button', { name: 'Salvar gasto' }).click()
  await expect(page.getByText(description)).toBeVisible()
})

test('abre o caixa diário e apresenta o resumo operacional', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop', 'Jornada mutável executada uma vez na base compartilhada')
  await login(page)
  await ensureCashIsOpen(page)
  await expect(page.getByText('Dinheiro esperado na gaveta')).toBeVisible()
})

test('encerra a sessão e protege rotas privadas', async ({ page }) => {
  await login(page)
  const mobileMenu = page.getByRole('button', { name: 'Abrir menu principal' })
  if (await mobileMenu.isVisible()) await mobileMenu.click()
  await page.getByRole('button', { name: 'Sair' }).click()
  await expect(page).toHaveURL(/\/login$/)
  await page.goto('/profile')
  await expect(page).toHaveURL(/\/login$/)
})
