CREATE TABLE IF NOT EXISTS monthly_goal_targets (
  month date PRIMARY KEY,
  revenue_target_cents bigint NOT NULL DEFAULT 0,
  labor_target_cents bigint NOT NULL DEFAULT 0,
  products_target_cents bigint NOT NULL DEFAULT 0,
  clients_target integer NOT NULL DEFAULT 0,
  net_profit_target_cents bigint NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);
