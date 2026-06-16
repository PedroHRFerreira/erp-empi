CREATE TABLE users (
  id CHAR(36) PRIMARY KEY,
  name VARCHAR(140) NOT NULL,
  cpf VARCHAR(11) NOT NULL UNIQUE,
  password_hash VARCHAR(255),
  type VARCHAR(20) NOT NULL,
  email VARCHAR(180),
  phone VARCHAR(20),
  markup_percent NUMERIC NOT NULL DEFAULT 10,
  address VARCHAR(255),
  notes VARCHAR(500),
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

CREATE TABLE stock_items (
  id CHAR(36) PRIMARY KEY,
  name VARCHAR(140) NOT NULL,
  description VARCHAR(500),
  cost_cents BIGINT NOT NULL,
  markup_percent NUMERIC NOT NULL DEFAULT 10,
  resale_price_cents BIGINT NOT NULL,
  quantity INTEGER NOT NULL DEFAULT 0,
  used_quantity INTEGER NOT NULL DEFAULT 0,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

CREATE TABLE receipts (
  id CHAR(36) PRIMARY KEY,
  user_id CHAR(36) NOT NULL REFERENCES users(id),
  vehicle_model VARCHAR(140) NOT NULL,
  vehicle_year INTEGER NOT NULL,
  vehicle_plate VARCHAR(12) NOT NULL,
  services VARCHAR(700) NOT NULL,
  price_cents BIGINT NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  notes VARCHAR(700),
  paid_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

CREATE TABLE receipt_items (
  id CHAR(36) PRIMARY KEY,
  receipt_id CHAR(36) NOT NULL REFERENCES receipts(id),
  stock_item_id CHAR(36) NOT NULL REFERENCES stock_items(id),
  quantity INTEGER NOT NULL,
  unit_cost_cents BIGINT NOT NULL,
  unit_resale_cents BIGINT NOT NULL,
  markup_percent NUMERIC NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);
