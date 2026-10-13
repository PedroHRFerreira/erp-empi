# Validação visual do Marketing

Data: 12/10/2026. Aplicação validada em `http://127.0.0.1:3004/marketing` com Chromium instalado nas dependências existentes do frontend. Backend local em `http://localhost:8080`. A validação ocorreu no servidor de desenvolvimento; a publicação final em `localhost:3000` precisa usar o mesmo código e catálogo.

## Fluxos exercitados

- Galeria carregada, seleção do primeiro modelo e edição do primeiro texto.
- Galeria e editor nos tamanhos 1440 × 1000, 1024 × 1000 e 390 × 1000: nenhuma rolagem horizontal da página.
- Formatos original (1179 × 2096 na primeira arte), Feed (1080 × 1350) e Story (1080 × 1920) conferidos pelas dimensões reais do Canvas.
- Substituição de foto por um PNG válido; envio concluído e imagem exibida.
- Legenda editada e copiada; conteúdo da área de transferência conferido.
- Tentativa de trocar para Modelos com alterações não salvas: confirmação exibida e cancelamento mantém a edição.
- Upload GIF rejeitado antes do envio.
- Falha de upload simulada com resposta 500: criação e texto mantidos.
- Falha de persistência simulada com resposta 500: texto `PRESERVADO APÓS ERRO` permaneceu no editor.
- Acesso sem autenticação à rota de Marketing redirecionado para login.

No tablet e celular a prévia foi colocada antes dos campos, evitando que o usuário precise atravessar todo o formulário para encontrar a arte.

## Evidências temporárias inspecionadas

- `/tmp/finish-editor-1440.png`: editor com texto alterado e indicação de seleção.
- `/tmp/finish-editor-390.png`: formulário do editor no celular, sem corte horizontal.
- `/tmp/finish-photo.png`: foto substituída na área configurada do primeiro modelo.
- `/tmp/finish-mobile-preview.png`: prévia no celular após falha de persistência, com texto preservado.
- Galeria capturada em `/tmp/finish-gallery-{1440,1024,390}.png`.

Os arquivos de teste e capturas ficam fora do repositório. A inspeção visual não substitui os testes de backend, testes unitários e checks do frontend.

## Pontos encontrados que exigem conferência após correção

1. Na primeira arte, “Reposição de óleos Ipiranga”, havia texto visível adicional (“Faça a t…”) ausente da lista de elementos. Textos rasterizados fora das áreas cadastradas não podem ser substituídos pelo editor; o catálogo precisa cobrir os textos que o usuário deseja editar.
2. A área de foto da primeira arte era um retângulo na parte direita. A substituição preservava boa parte do frasco antigo fora desse retângulo. A área precisa representar o conteúdo que se deseja substituir e ser conferida por arte.
3. O renderizador restaurava textos originais sobre fotos substituídas copiando um retângulo completo da imagem original. Quando as áreas se cruzam, isso também copia o fundo antigo; o resultado deve ser conferido após o tratamento de transparência.
4. Uma resposta 500 no envio de PNG válido dizia “Use PNG ou JPEG de até 5 MB”, e podia aparecer simultaneamente em dois alertas. O formato estava correto; a mensagem deveria explicar falha de envio.

A imagem original sem alterações e uma reconstrução após trocar texto/foto são casos distintos. A fidelidade de todas as 54 artes modificadas não é atestada por esta inspeção de uma arte. Conferência individual das áreas e testes de exportação permanecem necessários antes de declarar essa parte completa.

## Fluxo simplificado de criação — 12/10/2026

Fluxo aprovado: Modelo → Personalizar → Revisar. Validação Chromium em 1440 e 390 pixels: sem rolagem horizontal; voltar preserva título. Testes de ponta a ponta cobrem navegação, salvamento/reabertura/duplicação, exportação local, upload inválido e falha de persistência em desktop, tablet e celular. Ajustes de texto, posição, identidade e cores começam recolhidos. Legenda, salvamento e download aparecem na revisão.

Resultado final do fluxo: 12/12 testes de ponta a ponta passaram na versão local de produção (http://127.0.0.1:3000), nas três resoluções configuradas; 51/51 testes unitários passaram, typecheck e build passaram. Prévia conferida em Chromium em 1440 e 390 pixels.

## Personalização compacta com prévia

Validação Chromium em 1440 e 390 pixels: texto e foto visíveis, detalhes recolhidos, edição de texto altera os pixels da prévia sem chamadas à API; campos adicionais preservados em Mais opções. Teste de foto válida pelo controle simplificado e conferência da prévia durante rolagem.

Resultado: 15/15 testes de ponta a ponta passaram na versão de produção local, nas três resoluções; typecheck e build passaram. Foto válida enviada pelo campo simplificado e conferida na prévia. Prévia móvel posicionada abaixo do cabeçalho fixo.
