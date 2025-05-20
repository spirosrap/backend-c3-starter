-- Create roles table if not exists
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);

-- Create permissions table if not exists
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    UNIQUE(resource, action)
);

-- Create role_permissions table if not exists
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id),
    FOREIGN KEY (permission_id) REFERENCES permissions(id)
);

-- Create user_roles table if not exists
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (role_id) REFERENCES roles(id)
);

-- Insert default roles
INSERT INTO roles (id, name, created_at, updated_at)
VALUES 
    ('11111111-1111-1111-1111-111111111111', 'admin', NOW(), NOW()),
    ('22222222-2222-2222-2222-222222222222', 'user', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- Insert default permissions
INSERT INTO permissions (id, resource, action, created_at, updated_at)
VALUES 
    -- Task permissions
    ('11111111-1111-1111-1111-111111111111', 'tasks', 'create', NOW(), NOW()),
    ('22222222-2222-2222-2222-222222222222', 'tasks', 'read', NOW(), NOW()),
    ('33333333-3333-3333-3333-333333333333', 'tasks', 'update', NOW(), NOW()),
    ('44444444-4444-4444-4444-444444444444', 'tasks', 'delete', NOW(), NOW()),
    -- User permissions
    ('55555555-5555-5555-5555-555555555555', 'users', 'read', NOW(), NOW()),
    ('66666666-6666-6666-6666-666666666666', 'users', 'update', NOW(), NOW()),
    ('77777777-7777-7777-7777-777777777777', 'users', 'delete', NOW(), NOW())
ON CONFLICT (resource, action) DO NOTHING;

-- Assign permissions to admin role
INSERT INTO role_permissions (role_id, permission_id, created_at, updated_at)
SELECT 
    '11111111-1111-1111-1111-111111111111', -- admin role id
    id,
    NOW(),
    NOW()
FROM permissions
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assign permissions to user role
INSERT INTO role_permissions (role_id, permission_id, created_at, updated_at)
SELECT 
    '22222222-2222-2222-2222-222222222222', -- user role id
    id,
    NOW(),
    NOW()
FROM permissions 
WHERE resource = 'tasks' AND action IN ('create', 'read', 'update', 'delete')
ON CONFLICT (role_id, permission_id) DO NOTHING; 