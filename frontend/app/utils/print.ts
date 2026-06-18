import type { IReceipt, IStockItem } from "../../server/contracts/types";
import { formatCpf, formatCurrency, formatDateTime } from "./format";

const PRINT_STYLES = `
  @page {
    size: A4;
    margin: 14mm;
  }

  * {
    box-sizing: border-box;
  }

  body {
    margin: 0;
    color: #172033;
    font-family: Arial, Helvetica, sans-serif;
    font-size: 12px;
    line-height: 1.45;
  }

  .document {
    display: grid;
    gap: 18px;
  }

  .header {
    display: flex;
    justify-content: space-between;
    gap: 24px;
    padding-bottom: 14px;
    border-bottom: 2px solid #172033;
  }

  .brand {
    display: grid;
    gap: 4px;
  }

  .brand strong {
    font-size: 20px;
    letter-spacing: 0;
  }

  .muted {
    color: #5f6b7a;
  }

  h1,
  h2,
  p {
    margin: 0;
  }

  h1 {
    font-size: 18px;
    text-align: right;
  }

  h2 {
    font-size: 13px;
    margin-bottom: 8px;
    text-transform: uppercase;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }

  .box {
    padding: 12px;
    border: 1px solid #d8dee8;
    border-radius: 6px;
  }

  .total {
    display: flex;
    justify-content: space-between;
    padding: 14px;
    border: 2px solid #172033;
    border-radius: 6px;
    font-size: 16px;
    font-weight: 700;
  }

  table {
    width: 100%;
    border-collapse: collapse;
  }

  th,
  td {
    padding: 8px;
    border-bottom: 1px solid #d8dee8;
    text-align: left;
    vertical-align: top;
  }

  th {
    background: #f3f6fa;
    font-size: 11px;
    text-transform: uppercase;
  }

  .right {
    text-align: right;
  }

  .footer {
    display: grid;
    gap: 8px;
    padding-top: 24px;
    color: #5f6b7a;
    font-size: 11px;
  }
`;

export function printReceiptDocument(receipt: IReceipt) {
  const items = Array.isArray(receipt.items) ? receipt.items : [];
  const installments =
    receipt.paymentMethod === "credit_card" ? receipt.installments || 1 : 1;
  const installmentValueCents = Math.ceil(receipt.priceCents / installments);
  const rows = items.length
    ? items
        .map((item) => {
          const name = escapeHtml(item.stockItem?.name || item.stockItemId);
          return `
            <tr>
              <td>${name}</td>
              <td class="right">${item.quantity}</td>
            </tr>
          `;
        })
        .join("")
    : '<tr><td colspan="2">Nenhum produto vinculado.</td></tr>';

  openPrintDocument(
    `Recibo EMPI Autocenter`,
    `
      <main class="document">
        <header class="header">
          <div class="brand">
            <strong>EMPI Autocenter</strong>
            <span class="muted">Recibo de serviço</span>
          </div>
          <div>
            <h1>Recibo</h1>
            <p class="muted">Emissão: ${formatDate(receipt.createdAt)}</p>
            <p class="muted">Status: ${statusLabel(receipt.status)}</p>
            <p class="muted">Pagamento: ${paymentMethodLabel(receipt)}</p>
          </div>
        </header>

        <section class="grid">
          <article class="box">
            <h2>Cliente</h2>
            <p><strong>${escapeHtml(receipt.user.name)}</strong></p>
            <p>CPF: ${formatCpf(receipt.user.cpf || "")}</p>
            <p>Telefone: ${escapeHtml(receipt.user.phone || "-")}</p>
            <p>E-mail: ${escapeHtml(receipt.user.email || "-")}</p>
          </article>

          <article class="box">
            <h2>Veículo</h2>
            <p><strong>${escapeHtml(receipt.vehicleModel)} ${receipt.vehicleYear}</strong></p>
            <p>Placa: ${escapeHtml(receipt.vehiclePlate)}</p>
          </article>
        </section>

        <section class="box">
          <h2>Serviços</h2>
          <p>${escapeHtml(receipt.services)}</p>
          ${receipt.notes ? `<p class="muted">Observações: ${escapeHtml(receipt.notes)}</p>` : ""}
        </section>

        <section>
          <h2>Produtos utilizados</h2>
          <table>
            <thead>
              <tr>
                <th>Produto</th>
                <th class="right">Qtd.</th>
              </tr>
            </thead>
            <tbody>${rows}</tbody>
          </table>
        </section>

        ${
          receipt.paymentMethod === "credit_card"
            ? `
              <section class="box">
                <h2>Parcelamento</h2>
                <p>${installments}x de ${formatCurrency(installmentValueCents)}</p>
              </section>
            `
            : ""
        }

        <section class="total">
          <span>Valor total</span>
          <span>${formatCurrency(receipt.priceCents)}</span>
        </section>

        <footer class="footer">
          <p>Documento gerado pela EMPI Autocenter.</p>
          <p><strong>Este recibo não é uma nota fiscal.</strong></p>
        </footer>
      </main>
    `,
  );
}

