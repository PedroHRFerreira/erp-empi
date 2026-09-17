import type { IReceipt, IUser } from "../../server/contracts/types";
import { receiptClientName } from "./receiptDisplay";
import { buildReceiptDocument } from "./receiptDocument";

type PdfText = {
  kind: "text";
  text: string;
  x: number;
  y: number;
  size?: number;
  font?: "F1" | "F2";
  color?: "black" | "white";
};

type PdfRule = {
  kind: "rule";
  x1: number;
  x2: number;
  y: number;
};

type PdfBox = {
  kind: "box";
  x: number;
  y: number;
  width: number;
  height: number;
  fill?: "navy" | "gray";
};

type PdfElement = PdfText | PdfRule | PdfBox;

const PAGE_WIDTH = 595.28;
const PAGE_HEIGHT = 841.89;
const LEFT = 62;
const RIGHT = PAGE_WIDTH - 62;
const CONTENT_WIDTH = RIGHT - LEFT;

export function receiptWhatsAppMessage(
  receipt: IReceipt,
  company: IUser | null = null,
) {
  const document = buildReceiptDocument(receipt, company);
  const services = document.lines.filter((line) => line.kind === "service");
  const products = document.lines.filter((line) => line.kind === "product");
  const total = document.summaryRows.find((row) => row.label === "Total");
  const statusTotal =
    receipt.status === "paid" ? "TOTAL PAGO" : "TOTAL A PAGAR";
  const lineText = (line: (typeof document.lines)[number]) =>
    `- ${line.quantity}x ${line.description}: ${line.totalLabel}`;

  return [
    `Olá, ${document.customer.name}!`,
    "",
    `Segue o seu recibo da *${document.company.name}*.`,
    `*${document.receiptNumber}* - ${document.issuedAtLabel}`,
    "",
    `*Veículo:* ${document.vehicle.name}`,
    document.vehicle.lines.join(" | "),
    "",
    "*SERVIÇOS*",
    ...services.map(lineText),
    ...(products.length
      ? ["", "*PEÇAS E MATERIAIS*", ...products.map(lineText)]
      : []),
    "",
    `*Forma de pagamento:* ${document.payment.methodLabel}`,
    `*Status:* ${document.payment.statusLabel}`,
    `*${statusTotal}: ${total?.valueLabel || document.payment.amountLabel}*`,
    ...(document.notes ? ["", `*Observações:* ${document.notes}`] : []),
    "",
    "Obrigado pela confiança! O recibo em PDF segue em anexo.",
  ].join("\n");
}

export async function shareReceiptPdf(
  receipt: IReceipt,
  company: IUser | null = null,
) {
  const file = buildReceiptPdfFile(receipt, company);
  const text = receiptWhatsAppMessage(receipt, company);
  const document = buildReceiptDocument(receipt, company);
  const shareData = {
    title: `${document.receiptNumber} - ${receiptClientName(receipt)}`,
    text,
    files: [file],
  };

  if (navigator.canShare?.(shareData)) {
    try {
      await navigator.share(shareData);
      return true;
    } catch {
      downloadReceiptPdf(file);
      return false;
    }
  }

  downloadReceiptPdf(file);
  return false;
}

export function buildReceiptPdfFile(
  receipt: IReceipt,
  company: IUser | null = null,
) {
  const document = buildReceiptDocument(receipt, company);
  const bytes = buildReceiptPdfBytes(receipt, company);
  const filename = `${document.receiptNumber.toLowerCase().replace(/\s+/g, "-")}.pdf`;
  return new File([bytes], filename, { type: "application/pdf" });
}

export function buildReceiptPdfBytes(
  receipt: IReceipt,
  company: IUser | null = null,
) {
  const document = buildReceiptDocument(receipt, company);
  const selectedPages = layoutReceipt(document, 10);
  return createPdf(
    selectedPages.map((page) =>
      page.map((element) => drawElement(element)).join("\n"),
    ),
  );
}

