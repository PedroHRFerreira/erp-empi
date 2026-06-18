ALTER TABLE expenses
  ADD COLUMN IF NOT EXISTS receipt_id CHAR(36);

CREATE INDEX IF NOT EXISTS idx_expenses_receipt_id ON expenses(receipt_id);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'fk_expenses_receipt'
  ) THEN
    ALTER TABLE expenses
      ADD CONSTRAINT fk_expenses_receipt
      FOREIGN KEY (receipt_id)
      REFERENCES receipts(id)
      ON UPDATE CASCADE
      ON DELETE SET NULL;
  END IF;
END $$;
