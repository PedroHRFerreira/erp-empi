import { expect, test, type Page } from '@playwright/test'

async function openMarketing(page: Page, openEditor = true) {
  const response = await page.request.post('/api/auth/login', { data: {
    cpf: process.env.E2E_ADMIN_CPF || '52998224725',
    password: process.env.E2E_ADMIN_PASSWORD || 'change-me-admin-password'
  } })
  expect(response.ok()).toBe(true)
  const login = await response.json()
  const baseURL = new URL(response.url()).origin
  await page.context().addCookies([{ name: 'auth._token.laravelJWT', value: encodeURIComponent(`Bearer ${login.tokens.accessToken}`), url: baseURL }])
  await page.goto('/marketing')
  await expect(page.getByRole('heading', { name: 'Marketing', exact: true })).toBeVisible()
  if (openEditor) { await page.getByRole('tab', { name: 'Criar', exact: true }).click(); await page.getByRole('button', { name: 'Composição Destaque', exact: true }).click() }
}

test('cria, exporta localmente, reabre e duplica uma divulgação', async ({ page }) => {
  await openMarketing(page)
  const title = `Revisão preventiva ${Date.now()}`
  const editor = page.getByRole('region', { name: 'Editar criação' })
  await editor.getByLabel('Título', { exact: true }).fill(title)
  await expect(editor.getByLabel('Descrição', { exact: true })).not.toBeVisible()
  await editor.getByText('Mais opções', { exact: true }).click()
  await editor.getByLabel('Descrição', { exact: true }).fill('Segurança e cuidado para seu veículo.')
  await page.getByRole('button', { name: 'Revisar e finalizar' }).click()
  await editor.getByRole('button', { name: 'Sugestão 1' }).click()
  await expect(editor.getByLabel('Texto para publicar')).toHaveValue(new RegExp(title))
  await expect(page.locator('.marketing-editor__preview canvas')).toHaveAttribute('width', '1080')
  await page.waitForTimeout(300)
  const calls: string[] = []
  page.on('request', request => { if (request.url().includes('/api/')) calls.push(request.url()) })
  await page.getByRole('button', { name: 'Voltar', exact: true }).click()
  await editor.getByLabel('Chamada para ação').fill('Agende sua revisão')
  await editor.getByLabel('Formato').selectOption('story')
  await expect(page.locator('.marketing-editor__preview canvas')).toHaveAttribute('height', '1920')
  await page.getByRole('button', { name: 'Revisar e finalizar' }).click()
  const download = page.waitForEvent('download')
  await page.getByRole('button', { name: 'Baixar arte em PNG' }).click()
  expect((await download).suggestedFilename()).toBe('marketing-story.png')
  expect(calls, 'Edição e exportação devem ocorrer sem chamadas de API').toEqual([])
  await page.getByRole('button', { name: 'Salvar criação', exact: true }).click()
  await expect(page.getByText('Criação salva.', { exact: true })).toBeVisible()
  await page.getByRole('tab', { name: 'Minhas criações' }).click()
  const card = page.locator('article').filter({ has: page.getByRole('heading', { name: title, exact: true }) })
  await card.getByRole('button', { name: 'Abrir', exact: true }).click()
  await expect(editor.getByLabel('Título', { exact: true })).toHaveValue(title)
  await expect(editor.getByLabel('Formato')).toHaveValue('story')
  await page.getByRole('tab', { name: 'Minhas criações' }).click()
  await card.getByRole('button', { name: 'Duplicar', exact: true }).click()
  await expect(editor.getByLabel('Título', { exact: true })).toHaveValue(`${title} (cópia)`)
  page.once('dialog', dialog => dialog.dismiss())
  await page.getByRole('tab', { name: 'Minhas criações' }).click()
  await expect(editor).toBeVisible()
  page.once('dialog', dialog => dialog.accept())
  await page.getByRole('tab', { name: 'Minhas criações' }).click()
  await card.getByRole('button', { name: `Excluir ${title}`, exact: true }).click()
  await page.getByRole('button', { name: 'Confirmar exclusão' }).click()
  await expect(card).toHaveCount(0)
})

test('rejeita foto inválida e mantém o texto da criação', async ({ page }) => {
  await openMarketing(page)
  const editor = page.getByRole('region', { name: 'Editar criação' })
  await editor.getByLabel('Título', { exact: true }).fill('Meu rascunho preservado')
  await editor.getByLabel('Foto da criação').setInputFiles({ name: 'arquivo.txt', mimeType: 'text/plain', buffer: Buffer.from('não é uma foto') })
  await expect(page.getByText('Escolha uma imagem PNG ou JPEG de até 5 MB.', { exact: true })).toBeVisible()
  await expect(editor.getByLabel('Título', { exact: true })).toHaveValue('Meu rascunho preservado')
})