function layoutReceipt(
  document: ReturnType<typeof buildReceiptDocument>,
  fontSize: number,
): PdfElement[][] {
  const elements: PdfElement[] = [];
  let y = PAGE_HEIGHT - 62;
  const lineHeight = fontSize + 4;
  const text = (
    value: string,
    x = LEFT,
    width = CONTENT_WIDTH,
    font: "F1" | "F2" = "F1",
    size = fontSize,
  ) => {
    for (const part of wrapText(value || "-", maxCharacters(width, size))) {
      addLine(elements, part, x, y, size, font);
      y -= size + 4;
    }
  };
  const centered = (value: string, size: number, font: "F1" | "F2" = "F2") => {
    addLine(elements, value, centeredX(value, size), y, size, font);
    y -= size + 5;
  };
  const bar = (title: string) => {
    y -= 3;
    elements.push({
      kind: "box",
      x: LEFT,
      y: y - 21,
      width: CONTENT_WIDTH,
      height: 21,
      fill: "navy",
    });
    addLine(elements, title, centeredX(title, 12), y - 15, 12, "F2", "white");
    y -= 38;
  };
  const sectionHeading = (title: string) => {
    y -= 14;
    const textWidth = estimatedTextWidth(title, 13, true);
    const titleWidth = Math.max(164, textWidth + 24);
    const titleX = (PAGE_WIDTH - titleWidth) / 2;
    addRule(elements, y + 3, LEFT, titleX - 14);
    elements.push({
      kind: "box",
      x: titleX,
      y: y - 6,
      width: titleWidth,
      height: 20,
      fill: "gray",
    });
    addLine(elements, title, (PAGE_WIDTH - textWidth) / 2, y, 13, "F2");
    addRule(elements, y + 3, titleX + titleWidth + 14, RIGHT);
    y -= 34;
  };
  const itemSection = (title: string, lines: typeof document.lines) => {
    if (!lines.length) return;
    sectionHeading(title);
    addLine(elements, title.includes("SERVIÇOS") ? "SERVIÇO" : "DESCRIÇÃO", LEFT, y, fontSize, "F2");
    addLine(elements, "VALOR (R$)", RIGHT - 78, y, fontSize, "F2");
    y -= 10;
    addRule(elements, y, LEFT, RIGHT);
    y -= 16;
    for (const line of lines) {
      const parts = wrapText(
        line.quantity === "1"
          ? line.description
          : `${line.quantity}x ${line.description}`,
        maxCharacters(340, fontSize),
      );
      parts.forEach((part, index) => {
        addLine(elements, part, LEFT, y, fontSize);
        if (index === 0)
          addLine(
            elements,
            line.totalLabel.replace("R$ ", ""),
            RIGHT - 64,
            y,
            fontSize,
          );
        y -= lineHeight;
      });
      y -= 3;
      addRule(elements, y, LEFT, RIGHT);
      y -= 15;
    }
    const subtotal = lines.reduce((sum, line) => sum + line.totalCents, 0);
    addLine(
      elements,
      `Subtotal ${title.includes("SERVIÇOS") ? "Serviços" : "Peças"}:`,
      LEFT,
      y,
      fontSize,
      "F2",
    );
    addLine(elements, formatCents(subtotal), RIGHT - 64, y, fontSize, "F2");
    y -= 20;
  };

  centered(document.company.name.toUpperCase(), 31);
  centered("-  Serviços Automotivos  -", 16);
  addRule(elements, y, LEFT, RIGHT);
  y -= 23;
  addLine(elements, document.receiptNumber.replace("Recibo ", "Recibo Nº: "), LEFT, y, 14, "F2");
  addLine(
    elements,
    `Data: ${document.issuedAtLabel}  |  Governador Valadares - MG`,
    RIGHT - 245,
    y,
    10,
    "F2",
  );
  y -= 20;
  addRule(elements, y, LEFT, RIGHT);
  y -= 16;
  text(document.company.lines.join("  |  "), LEFT, CONTENT_WIDTH, "F1", 8);

  bar("DADOS DO CLIENTE");
  text(`Cliente: ${document.customer.name}`, LEFT, CONTENT_WIDTH, "F2", 14);
  text(
    document.customer.lines.join("  |  ") || "-",
    LEFT,
    CONTENT_WIDTH,
    "F1",
    10,
  );
  bar("DADOS DO VEÍCULO");
  text(
    `Veículo: ${document.vehicle.name.replace(/\s+\d{4}$/, "")}  |  ${document.vehicle.lines.join("  |  ")}`,
    LEFT,
    CONTENT_WIDTH,
    "F1",
    12,
  );

  itemSection(
    "DESCRIÇÃO DOS SERVIÇOS",
    document.lines.filter((line) => line.kind === "service"),
  );
  itemSection(
    "PEÇAS E MATERIAIS",
    document.lines.filter((line) => line.kind === "product"),
  );

  const detailRows = document.summaryRows.filter((row) => !row.strong);
  if (detailRows.some((row) => row.label === "Desconto")) {
    const netTotal = document.summaryRows.find((row) => row.strong)?.valueCents || 0;
    const discount = Math.abs(detailRows.find((row) => row.label === "Desconto")?.valueCents || 0);
    const grossTotal = netTotal + discount;
    addLine(elements, "TOTAL:", LEFT, y, 12);
    addLine(elements, `R$ ${formatCents(grossTotal)}`, RIGHT - 90, y, 12);
    y -= 18;
  }
  for (const row of detailRows) {
    addLine(elements, `${row.label.toUpperCase()}:`, LEFT, y, 12);
    addLine(elements, row.valueLabel, RIGHT - 90, y, 12);
    y -= 18;
  }
  const total = document.summaryRows.find((row) => row.strong);
  y -= 7;
  elements.push({
    kind: "box",
    x: LEFT,
    y: y - 7,
    width: CONTENT_WIDTH,
    height: 30,
    fill: "navy",
  });
  addLine(
    elements,
    document.payment.statusLabel === "Pago" ? "TOTAL PAGO:" : "TOTAL A PAGAR:",
    LEFT + 8,
    y,
    16,
    "F2",
    "white",
  );
  addLine(
    elements,
    total?.valueLabel || document.payment.amountLabel,
    RIGHT - 110,
    y,
    18,
    "F2",
    "white",
  );
  y -= 37;
  text(
    `Forma de pagamento: ${document.payment.methodLabel} (${document.payment.dateLabel}) - ${document.payment.statusLabel}`,
    LEFT,
    CONTENT_WIDTH,
    "F2",
    10,
  );
  if (document.notes)
    text(`Observações: ${document.notes}`, LEFT, CONTENT_WIDTH, "F1", 10);
  y -= 26;
  addRule(elements, y, LEFT, LEFT + 200);
  addRule(elements, y, RIGHT - 200, RIGHT);
  y -= 15;
  addLine(elements, "Assinatura do Cliente", LEFT + 48, y, 9);
  addLine(
    elements,
    `Assinatura / Carimbo - ${document.company.name}`,
    RIGHT - 190,
    y,
    9,
  );
  addLine(
    elements,
    `Obrigado pela confiança! ${document.company.name} - Atendimento de segunda a sexta 08:00-18:00 | sábado 08:00-12:00`,
    LEFT,
    44,
    8,
    "F2",
  );
  addLine(elements, document.legalNotice, LEFT, 30, 7);
  return [elements];
}

