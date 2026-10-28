-- One-off correction approved on 2026-10-27. Default: read-only preview.
-- Apply in the same psql session with: SET empi.apply_ss_stock = 'true';
-- Capture/review the preview JSON before applying. No records are deleted.
BEGIN;
SET LOCAL lock_timeout = '3s';
SET LOCAL statement_timeout = '15s';

DO $$
DECLARE
  product stock_items%ROWTYPE;
  purchase stock_purchases%ROWTYPE;
  part payable_installments%ROWTYPE;
BEGIN
  SELECT * INTO STRICT product FROM stock_items
    WHERE id = 'a1fec4a7-ca57-4940-b6a5-36d8737675dd' FOR UPDATE;
  SELECT * INTO STRICT purchase FROM stock_purchases
    WHERE id = '2db5a754-f69a-4a6e-be92-0d363c5043c9' FOR UPDATE;
  SELECT * INTO STRICT part FROM payable_installments
    WHERE id = 'aaf6b07c-9b40-448b-9b17-73f1cb8364e2' FOR UPDATE;

  IF product.name <> 'RESERVATORIO DE ÁGUA E TAMPA'
    OR product.cost_cents <> 9100 OR product.resale_price_cents <> 13100
    OR product.quantity <> 1 OR product.used_quantity <> 1 OR NOT product.active
    OR purchase.stock_item_id <> product.id OR purchase.supplier_name <> 'SS PEÇAS'
    OR purchase.status <> 'confirmed' OR purchase.total_cents <> 13100
    OR part.stock_purchase_id <> purchase.id OR part.status <> 'pending'
    OR part.amount_cents <> 13100 OR part.cash_entry_id IS NOT NULL
    OR (SELECT count(*) FROM payable_installments WHERE stock_purchase_id = purchase.id) <> 1
    OR (SELECT count(*) FROM stock_purchase_items WHERE stock_purchase_id = purchase.id) <> 1
    OR NOT EXISTS (SELECT 1 FROM stock_purchase_items WHERE stock_purchase_id = purchase.id
      AND stock_item_id = product.id AND quantity = 1 AND unit_cost_cents = 13100)
    OR EXISTS (SELECT 1 FROM cash_entries WHERE reference_id IN (purchase.id, part.id))
    OR EXISTS (SELECT 1 FROM receipt_items i JOIN receipts r ON r.id = i.receipt_id
      WHERE i.stock_item_id = product.id AND r.status = 'pending')
    OR NOT EXISTS (SELECT 1 FROM receipts r JOIN receipt_items i ON i.receipt_id = r.id
      WHERE r.id = '1f3e774d-46d5-417c-8354-4572f4beca68' AND r.status = 'paid'
      AND r.price_cents = 124149 AND i.stock_item_id = product.id AND i.quantity = 1)
    OR NOT EXISTS (SELECT 1 FROM payable_installments
      WHERE id = 'a9bb9f92-bf83-4e9f-9e93-5f6c5c2634cc' AND status = 'paid' AND amount_cents = 9100)
  THEN
    RAISE EXCEPTION 'State differs from reviewed correction; no changes applied';
  END IF;
END $$;

-- Exact original rows form the recoverable backup; no customer data is selected.
SELECT json_build_object(
  'product', (SELECT row_to_json(s) FROM stock_items s WHERE id = 'a1fec4a7-ca57-4940-b6a5-36d8737675dd'),
  'purchase', (SELECT row_to_json(p) FROM stock_purchases p WHERE id = '2db5a754-f69a-4a6e-be92-0d363c5043c9'),
  'installment', (SELECT row_to_json(i) FROM payable_installments i WHERE id = 'aaf6b07c-9b40-448b-9b17-73f1cb8364e2')
) AS before_correction;

DO $$
BEGIN
  IF COALESCE(current_setting('empi.apply_ss_stock', true), 'false') = 'true' THEN
    UPDATE payable_installments SET status = 'cancelled', updated_at = now()
      WHERE id = 'aaf6b07c-9b40-448b-9b17-73f1cb8364e2';
    UPDATE stock_purchases SET status = 'cancelled', cancelled_at = now(), updated_at = now()
      WHERE id = '2db5a754-f69a-4a6e-be92-0d363c5043c9';
    UPDATE stock_items SET quantity = 0, active = false, updated_at = now()
      WHERE id = 'a1fec4a7-ca57-4940-b6a5-36d8737675dd';
  END IF;
END $$;
COMMIT;
