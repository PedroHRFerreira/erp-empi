import type { IReceipt, IStockItem, IUser } from "../../server/contracts/types";
import { formatCurrency, formatDateTime } from "./format";
import {
  buildReceiptDocument,
  buildReceiptInvoiceData,
  type IReceiptDocumentLine,
  type IReceiptDocumentMoneyRow,
  type IReceiptDocumentParty,
  type IReceiptInvoicePortalRow,
} from "./receiptDocument";

export const NFSE_PORTAL_URL = "https://www.nfse.gov.br/EmissorNacional/Login?ReturnUrl=%2fEmissorNacional";

const PRINT_STYLES = `
  @page {
    size: A4;
    margin: 0;
  }

  * {
    box-sizing: border-box;
  }

  body {
    margin: 0;
    color: #171717;
    background: #f1f4f8;
    font-family: Arial, Helvetica, sans-serif;
    font-size: 12px;
    line-height: 1.45;
  }

  .document {
    display: grid;
    gap: 18px;
    max-width: 780px;
    margin: 0 auto;
    padding: 36px;
    background: #ffffff;
    box-shadow: 0 18px 42px rgba(14, 23, 38, 0.18);
  }

  .document--receipt {
    gap: 24px;
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

  .receipt-header {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 28px;
    align-items: start;
  }

  .company {
    display: grid;
    gap: 18px;
    justify-items: start;
  }

  .company__badge {
    display: grid;
    width: 72px;
    height: 72px;
    place-items: center;
    color: #9b5a63;
    background: #fbf1f2;
    border-radius: 4px;
    font-size: 18px;
    font-weight: 800;
  }

  .company__details {
    display: grid;
    gap: 3px;
  }

  .company__details strong {
    margin-bottom: 4px;
    font-size: 15px;
  }

  .document-title {
    display: grid;
    gap: 8px;
    justify-items: end;
    text-align: right;
  }

  .document-title h1 {
    font-size: 30px;
    line-height: 1;
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

  .section-title {
    margin-bottom: 10px;
    font-size: 16px;
    text-transform: none;
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

  .divider {
    height: 1px;
    background: #e7eaf0;
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

  .receipt-items th {
    color: #171717;
    background: transparent;
    border-bottom: 2px solid #7b7b7b;
    font-size: 15px;
    text-transform: none;
  }

  .receipt-items td {
    padding-top: 14px;
    padding-bottom: 14px;
    border-bottom-color: #edf0f4;
  }

  .receipt-items tbody tr:last-child td {
    border-bottom: 2px solid #7b7b7b;
  }

  .right {
    text-align: right;
  }

  .summary {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(260px, 320px);
    gap: 24px;
  }

  .summary__box {
    display: grid;
    gap: 8px;
  }

  .money-row,
  .payment-row {
    display: flex;
    justify-content: space-between;
    gap: 18px;
    border-bottom: 1px solid #edf0f4;
    padding: 5px 0;
  }

  .money-row--strong {
    font-size: 16px;
    font-weight: 800;
  }

  .payment-details {
    display: grid;
    gap: 10px;
  }

  .payment-details__list {
    display: flex;
    flex-wrap: wrap;
    gap: 0;
  }

  .payment-details__list span {
    padding: 0 22px;
    border-left: 1px solid #e2e7ef;
  }

  .payment-details__list span:first-child {
    padding-left: 0;
    border-left: 0;
  }

  .thank-you {
    display: grid;
    gap: 4px;
  }

  .invoice-data {
    display: grid;
    gap: 18px;
  }

  .invoice-data__notice {
    padding: 12px;
    border: 1px solid #d8dee8;
    border-radius: 6px;
    color: #172033;
    background: #f8fafc;
    font-weight: 700;
  }

  .invoice-data__section {
    display: grid;
    gap: 8px;
  }

  .invoice-data__party {
    padding: 10px 0;
    border-bottom: 1px solid #edf0f4;
  }

  .invoice-data__party strong,
  .invoice-data__party span {
    display: block;
  }

  .footer {
    display: grid;
    gap: 8px;
    padding-top: 24px;
    color: #5f6b7a;
    font-size: 11px;
  }

  @media print {
    body {
      padding: 14mm;
      background: #ffffff;
    }

    .document {
      max-width: none;
      padding: 0;
      box-shadow: none;
    }
  }
`;

export function printReceiptDocument(receipt: IReceipt, company: IUser | null = null) {
  const document = buildReceiptDocument(receipt, company);

  openPrintDocument(
    document.receiptNumber,
    `
      <main class="document document--receipt">
        <header class="receipt-header">
          <div class="company">
            <span class="company__badge">${escapeHtml(document.company.initials)}</span>
            <div class="company__details">
              <strong>${escapeHtml(document.company.name)}</strong>
              ${renderTextLines(document.company.lines)}
            </div>
          </div>
          <div class="document-title">
            <h1>${escapeHtml(document.receiptNumber)}</h1>
            <p class="muted">${escapeHtml(document.issuedAtLabel)}</p>
          </div>
        </header>

        <div class="divider"></div>

        <section>
          <h2 class="section-title">Dados do cliente</h2>
          <p><strong>${escapeHtml(document.customer.name)}</strong></p>
          ${renderTextLines(document.customer.lines)}
          <p class="muted">${escapeHtml(document.vehicle.name)} - ${escapeHtml(document.vehicle.lines.join(" "))}</p>
        </section>

        <section>
          <table class="receipt-items">
            <thead>
              <tr>
                <th>Itens</th>
                <th class="right">Quantidade</th>
                <th class="right">Preço</th>
                <th class="right">Total da linha</th>
              </tr>
            </thead>
            <tbody>${renderReceiptLineRows(document.lines)}</tbody>
          </table>
        </section>

        <section class="summary">
          <span aria-hidden="true"></span>
          <div class="summary__box">
            ${renderSummaryRows(document.summaryRows)}
          </div>
        </section>

        <div class="divider"></div>

        <section class="payment-details">
          <h2 class="section-title">Detalhes do pagamento</h2>
          <div class="payment-details__list">
            <span>${escapeHtml(document.payment.dateLabel)}</span>
            <span>${escapeHtml(document.payment.methodLabel)}</span>
            <span>${escapeHtml(document.payment.amountLabel)}</span>
          </div>
        </section>

        <div class="divider"></div>

        <section class="thank-you">
          <h2 class="section-title">${escapeHtml(document.thankYouTitle)}</h2>
          <p>${escapeHtml(document.thankYouMessage)}</p>
        </section>

        <footer class="footer">
          <p><strong>${escapeHtml(document.legalNotice)}</strong></p>
        </footer>
      </main>
    `,
  );
}

