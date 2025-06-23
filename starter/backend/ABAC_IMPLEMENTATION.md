# ABAC (Attribute-Based Access Control) Implementation

This document describes the comprehensive ABAC implementation for the Taskify application, which enforces role-based and permission-based access control across all endpoints.

## Overview

The ABAC system implements a multi-layered security approach:

1. **Authentication**: JWT token validation
2. **Authorization**: Role and permission-based access control
3. **Resource Ownership**: Users can only access their own resources (unless admin)
4. **Admin Privileges**: Admins have full access to all resources

## Middleware Functions

### Core Authentication Middleware

#### `AuthMiddleware()`
- Validates JWT access tokens
- Extracts user information (ID, username, roles, permissions)
- Sets user context for downstream middleware and handlers

### Authorization Middleware

#### `RequireRole(roles ...string)`
- Checks if the user has any of the specified roles
- Returns 403 Forbidden if user lacks required roles

#### `RequirePermission(resource, action string)`
- Checks if the user has the specific permission (format: `resource:action`)
- Returns 403 Forbidden if user lacks required permission

#### `RequireRoleAndPermission(role, resource, action string)`
- Combines role and permission checks
- User must have BOTH the specified role AND permission
- Returns 403 Forbidden if either check fails

#### `RequireOwnershipOrAdmin(resourceIDParam string)`
- Checks if user owns the resource or is an admin
- Compares resource ID from URL parameter with user ID
- Returns 403 Forbidden if neither condition is met

### Specialized ABAC Middleware

#### `RequireTaskAccess()`
- Checks if user can access tasks
- Admins can access all tasks
- Regular users can only access their own tasks (enforced at service layer)

#### `RequireUserManagementAccess()`
- Only admins can manage users
- Returns 403 Forbidden for non-admin users

#### `RequireTaskManagementAccess()`
- Admins can manage all tasks
- Regular users can manage their own tasks (enforced at service layer)

## Route Policies

### Task Routes (`/api/v1/tasks`)

| Endpoint | Method | Policy | Description |
|----------|--------|--------|-------------|
| `/tasks` | POST | `RequirePermission("tasks", "create")` | Create new task (any authenticated user) |
| `/tasks/:id` | PUT | `RequirePermission("tasks", "update")` | Update task (owner or admin) |
| `/tasks/:id` | DELETE | `RequirePermission("tasks", "delete")` | Delete task (owner or admin) |
| `/tasks/:id` | GET | `RequirePermission("tasks", "read")` | Get specific task (owner or admin) |
| `/tasks` | GET | `RequireRoleAndPermission("admin", "tasks", "read")` | Get all tasks (admin only) |

### User Routes (`/api/v1/users`)

| Endpoint | Method | Policy | Description |
|----------|--------|--------|-------------|
| `/users/:user_id` | DELETE | `RequireRoleAndPermission("admin", "users", "delete")` | Delete user (admin only) |
| `/users` | GET | `RequireRoleAndPermission("admin", "users", "read")` | Get all users (admin only) |
| `/users/:user_id/tasks` | GET | `RequirePermission("tasks", "read")` | Get user's tasks (owner or admin) |
| `/users/profile` | GET | None (authenticated only) | Get own profile |
| `/users/profile/:user_id` | GET | `RequireRoleAndPermission("admin", "users", "read")` | Get user profile (admin only) |

## Permission Structure

### Available Permissions

#### Task Permissions
- `tasks:create` - Create new tasks
- `tasks:read` - Read tasks
- `tasks:update` - Update tasks
- `tasks:delete` - Delete tasks

#### User Permissions
- `users:create` - Create new users
- `users:read` - Read user information
- `users:update` - Update user information
- `users:delete` - Delete users

#### Role Permissions
- `roles:assign` - Assign roles to users
- `roles:revoke` - Revoke roles from users

### Role-Permission Mapping

#### Admin Role
- Has ALL permissions
- Can access ALL resources
- Can manage users and tasks

#### User Role
- Has task-related permissions only:
  - `tasks:create`
  - `tasks:read`
  - `tasks:update`
  - `tasks:delete`
- Can only access their own resources
- Cannot manage other users

## Implementation Details

