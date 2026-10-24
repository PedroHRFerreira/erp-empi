# Recibo com duas formas de pagamento

Plano confirmado pelo usuário em 23/10/2026.

## Comportamento

- Recibos normais e rápidos aceitam uma ou duas formas distintas: dinheiro, PIX, débito ou crédito.
- As duas partes quitam o recibo juntas. Recebimento em datas diferentes está fora desta entrega.
- Os controles ficam em Finalizar recibo. O usuário informa o valor base da primeira parte; a segunda recebe o restante do subtotal após desconto.
- Taxas incidem somente sobre a base da parte do cartão, usando as configurações atuais e arredondamento em centavos. Crédito permite de 1 a 12 parcelas por parte.
- Exemplo: subtotal de R$ 100, R$ 60 PIX + R$ 40 cartão com 5% = R$ 60 PIX + R$ 42 cartão, total R$ 102.

## Contrato e persistência

- Nova tabela receipt_payments, ordenada e vinculada ao recibo, com paymentMethod, baseAmountCents, cardFeePercent, cardFeeCents, amountCents e installments.
- Criação e edição recebem payments com uma ou duas partes; servidor calcula taxas e totais e valida valores, formas distintas e soma das bases igual ao subtotal.
- priceCents e cardFeeCents do recibo agregam as partes. paymentMethod é mixed quando dividido; formas efetivas de caixa/despesas não aceitam mixed.
- Requisições antigas continuam aceitas. Recibos antigos são apresentados como uma parte com os valores históricos, sem recalcular taxas nem alterar lançamentos.
- Partes persistidas são editáveis apenas enquanto o recibo está pendente.

## Quitação e caixa

- Um lançamento por parte, na mesma transação da quitação e baixa de estoque.
- Apenas dinheiro compõe a gaveta e exige caixa aberto. PIX e cartões alimentam os próprios saldos.
- Uma falha desfaz a operação inteira. Repetir confirmação não duplica lançamentos nem estoque.
- Crédito parcelado registra o total da parte no saldo de crédito ao quitar. Métricas contam o recibo uma vez.
- Listagem, recuperação, histórico do cliente, resumo, impressão, PDF e WhatsApp apresentam a divisão.

## Aceitação

- Testar PIX + dinheiro, cartão + PIX, crédito + débito, descontos, taxas diferentes, parcelamento e centavos.
- Rejeitar soma incorreta, partes zeradas/negativas na divisão, formas repetidas, terceira parte e parcelas inválidas.
- Cobrir edição, legado, caixa fechado, falha na segunda entrada, confirmação repetida e uma única baixa de estoque.
- Validar testes Go, testes da interface, tipos e navegador em desktop e celular, incluindo documentos gerados.
