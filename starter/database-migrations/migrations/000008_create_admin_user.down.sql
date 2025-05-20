-- Remove admin user's role assignment
DELETE FROM user_roles
WHERE user_id IN (SELECT id FROM users WHERE username = 'admin');

-- Remove admin user
DELETE FROM users WHERE username = 'admin'; 