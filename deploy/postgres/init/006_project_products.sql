CREATE TABLE IF NOT EXISTS project_products (
  project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (project_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_project_products_project ON project_products(project_id);
CREATE INDEX IF NOT EXISTS idx_project_products_product ON project_products(product_id);
