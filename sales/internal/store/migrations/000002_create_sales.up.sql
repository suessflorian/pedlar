CREATE TABLE IF NOT EXISTS sales (
  id SERIAL PRIMARY KEY,
  retailer_id INTEGER NOT NULL,
  customer_id INTEGER NOT NULL,
  sale_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (retailer_id) REFERENCES retailers (id) ON DELETE CASCADE
);

CREATE TABLE line_items (
    id SERIAL PRIMARY KEY,
    sale_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price NUMERIC(10, 2) NOT NULL,
    FOREIGN KEY (sale_id) REFERENCES sales (id) ON DELETE CASCADE
);
