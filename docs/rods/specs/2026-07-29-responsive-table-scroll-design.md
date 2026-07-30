# Rolagem horizontal responsiva de tabelas

## Decisão aprovada

Aplicar o comportamento de rolagem horizontal no wrapper global `.table-wrap`, usado por Estoque, Clientes, Recuperação e tabelas de Métricas.

## Comportamento

- Em telas estreitas, o painel não ultrapassa a largura disponível.
- A tabela preserva a distribuição normal de colunas no desktop e pode ser arrastada horizontalmente, inclusive por toque, quando exceder a tela.
- Recibos e Gastos mantêm o layout em cartões no celular, pois já expõem todos os campos sem ocultar colunas.

## Aceite

Em 390px de largura, tabelas largas não cortam colunas nem alargam a página; o usuário consegue deslizar o conteúdo lateralmente. Paginação e filtros sempre consultam dados novos na API.