function centeredX(value: string, size: number) {
  return Math.max(LEFT, (PAGE_WIDTH - estimatedTextWidth(value, size)) / 2);
}

function estimatedTextWidth(value: string, size: number, bold = false) {
  return value.length * size * (bold ? 0.56 : 0.52);
}

function formatCents(value: number) {
  return (value / 100).toLocaleString("pt-BR", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

function maxCharacters(width: number, size: number) {
  return Math.max(8, Math.floor(width / Math.max(size * 0.52, 3.2)));
}

function wrapText(value: string, maxLength: number) {
  const words = String(value || "-")
    .trim()
    .split(/\s+/);
  const lines: string[] = [];
  let line = "";
  for (const word of words) {
    if (!line) {
      line = word;
      continue;
    }
    if (`${line} ${word}`.length <= maxLength) {
      line += ` ${word}`;
      continue;
    }
    lines.push(line);
    line = word;
  }
  if (line) lines.push(line);
  return lines;
}

function addLine(
  elements: PdfElement[],
  text: string,
  x: number,
  y: number,
  size = 10,
  font: "F1" | "F2" = "F1",
  color: "black" | "white" = "black",
) {
  elements.push({ kind: "text", text, x, y, size, font, color });
}

function addRule(elements: PdfElement[], y: number, x1 = LEFT, x2 = RIGHT) {
  elements.push({ kind: "rule", x1, x2, y });
}

function drawElement(element: PdfElement) {
  if (element.kind === "box") {
    const fill = element.fill === "navy" ? "0.04 0.13 0.25 rg" : "0.93 g";
    return `${fill} ${element.x} ${element.y} ${element.width} ${element.height} re f 0 g 0 G`;
  }
  if (element.kind === "rule")
    return `0.6 w ${element.x1} ${element.y} m ${element.x2} ${element.y} l S`;
  return drawText(element);
}

function drawText(line: PdfText) {
  const color = line.color === "white" ? "1 g " : "0 g ";
  return `${color}BT /${line.font || "F1"} ${line.size || 10} Tf 1 0 0 1 ${line.x} ${line.y} Tm (${escapePdfString(line.text)}) Tj ET 0 g`;
}

function createPdf(contentStreams: string[]) {
  const pageObjectStart = 3;
  const fontRegularObject = pageObjectStart + contentStreams.length * 2;
  const fontBoldObject = fontRegularObject + 1;
  const objects = [
    "<< /Type /Catalog /Pages 2 0 R >>",
    `<< /Type /Pages /Kids [${contentStreams.map((_, index) => `${pageObjectStart + index * 2} 0 R`).join(" ")}] /Count ${contentStreams.length} >>`,
    ...contentStreams.flatMap((contentStream, index) => {
      const pageObject = `<< /Type /Page /Parent 2 0 R /MediaBox [0 0 ${PAGE_WIDTH} ${PAGE_HEIGHT}] /Resources << /Font << /F1 ${fontRegularObject} 0 R /F2 ${fontBoldObject} 0 R >> >> /Contents ${pageObjectStart + index * 2 + 1} 0 R >>`;
      const contentObject = `<< /Length ${latin1Length(contentStream)} >>\nstream\n${contentStream}\nendstream`;
      return [pageObject, contentObject];
    }),
    "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>",
    "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>",
  ];

  const header = "%PDF-1.4\n";
  const chunks: string[] = [header];
  const offsets: number[] = [0];
  let length = latin1Length(header);

  objects.forEach((object, index) => {
    offsets.push(length);
    const chunk = `${index + 1} 0 obj\n${object}\nendobj\n`;
    chunks.push(chunk);
    length += latin1Length(chunk);
  });

  const xrefOffset = length;
  const xref = [
    `xref\n0 ${objects.length + 1}\n`,
    "0000000000 65535 f \n",
    ...offsets
      .slice(1)
      .map((offset) => `${String(offset).padStart(10, "0")} 00000 n \n`),
    `trailer\n<< /Size ${objects.length + 1} /Root 1 0 R >>\nstartxref\n${xrefOffset}\n%%EOF`,
  ].join("");

  chunks.push(xref);
  return latin1Bytes(chunks.join(""));
}

function downloadReceiptPdf(file: File) {
  const href = URL.createObjectURL(file);
  const link = document.createElement("a");
  link.href = href;
  link.download = file.name;
  link.click();
  URL.revokeObjectURL(href);
}

function escapePdfString(value: string) {
  return value
    .replaceAll("\\", "\\\\")
    .replaceAll("(", "\\(")
    .replaceAll(")", "\\)");
}

function latin1Length(value: string) {
  return latin1Bytes(value).length;
}

function latin1Bytes(value: string) {
  const bytes = new Uint8Array(value.length);
  for (let index = 0; index < value.length; index += 1) {
    const code = value.charCodeAt(index);
    bytes[index] = code <= 255 ? code : 63;
  }
  return bytes;
}
