CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(resource, action)
);

-- Create indexes for faster lookups
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_action ON permissions(action);

-- Insert default permissions
INSERT INTO permissions (id, resource, action) VALUES
    -- Task permissions
    (gen_random_uuid(), 'tasks', 'create'),
    (gen_random_uuid(), 'tasks', 'read'),
    (gen_random_uuid(), 'tasks', 'update'),
    (gen_random_uuid(), 'tasks', 'delete'),
    -- User permissions
    (gen_random_uuid(), 'users', 'create'),
    (gen_random_uuid(), 'users', 'read'),
    (gen_random_uuid(), 'users', 'update'),
    (gen_random_uuid(), 'users', 'delete'),
    -- Role permissions
    (gen_random_uuid(), 'roles', 'assign'),
    (gen_random_uuid(), 'roles', 'revoke'); 