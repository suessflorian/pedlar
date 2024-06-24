CREATE TABLE IF NOT EXISTS items (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  unit_scale VARCHAR(255) DEFAULT 'unit'
);

CREATE TABLE IF NOT EXISTS item_relationships (
  parent_id INTEGER NOT NULL,
  child_id INTEGER NOT NULL,
  PRIMARY KEY (parent_id, child_id),
  FOREIGN KEY (parent_id) REFERENCES items(id) ON DELETE CASCADE,
  FOREIGN KEY (child_id) REFERENCES items(id) ON DELETE CASCADE
);

CREATE OR REPLACE FUNCTION check_cycle() RETURNS TRIGGER AS $$
DECLARE
  cycle BOOLEAN;
BEGIN
  WITH RECURSIVE cycle_check AS (
    SELECT parent_id, child_id
    FROM item_relationships
    WHERE parent_id = NEW.child_id

    UNION ALL

    SELECT ir.parent_id, ic.child_id
    FROM item_relationships ir
    JOIN cycle_check ic ON ir.child_id = ic.parent_id
  )
  SELECT TRUE INTO cycle
  FROM cycle_check
  WHERE child_id = NEW.parent_id;

  IF cycle THEN
    RAISE EXCEPTION 'cyclical relationship detected';
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER check_cycle_trigger
BEFORE INSERT ON item_relationships
FOR EACH ROW
EXECUTE FUNCTION check_cycle();
