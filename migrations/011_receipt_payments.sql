CREATE TABLE IF NOT EXISTS receipt_payments (
 id CHAR(36) PRIMARY KEY,
 receipt_id CHAR(36) NOT NULL REFERENCES receipts(id) ON UPDATE CASCADE ON DELETE CASCADE,
 position INTEGER NOT NULL,
 payment_method VARCHAR(30) NOT NULL,
 base_amount_cents BIGINT NOT NULL,
 card_fee_percent NUMERIC NOT NULL DEFAULT 0,
 card_fee_cents BIGINT NOT NULL DEFAULT 0,
 amount_cents BIGINT NOT NULL,
 installments INTEGER NOT NULL DEFAULT 1,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_receipt_payment_position ON receipt_payments(receipt_id, position);
-- Historical receipts are synthesized on read; no past fees or cash entries change.
