CREATE TABLE IF NOT EXISTS expenses (
  id CHAR(36) PRIMARY KEY,
  description VARCHAR(180) NOT NULL,
  category VARCHAR(80) NOT NULL,
  amount_cents BIGINT NOT NULL,
  spent_at TIMESTAMP NOT NULL,
  notes VARCHAR(700),
  archived_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses(category);
CREATE INDEX IF NOT EXISTS idx_expenses_spent_at ON expenses(spent_at);
CREATE INDEX IF NOT EXISTS idx_expenses_archived_at ON expenses(archived_at);