export function printReceiptInvoiceData(receipt: IReceipt, company: IUser | null = null) {
  const document = buildReceiptInvoiceData(receipt, company);

  openPrintDocument(
    document.title,
    `
      <main class="document invoice-data">
        <header class="receipt-header">
          <div class="company">
            <span class="company__badge">${escapeHtml(document.receipt.company.initials)}</span>
            <div class="company__details">
              <strong>${escapeHtml(document.receipt.company.name)}</strong>
              ${renderTextLines(document.receipt.company.lines)}
            </div>
          </div>
          <div class="document-title">
            <h1>Dados para nota fiscal</h1>
            <p class="muted">${escapeHtml(document.receipt.receiptNumber)}</p>
          </div>
        </header>

        <p class="invoice-data__notice">${escapeHtml(document.notice)}</p>

        <section class="invoice-data__section">
          <h2 class="section-title">Preenchimento sugerido no portal NFS-e</h2>
          <table>
            <tbody>${renderPortalRows(document.portalRows)}</tbody>
          </table>
        </section>

        <section class="invoice-data__section">
          <h2 class="section-title">Prestador</h2>
          ${renderParties(document.providerRows)}
        </section>

        <section class="invoice-data__section">
          <h2 class="section-title">Tomador e veículo</h2>
          ${renderParties(document.customerRows)}
        </section>

        <section>
          <h2 class="section-title">Serviços, produtos e valores</h2>
          <table class="receipt-items">
            <thead>
              <tr>
                <th>Descrição</th>
                <th class="right">Quantidade</th>
                <th class="right">Valor unitário</th>
                <th class="right">Total</th>
              </tr>
            </thead>
            <tbody>${renderReceiptLineRows(document.serviceRows)}</tbody>
          </table>
        </section>

        <section class="summary">
          <span aria-hidden="true"></span>
          <div class="summary__box">
            ${renderSummaryRows(document.summaryRows)}
          </div>
        </section>

        <footer class="footer">
          <p><strong>${escapeHtml(document.notice)}</strong></p>
        </footer>
      </main>
    `,
  );
}

export function prepareReceiptInvoiceIssue(receipt: IReceipt, company: IUser | null = null) {
  const portal = window.open(NFSE_PORTAL_URL, "_blank");
  if (portal) {
    portal.opener = null;
  }

  printReceiptInvoiceData(receipt, company);

  const text = buildReceiptInvoiceClipboardText(receipt, company);
  if (!navigator.clipboard) return Promise.resolve(false);

  return navigator.clipboard
    .writeText(text)
    .then(() => true)
    .catch(() => false);
}

export function buildReceiptInvoiceClipboardText(receipt: IReceipt, company: IUser | null = null) {
  const document = buildReceiptInvoiceData(receipt, company);

  return [
    document.title,
    "",
    "Dados para preencher no portal:",
    ...document.portalRows.map((row) => `${row.label}: ${row.value}`),
    "",
    "Atenção:",
    document.notice,
  ].join("\n");
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

function renderReceiptLineRows(lines: IReceiptDocumentLine[]) {
  return lines
    .map((line) => {
      return `
        <tr>
          <td>${escapeHtml(line.description)}</td>
          <td class="right">${escapeHtml(line.quantity)}</td>
          <td class="right">${escapeHtml(line.priceLabel)}</td>
          <td class="right"><strong>${escapeHtml(line.totalLabel)}</strong></td>
        </tr>
      `;
    })
    .join("");
}

function renderSummaryRows(rows: IReceiptDocumentMoneyRow[]) {
  return rows
    .map((row) => {
      return `
        <div class="money-row ${row.strong ? "money-row--strong" : ""}">
          <span>${escapeHtml(row.label)}</span>
          <span>${escapeHtml(row.valueLabel)}</span>
        </div>
      `;
    })
    .join("");
}

function renderPortalRows(rows: IReceiptInvoicePortalRow[]) {
  return rows
    .map((row) => {
      return `
        <tr>
          <td><strong>${escapeHtml(row.label)}</strong></td>
          <td>${escapeHtml(row.value)}</td>
        </tr>
      `;
    })
    .join("");
}

function renderParties(parties: IReceiptDocumentParty[]) {
  return parties
    .map((party) => {
      return `
        <div class="invoice-data__party">
          <strong>${escapeHtml(party.name)}</strong>
          ${renderTextLines(party.lines)}
        </div>
      `;
    })
    .join("");
}

function renderTextLines(lines: string[]) {
  return lines.map((line) => `<p class="muted">${escapeHtml(line)}</p>`).join("");
}

function escapeHtml(value: string) {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function formatDate(value: string) {
  if (!value) return "-";

  return new Intl.DateTimeFormat("pt-BR", {
    dateStyle: "short",
    timeStyle: "short",
  }).format(new Date(value));
}
