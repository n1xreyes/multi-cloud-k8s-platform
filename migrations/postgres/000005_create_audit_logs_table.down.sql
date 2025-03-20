DROP TABLE IF EXISTS audit_logs;

-- Since this is the last migration, we should also drop the shared function
-- that was created in the first migration
DROP FUNCTION IF EXISTS update_modified_column();