-- Cloud provider credentials
CREATE TABLE cloud_credentials (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(10) NOT NULL, -- 'aws', 'gcp', 'azure'
    name VARCHAR(50) NOT NULL,
    credentials JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, provider, name)
);

-- Triggers to update timestamps
CREATE TRIGGER update_cloud_credentials_timestamp BEFORE UPDATE ON cloud_credentials
FOR EACH ROW EXECUTE FUNCTION update_modified_column();

-- Indexes for better performance
CREATE INDEX idx_cloud_credentials_user_id ON cloud_credentials(user_id);