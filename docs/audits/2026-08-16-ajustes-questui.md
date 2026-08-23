# Ajustes QuestUI e modernização — 16/08/2026

## Resultado

A interface foi consolidada no QuestUI sem perder a linguagem operacional de ERP. A aplicação agora usa fundos marrons em camadas, superfícies ornamentadas, ouro como ação principal, vermelho para erro/perigo, roxo apenas em destaques especiais e texto em tom de pergaminho.

## Ajustes aplicados e motivos

### Fundamentos visuais

- Tokens globais substituídos pelos valores QuestUI e aliases legados mantidos para migrar a aplicação sem quebrar componentes existentes.
- Cinzel aplicada a títulos e rótulos de interface, Spectral ao corpo e Fira Code a conteúdo monoespaçado.
- Escala de espaçamento baseada em 8 px, raios angulares, bordas quentes e elevação com brilho dourado.
- Estados de botão, input, card, chip, lista, foco, erro e desabilitado padronizados.

Isso remove divergências entre páginas e cria uma base reutilizável para novos módulos.

### Layout e navegação

- Sidebar persistente em desktop, compacta em tablet e drawer controlável em mobile.
- Botão móvel com ícones e nomes acessíveis; o conteúdo principal deixa de ficar comprimido ou oculto.
- Cabeçalhos, cards, métricas, gráficos e autenticação receberam hierarquia visual coerente e acentos dourados discretos.
- Tabelas preservam densidade de ERP, mas oferecem regiões roláveis focáveis e rotuladas.

Isso melhora orientação, uso por teclado e leitura em telas menores sem transformar o produto em uma interface genérica.

### Usabilidade e feedback

- Mensagens críticas usam regiões de alerta.
- O formulário de gasto explica que salvar gera uma saída e exige caixa diário aberto.
- Datas e horários têm formatação determinística em `America/Sao_Paulo`, eliminando mudanças visuais durante a hidratação.
- Estados de foco visível e contraste foram reforçados segundo o alvo WCAG 2.2 AA.
- Animações foram limitadas a transições sutis, respeitando `prefers-reduced-motion`.

### Fluxos funcionais integrados

- Caixa diário integrado a gastos e movimentações financeiras.
- Criação de gasto permanece atômica com o lançamento de saída: falhas não deixam registros parciais.
- Jornada de Recibo rápido validada de ponta a ponta.
- Cadastro de estoque validado nos três tamanhos de tela.
- Autenticação, logout e proteção de rotas retestados.

### Automação adicionada

- Configuração Playwright com projetos desktop, tablet e mobile.
- Suite crítica cobrindo autenticação, todos os módulos, acessibilidade, console, HTTP 5xx, estoque, recibo, gasto, caixa e logout.
- axe-core integrado para bloquear violações sérias ou críticas.
- Artefatos de falha (screenshot, vídeo e trace) configurados para facilitar diagnóstico futuro.

## Regras permanentes

O QuestUI foi registrado em `spec/main/rules.md` como convenção durável. A especificação da modernização está em `docs/rods/specs/2026-08-16-questui-modernizacao-completa-design.md` e deve orientar novas telas, componentes e revisões.

## Próximos passos recomendados

- Executar a suíte E2E em CI com banco efêmero por job.
- Acrescentar testes de fechamento e reabertura de caixa, pagamento de parcelas e edição/arquivamento de gasto.
- Fazer uma rodada manual com NVDA/VoiceOver e zoom de 200%.
- Validar impressão/PDF em impressoras e dispositivos usados na oficina.
