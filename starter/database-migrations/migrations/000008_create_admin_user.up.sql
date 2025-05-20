-- Create admin user with password 'admin123' (hashed with bcrypt)
INSERT INTO users (id, username, email, password, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'admin',
    'admin@taskify.com',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', -- bcrypt hash of 'admin123'
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Assign admin role to the admin user
INSERT INTO user_roles (user_id, role_id)
SELECT 
    u.id,
    r.id
FROM users u
CROSS JOIN roles r
WHERE u.username = 'admin'
AND r.name = 'admin'; 