export function printStockReport(items: IStockItem[]) {
  const rows = items.length
    ? items
        .map((item) => {
          return `
            <tr>
              <td>${escapeHtml(item.name)}</td>
              <td>${escapeHtml(item.description || "-")}</td>
              <td class="right">${formatCurrency(item.costCents)}</td>
              <td class="right">${formatCurrency(item.resalePriceCents)}</td>
              <td class="right">${item.quantity}</td>
              <td class="right">${item.usedQuantity}</td>
              <td>${formatDateTime(item.createdAt)}</td>
            </tr>
          `;
        })
        .join("")
    : '<tr><td colspan="7">Nenhum produto cadastrado.</td></tr>';

  openPrintDocument(
    "Relatório de estoque",
    `
      <main class="document">
        <header class="header">
          <div class="brand">
            <strong>EMPI Autocenter</strong>
            <span class="muted">Controle de estoque</span>
          </div>
          <div>
            <h1>Relatório de estoque</h1>
            <p class="muted">Emissão: ${formatDate(new Date().toISOString())}</p>
            <p class="muted">Produtos listados: ${items.length}</p>
          </div>
        </header>

        <section>
          <table>
            <thead>
              <tr>
                <th>Produto</th>
                <th>Descrição</th>
                <th class="right">Custo</th>
                <th class="right">Revenda</th>
                <th class="right">Qtd.</th>
                <th class="right">Usados</th>
                <th>Cadastrado em</th>
              </tr>
            </thead>
            <tbody>${rows}</tbody>
          </table>
        </section>
      </main>
    `,
  );
}

function openPrintDocument(title: string, body: string) {
  const popup = window.open("", "_blank");
  if (!popup) return;

  popup.opener = null;
  popup.document.write(`
    <!doctype html>
    <html lang="pt-BR">
      <head>
        <meta charset="utf-8" />
        <title>${escapeHtml(title)}</title>
        <style>${PRINT_STYLES}</style>
      </head>
      <body>
        ${body}
        <script>
          window.addEventListener('load', function () {
            window.focus();
            window.print();
          });
        </script>
      </body>
    </html>
  `);
  popup.document.close();
}

function escapeHtml(value: string) {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function statusLabel(status: IReceipt["status"]) {
  if (status === "paid") return "Pago";
  if (status === "cancelled") return "Cancelado";
  return "Pendente";
}

function paymentMethodLabel(receipt: IReceipt) {
  if (receipt.paymentMethod === "credit_card")
    return `Cartão de crédito (${receipt.installments || 1}x)`;
  if (receipt.paymentMethod === "debit_card") return "Cartão de débito";
  if (receipt.paymentMethod === "pix") return "Pix";
  return "Dinheiro";
}

function formatDate(value: string) {
  if (!value) return "-";

  return new Intl.DateTimeFormat("pt-BR", {
    dateStyle: "short",
    timeStyle: "short",
  }).format(new Date(value));
}
