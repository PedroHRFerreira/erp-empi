# Recibos e saldos do caixa

## Problemas

- Um único envio de recibo rápido ou normal cria dois registros de recibo.
- A tela Caixa separa corretamente o dinheiro físico, mas não apresenta com clareza o saldo consolidado e o resultado diário.

## Comportamento aprovado

- Cada envio de recibo rápido ou normal deve criar exatamente um registro.
- **Dinheiro esperado na gaveta** continua representando apenas o dinheiro físico: fundo inicial mais entradas em dinheiro menos saídas em dinheiro.
- **Saldo total disponível** consolida o fundo inicial e todos os meios de pagamento: `fundo inicial + todas as entradas - todas as saídas`.
- **Resultado do dia** considera apenas o movimento do dia: `entradas do dia - saídas do dia`. Valor positivo é lucro; valor negativo é prejuízo.
- A interface deve explicar as duas fórmulas e informar que o fundo inicial compõe o saldo disponível, mas não o lucro ou prejuízo do dia.

## Exemplo de aceitação

Com fundo inicial de R$ 1.271,38, entrada em dinheiro de R$ 45,00 e saída em PIX de R$ 129,95:

- dinheiro esperado na gaveta: R$ 1.316,38;
- saldo total disponível: R$ 1.186,43;
- resultado do dia: -R$ 84,95 (prejuízo).

## Critérios de aceitação

1. Salvar uma vez qualquer modalidade de recibo gera somente um registro no banco.
2. Reenvios involuntários durante o processamento não geram registros adicionais.
3. Os três valores do exemplo são apresentados corretamente na tela Caixa.
4. As fórmulas ficam visíveis em texto explicativo e compreensível.
5. Testes cobrem os cálculos e a prevenção da duplicação.
