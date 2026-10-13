import type { MarketingBrand, MarketingCreation, MarketingCreationInput, MarketingTemplate } from "~/types/marketing"

/** JSON data may arrive as Vue proxies; every editor owns an independent draft. */
export function cloneMarketingInput(input: MarketingCreationInput): MarketingCreationInput {
  return JSON.parse(JSON.stringify(input)) as MarketingCreationInput
}

export function reopenMarketingCreation(creation: MarketingCreation, template?: MarketingTemplate): MarketingCreationInput {
  const { id, createdAt, updatedAt, ...data } = creation
  const input = cloneMarketingInput(data)
  // Missing layers identify creations saved before layer editing was available.
  // An explicitly empty list is intentional and must survive reopening.
  if (input.layers == null && template?.layers) input.layers = JSON.parse(JSON.stringify(template.layers))
  return input
}

/** A saved creation owns its text and colors, independently of later brand edits. */
export function createFromMarketingTemplate(
  template: MarketingTemplate,
  brand: MarketingBrand | null,
): MarketingCreationInput {
  return {
    template: template.id,
    format: "original",
    ...(template.layers ? { layers: JSON.parse(JSON.stringify(template.layers)) } : {}),
    kind: template.defaults.kind,
    objective: template.defaults.objective,
    title: template.defaults.title,
    description: template.defaults.description,
    caption: template.defaults.caption,
    cta: template.defaults.cta,
    primaryColor: template.defaults.primaryColor,
    secondaryColor: template.defaults.secondaryColor,
    priceCents: null,
    imageId: null,
    productId: null,
    logoImageId: brand?.logoImageId ?? null,
    brandName: brand?.name.trim() || "EMPI AutoCenter",
    brandContact: brand?.contact.trim() || "(33) 98735-1922 · Rua Tupã, 587",
  }
}
