-- Application configurations
CREATE TABLE application_configs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    namespace VARCHAR(100) NOT NULL DEFAULT 'default',
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    config_data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, namespace, user_id)
);

-- Triggers to update timestamps
CREATE TRIGGER update_application_configs_timestamp BEFORE UPDATE ON application_configs
FOR EACH ROW EXECUTE FUNCTION update_modified_column();

-- Indexes for better performance
CREATE INDEX idx_application_configs_user_id ON application_configs(user_id);