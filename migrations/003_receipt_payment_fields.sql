ALTER TABLE IF EXISTS users
  ADD COLUMN IF NOT EXISTS machine_fee_percent NUMERIC NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS installment_fee_percent NUMERIC NOT NULL DEFAULT 0;

ALTER TABLE IF EXISTS receipts
  ADD COLUMN IF NOT EXISTS labor_price_cents BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS products_total_cents BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS subtotal_cents BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS card_fee_percent NUMERIC NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS card_fee_cents BIGINT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS payment_method VARCHAR(30) NOT NULL DEFAULT 'cash',
  ADD COLUMN IF NOT EXISTS installments INTEGER NOT NULL DEFAULT 1;

UPDATE receipts
SET
  labor_price_cents = price_cents,
  subtotal_cents = price_cents
WHERE labor_price_cents = 0
  AND products_total_cents = 0
  AND subtotal_cents = 0;