test('preserva a criação quando o servidor falha ao salvar', async ({ page }) => {
  await openMarketing(page)
  const editor = page.getByRole('region', { name: 'Editar criação' })
  await editor.getByLabel('Título', { exact: true }).fill('Oferta que preciso manter')
  await page.getByRole('button', { name: 'Revisar e finalizar' }).click()
  await page.route('**/api/marketing/creations', async route => {
    if (route.request().method() === 'POST') await route.fulfill({ status: 503, contentType: 'application/json', body: JSON.stringify({ message: 'Falha temporária' }) })
    else await route.continue()
  })
  await page.getByRole('button', { name: 'Salvar criação', exact: true }).click()
  await expect(page.getByText('Não foi possível salvar a criação. Seu conteúdo continua disponível.', { exact: true })).toBeVisible()
  await expect(editor.getByLabel('Título', { exact: true })).toHaveValue('Oferta que preciso manter')
  await expect(page.getByRole('button', { name: 'Salvar criação', exact: true })).toBeEnabled()
})

test('preserva a composição original na prévia e no PNG', async ({ page }) => {
  await openMarketing(page, false)
  const gallery = page.getByRole('region', { name: 'Modelos da EMPI' })
  await gallery.getByLabel('Buscar modelos').fill('feriado chegando')
  await gallery.getByRole('button', { name: 'Usar modelo Feriado chegando!', exact: true }).click()
  const canvas = page.locator('.marketing-editor__preview canvas')
  await expect(canvas).toHaveAttribute('width', '1254')
  await expect(canvas).toHaveAttribute('height', '1254')
  await expect(page.getByLabel('Título', { exact: true })).toHaveCount(0)
  const equal = await canvas.evaluate(async node => {
    const actual = node as HTMLCanvasElement
    const image = new Image()
    image.src = '/marketing/references/DctxeNQD7Jp.jpg'
    await image.decode()
    const expected = document.createElement('canvas')
    expected.width = image.naturalWidth; expected.height = image.naturalHeight
    expected.getContext('2d')!.drawImage(image, 0, 0)
    return actual.toDataURL() === expected.toDataURL()
  })
  expect(equal).toBe(true)
  const requests: string[] = []
  page.on('request', req => { if (req.url().includes('/api/')) requests.push(req.url()) })
  await page.getByRole('button', { name: 'Revisar e finalizar' }).click()
  await page.getByLabel('Texto para publicar').fill('Legenda independente da arte original')
  const downloaded = page.waitForEvent('download')
  await page.getByRole('button', { name: 'Baixar arte em PNG' }).click()
  const download = await downloaded
  const stream = await download.createReadStream()
  const chunks: Buffer[] = []
  for await (const chunk of stream!) chunks.push(Buffer.from(chunk))
  const png = Buffer.concat(chunks)
  expect(png.readUInt32BE(16)).toBe(1254)
  expect(png.readUInt32BE(20)).toBe(1254)
  expect(requests).toEqual([])
})


test('personalização compacta mostra cada alteração na prévia', async ({ page }) => {
  await openMarketing(page, false)
  const gallery = page.getByRole('region', { name: 'Modelos da EMPI' })
  await gallery.getByLabel('Buscar modelos').fill('feriado chegando')
  await gallery.getByRole('button', { name: 'Usar modelo Feriado chegando!', exact: true }).click()
  const editor = page.getByRole('region', { name: 'Editar criação' })
  await expect(editor.getByLabel('Texto principal')).toBeVisible()
  await expect(editor.getByLabel('Foto da criação')).toBeVisible()
  await expect(editor.getByLabel('Nome da criação')).not.toBeVisible()
  await expect(editor.getByLabel('Formato')).not.toBeVisible()
  const canvas = page.locator('.marketing-editor__preview canvas')
  await expect(canvas).toHaveAttribute('width', '1254')
  await expect(page.getByText('Prévia atualizada', { exact: true })).toBeVisible()
  const original = await canvas.evaluate(node => (node as HTMLCanvasElement).toDataURL())
  const requests: string[] = []
  page.on('request', req => { if (req.url().includes('/api/')) requests.push(req.url()) })
  await editor.getByLabel('Texto principal').fill('OFERTA ESPECIAL')
  await expect.poll(() => canvas.evaluate(node => (node as HTMLCanvasElement).toDataURL())).not.toBe(original)
  expect(requests).toEqual([])
  await editor.getByLabel('Texto principal').scrollIntoViewIfNeeded()
  const preview = await canvas.boundingBox()
  expect(preview!.y).toBeGreaterThanOrEqual(page.viewportSize()!.width < 1200 ? 86 : 0)
  expect(preview!.y + preview!.height).toBeLessThanOrEqual(page.viewportSize()!.height)
  await editor.getByText('Mais opções', { exact: true }).click()
  await expect(editor.getByLabel('Nome da criação')).toBeVisible()
  await expect(editor.getByLabel('Texto principal')).toHaveValue('OFERTA ESPECIAL')
})
