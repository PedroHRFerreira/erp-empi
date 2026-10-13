## Estado atual — fidelidade da arte original (12/10/2026)

As 54 referências foram importadas no Figma:
https://www.figma.com/design/Z9hsPOVKMQYmiZSFLI0jKq

Ainda são imagens únicas, não desenhos em camadas. A reconstrução de textos,
fotos e elementos editáveis não foi concluída.

O catálogo agora abre e exporta a própria imagem original, em suas dimensões
nativas, sem substituir a composição por um dos sete layouts genéricos.
A legenda continua editável. Os controles de edição visual ficam disponíveis
nas composições básicas; não são apresentados como capazes de editar textos
embutidos nas referências. Criações antigas com templates do catálogo também
passam a mostrar a imagem original; seus dados persistidos não são apagados.

As seções abaixo registram o trabalho anterior e os recursos preparados.
As composições genéricas descritas anteriormente não são mais usadas para
renderizar os templates do catálogo.

# Modelos da EMPI

Catálogo importado do perfil https://www.instagram.com/empi.autocenter_/ em 12/10/2026: **46 publicações e 54 referências**, contando separadamente as páginas dos três carrosséis (3, 5 e 3 páginas). Capas dos Reels entram como referências de imagem; este módulo exporta artes estáticas.

## Como usar

Abra **Marketing → Modelos**, busque pelo nome ou categoria e escolha **Usar modelo**. A aba Criar recebe título, descrição, legenda, cores e composição; ajuste os campos e a foto, salve ou baixe o PNG. O formato inicial é Quadrado 1080×1080, com Feed e Story também disponíveis.

A prévia da biblioteca é a referência publicada no Instagram. A versão editável é uma adaptação em camadas independentes e pode variar da composição original. Não existe conversão automática de todos os textos embutidos na imagem em camadas. As referências ficam preservadas e também podem ser baixadas sem alterações.

Sete composições cobrem os estilos observados: institucional, foto com texto lateral, foto superior, checklist, promoção, diagnóstico técnico e apresentação da marca. Descrições com linhas separadas formam os itens nos modelos institucional e checklist. Nome, contato, cores, foto e logo podem ser ajustados no editor; o logo configurado na identidade visual substitui o logo padrão da EMPI.

## Integração

O catálogo versionado fica em `internal/domain/marketing/templates/catalog.json`. O backend incorpora esse arquivo, valida os IDs ao salvar criações e disponibiliza `GET /api/marketing/templates` apenas para administradores. As criações guardam o ID do modelo e os textos/cores próprios. Os IDs antigos `bold` e `minimal` continuam aceitos.

As imagens locais ficam em `frontend/public/marketing/references/`. O catálogo não depende de estar conectado ao Instagram durante o uso e não sincroniza novas publicações automaticamente. Fotos adicionadas ao editor seguem a validação de upload existente.

Três fundos limpos, preparados a partir das referências com a ferramenta interna de edição de imagens, ficam em `frontend/public/marketing/backgrounds/parts.png`, `workshop.png` e `travel.png`. Esses fundos foram inspecionados e não contêm os textos ou logos antigos. O logo padrão preservado está em `frontend/public/marketing/brand-logo.jpg`. A ferramenta de IA foi usada somente na preparação desses arquivos; edição e exportação no aplicativo usam Canvas local.

## Prompts finais usados na preparação dos fundos

+### workshop.png

Use case: precise-object-edit. Edit target: supplied EMPI mechanic artwork. Create a clean reusable photographic background by removing ALL overlaid logos, advertising text, phone/address, graphic icons and badges. Keep the mechanic, his identity, pose, hands, clothing, car, wheel well and power tool unchanged. Reconstruct the photographic workshop scene naturally in areas covered by graphics; retain realistic lighting and the dark automotive mood. Output a square full-bleed photograph only. No text, no logos, no watermarks, no invented extra people. This is a background for editable marketing templates; graphics will be drawn separately in Canvas.

### parts.png

Use case: precise-object-edit. Edit target: supplied EMPI institutional artwork. Extract/recreate ONLY the photographic car parts arrangement in the lower-left: brake disc, brake pads, battery, belt, ratchet tool, oil filter and spark plugs. Remove ALL advertising text, logos, icons, badges and footer. Preserve the parts' recognizable appearance/materials, proportions and photographic realism. Place these same parts as an orderly product photo on a clean warm off-white studio floor and background. Square photo, cluster occupying lower half, upper half negative space. No text, no logo, no watermark. Editable marketing graphics are added separately in Canvas.

### travel.png

Use case: precise-object-edit. Edit target: supplied EMPI family road-trip artwork. Remove ALL overlaid EMPI badge/logo and all headlines, colored CTA bar, text and icons. Keep the family car interior, people, identities, clothing, smiles, seating positions, seatbelts and warm sunset lighting as close to original as possible. Reconstruct only regions hidden by advertising graphics into a natural photograph. Square full-bleed clean photo, no text, no logos or watermark. This will be a reusable photo background with independently editable Canvas typography.