### JWT Claims Structure

```go
type Claims struct {
    UserID      uuid.UUID `json:"user_id"`
    Username    string    `json:"username"`
    Roles       []string  `json:"roles"`
    Permissions []string  `json:"permissions"`
    jwt.RegisteredClaims
}
```

### Handler-Level Enforcement

In addition to middleware-level checks, handlers implement resource ownership validation:

```go
// Example from TaskHandler.GetTaskByID
func (h *TaskHandler) GetTaskByID(c *gin.Context) {
    // Get user info from context
    userID, exists := c.Get("user_id")
    userUUID := userID.(uuid.UUID)
    
    // Check if user is admin
    userRoles, exists := c.Get("roles")
    isAdmin := false
    if exists {
        rolesList := userRoles.([]string)
        for _, role := range rolesList {
            if role == "admin" {
                isAdmin = true
                break
            }
        }
    }

    // Get task and check ownership
    task, err := h.taskService.GetTaskByID(h.db, taskID)
    if !isAdmin && task.UserID != userUUID {
        c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
        return
    }
    
    c.JSON(http.StatusOK, task)
}
```

## Testing

The ABAC implementation includes comprehensive tests in `tests/abac_test.go`:

### Test Coverage

1. **Task Access Policies**
   - Admin can access all tasks
   - Regular users cannot access all tasks
   - Users can access their own tasks
   - Users cannot access other users' tasks
   - Admin can access any user's task
   - Users can update their own tasks
   - Users cannot update other users' tasks

2. **User Access Policies**
   - Admin can get all users
   - Regular users cannot get all users
   - Admin can get any user's profile
   - Regular users cannot get other users' profiles
   - Users can get their own profile
   - Admin can delete users
   - Regular users cannot delete users

3. **Permission Enforcement**
   - Admin has all permissions
   - Users have limited permissions
   - Role and permission combinations work correctly

## Usage Examples

### Adding New Protected Routes

```go
// Route requiring admin role only
router.GET("/admin-only", 
    middleware.AuthMiddleware(), 
    middleware.RequireRole("admin"), 
    handler.AdminOnlyHandler)

// Route requiring specific permission
router.POST("/tasks", 
    middleware.AuthMiddleware(), 
    middleware.RequirePermission("tasks", "create"), 
    handler.CreateTask)

// Route requiring both role and permission
router.DELETE("/users/:id", 
    middleware.AuthMiddleware(), 
    middleware.RequireRoleAndPermission("admin", "users", "delete"), 
    handler.DeleteUser)
```

### Adding New Permissions

1. Add permission to database migration:
```sql
INSERT INTO permissions (id, resource, action) VALUES 
    (gen_random_uuid(), 'reports', 'generate');
```

2. Assign to roles:
```sql
-- Admin gets all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT 
    (SELECT id FROM roles WHERE name = 'admin'),
    (SELECT id FROM permissions WHERE resource = 'reports' AND action = 'generate');
```

3. Use in routes:
```go
router.GET("/reports", 
    middleware.AuthMiddleware(), 
    middleware.RequirePermission("reports", "generate"), 
    handler.GenerateReport)
```

## Security Considerations

1. **Defense in Depth**: Multiple layers of security (middleware + handler)
2. **Principle of Least Privilege**: Users only get necessary permissions
3. **Resource Isolation**: Users can only access their own resources
4. **Admin Oversight**: Admins have full access for management purposes
5. **Token Validation**: JWT tokens are validated on every request
6. **Permission Granularity**: Fine-grained permissions for different actions

## Error Responses

The ABAC system returns appropriate HTTP status codes:

- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found (when appropriate)

Error messages are descriptive and help with debugging:

```json
{
  "error": "insufficient permissions - role required"
}
```

```json
{
  "error": "access denied - you can only access your own tasks"
}
```

## Best Practices

1. **Always use middleware**: Don't rely solely on handler-level checks
2. **Test thoroughly**: Use the provided test suite as a template
3. **Document policies**: Keep this document updated with new routes
4. **Monitor access**: Log access attempts for security auditing
5. **Regular reviews**: Periodically review and update permissions
6. **Fail securely**: Default to denying access when in doubt 