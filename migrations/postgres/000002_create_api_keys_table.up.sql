-- API keys for service authentication
CREATE TABLE api_keys (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(100) NOT NULL,
    name VARCHAR(50) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Triggers to update timestamps
CREATE TRIGGER update_api_keys_timestamp BEFORE UPDATE ON api_keys
FOR EACH ROW EXECUTE FUNCTION update_modified_column();

-- Indexes for better performance
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);