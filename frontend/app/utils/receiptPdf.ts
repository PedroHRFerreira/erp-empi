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
};

type PdfRule = {
  kind: "rule";
  x1: number;
  x2: number;
  y: number;
};

type PdfElement = PdfText | PdfRule;

const PAGE_WIDTH = 226.77;
const PAGE_HEIGHT = 510.24;
const LEFT = 12;
const RIGHT = PAGE_WIDTH - 12;
const CONTENT_WIDTH = RIGHT - LEFT;

export function receiptWhatsAppMessage(
  receipt: IReceipt,
  company: IUser | null = null,
) {
  const document = buildReceiptDocument(receipt, company);
  const itemLines = document.lines.map(
    (line) => `- ${line.quantity}x ${line.description}: ${line.totalLabel}`,
  );

  return [
    document.receiptNumber,
    document.company.name,
    ...document.company.lines,
    `Cliente: ${document.customer.name}`,
    `Veículo: ${document.vehicle.name}`,
    ...document.vehicle.lines,
    "Serviços:",
    ...itemLines,
    `Pagamento: ${document.payment.methodLabel}`,
    ...document.summaryRows.map((row) => `${row.label}: ${row.valueLabel}`),
    document.legalNotice,
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
  const pages = [9, 8, 7]
    .map((fontSize) => layoutReceipt(document, fontSize))
    .find((candidate) => candidate.length === 1);
  const selectedPages = pages || layoutReceipt(document, 7);
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
  const pages: PdfElement[][] = [];
  let elements: PdfElement[] = [];
  pages.push(elements);
  let y = PAGE_HEIGHT - 14;
  const lineHeight = fontSize + 2;
  const sectionTitleSize = fontSize + 2;
  const ensure = (height: number) => {
    if (y - height >= 14) return;
    elements = [];
    pages.push(elements);
    y = PAGE_HEIGHT - 14;
    addHeader(
      elements,
      document,
      fontSize,
      () => y,
      (value) => {
        y = value;
      },
    );
  };
  const text = (
    value: string,
    x = LEFT,
    width = CONTENT_WIDTH,
    font: "F1" | "F2" = "F1",
    size = fontSize,
  ) => {
    const lines = wrapText(value || "-", maxCharacters(width, size));
    ensure(lines.length * lineHeight);
    for (const part of lines) {
      addLine(elements, part, x, y, size, font);
      y -= lineHeight;
    }
  };
  const rule = () => {
    ensure(4);
    addRule(elements, y, LEFT, RIGHT);
    y -= 5;
  };
  const heading = (value: string) => {
    ensure(sectionTitleSize + 8);
    addLine(elements, value, LEFT, y, sectionTitleSize, "F2");
    y -= sectionTitleSize + 4;
    rule();
  };
  const section = (title: string, lines: typeof document.lines) => {
    if (!lines.length) return;
    heading(title);
    for (const line of lines) {
      ensure(lineHeight * 2 + 3);
      text(line.description, LEFT, 112);
      addLine(elements, `Qtd. ${line.quantity}`, 128, y + lineHeight, fontSize);
      addLine(elements, line.totalLabel, 176, y + lineHeight, fontSize, "F2");
      y -= 3;
    }
  };

  addHeader(
    elements,
    document,
    fontSize,
    () => y,
    (value) => {
      y = value;
    },
  );
  heading("Cliente e veículo");
  text(document.customer.name, LEFT, CONTENT_WIDTH, "F2");
  text(document.customer.lines.join(" | ") || "-", LEFT);
  text(`Veículo: ${document.vehicle.name}`, LEFT, CONTENT_WIDTH, "F2");
  text(document.vehicle.lines.join(" | "), LEFT);
  section(
    "Serviços",
    document.lines.filter((line) => line.kind !== "product"),
  );
  section(
    "Produtos",
    document.lines.filter((line) => line.kind === "product"),
  );
  rule();
  for (const row of document.summaryRows) {
    ensure(lineHeight);
    addLine(
      elements,
      row.label,
      118,
      y,
      row.strong ? fontSize + 2 : fontSize,
      row.strong ? "F2" : "F1",
    );
    addLine(
      elements,
      row.valueLabel,
      176,
      y,
      row.strong ? fontSize + 2 : fontSize,
      row.strong ? "F2" : "F1",
    );
    y -= row.strong ? lineHeight + 2 : lineHeight;
  }
  heading("Pagamento");
  text(
    `${document.payment.methodLabel} — ${document.payment.amountLabel}`,
    LEFT,
    CONTENT_WIDTH,
    "F2",
  );
  text(`Data: ${document.payment.dateLabel}`, LEFT);
  if (document.notes) {
    heading("Observações");
    text(document.notes);
  }
  heading(document.thankYouTitle);
  text(document.thankYouMessage);
  ensure(lineHeight * 4);
  addRule(elements, y - lineHeight, LEFT, 94);
  addRule(elements, y - lineHeight, 132, RIGHT);
  addLine(
    elements,
    document.company.name,
    LEFT,
    y - lineHeight * 2,
    fontSize - 1,
    "F2",
  );
  addLine(
    elements,
    document.customer.name,
    132,
    y - lineHeight * 2,
    fontSize - 1,
    "F2",
  );
  y -= lineHeight * 3;
  text(
    document.legalNotice,
    LEFT,
    CONTENT_WIDTH,
    "F2",
    Math.max(fontSize - 1, 6),
  );
  return pages;
}

function addHeader(
  elements: PdfElement[],
  document: ReturnType<typeof buildReceiptDocument>,
  fontSize: number,
  getY: () => number,
  setY: (value: number) => void,
) {
  let y = getY();
  addLine(elements, document.company.name, LEFT, y, fontSize + 3, "F2");
  y -= fontSize + 4;
  for (const line of document.company.lines) {
    for (const part of wrapText(
      line,
      maxCharacters(CONTENT_WIDTH, fontSize - 1),
    )) {
      addLine(elements, part, LEFT, y, Math.max(fontSize - 1, 6));
      y -= fontSize + 1;
    }
  }
  addLine(elements, document.receiptNumber, LEFT, y, fontSize + 1, "F2");
  addLine(elements, document.issuedAtLabel, 164, y, fontSize);
  y -= fontSize + 5;
  addRule(elements, y, LEFT, RIGHT);
  setY(y - 6);
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
) {
  elements.push({ kind: "text", text, x, y, size, font });
}

function addRule(elements: PdfElement[], y: number, x1 = LEFT, x2 = RIGHT) {
  elements.push({ kind: "rule", x1, x2, y });
}

function drawElement(element: PdfElement) {
  if (element.kind === "rule")
    return `0.6 w ${element.x1} ${element.y} m ${element.x2} ${element.y} l S`;
  return drawText(element);
}

function drawText(line: PdfText) {
  return `BT /${line.font || "F1"} ${line.size || 10} Tf 1 0 0 1 ${line.x} ${line.y} Tm (${escapePdfString(line.text)}) Tj ET`;
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
