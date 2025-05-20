# Taskify Implementation Guide

This document details the implementation of the Taskify project, focusing on the user registration, authentication system, and role-based access control (RBAC).

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Database Setup](#database-setup)
3. [Tasks Table Migration](#tasks-table-migration)
4. [Role-Based Access Control](#role-based-access-control)
5. [Environment Configuration](#environment-configuration)
6. [Running the Application](#running-the-application)
7. [API Endpoints](#api-endpoints)
8. [Testing the Implementation](#testing-the-implementation)
9. [Security Features](#security-features)
10. [Troubleshooting](#troubleshooting)

## Prerequisites

Before starting, ensure you have the following installed:
- Go 1.23 or higher
- PostgreSQL 17
- Docker (optional, for running PostgreSQL)

## Database Setup

### Using Docker (Recommended)
```bash
# Start PostgreSQL container
docker run -itd --name postgres --restart=always \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=strongpass123 \
  -e POSTGRES_USER=taskmanager \
  -e POSTGRES_DB=taskmanager \
  postgres:17
```

### Manual PostgreSQL Setup
1. Install PostgreSQL 17
2. Create a new database and user:
```sql
CREATE DATABASE taskmanager;
CREATE USER taskmanager WITH PASSWORD 'strongpass123';
GRANT ALL PRIVILEGES ON DATABASE taskmanager TO taskmanager;
```

## Tasks Table Migration

### Creating the Migration

To add support for tasks, create a new migration in `database-migrations/migrations`:
- `000003_create_tasks_table.up.sql` (creates the table and indexes)
- `000003_create_tasks_table.down.sql` (drops the table)

**Schema:**
- `id` (UUID, primary key)
- `title` (string, required)
- `description` (text, optional)
- `status` (string, default 'pending')
- `priority` (string, default 'medium')
- `due_date` (timestamp, optional)
- `user_id` (UUID, required, foreign key to users)
- `created_at`, `updated_at`, `deleted_at` (timestamps)
- Indexes on `user_id`, `status`, `priority`, `due_date`

**Example up migration:**
```sql
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    priority VARCHAR(50) NOT NULL DEFAULT 'medium',
    due_date TIMESTAMP,
    user_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_tasks_user_id ON tasks(user_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);
```

**Example down migration:**
```sql
DROP TABLE IF EXISTS tasks;
```

### Running the Migration

1. Ensure your database is running and environment variables are set.
2. From the `database-migrations` directory, run:
   ```bash
   go run main.go up
   ```
3. If you encounter a dirty migration state, manually update the `schema_migrations` table:
   ```sql
   UPDATE schema_migrations SET version = <latest_version>, dirty = false;
   ```
   Then re-run the migration.

### Verifying the Migration

Check that the `tasks` table exists:
```bash
psql -h localhost -p 5432 -U taskmanager -d taskmanager -c "\dt"
```
You should see `tasks` in the list of tables.

## Role-Based Access Control

### Database Schema

The RBAC system consists of four main tables:

1. **Roles Table**
```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

2. **Permissions Table**
```sql
CREATE TABLE permissions (
    id UUID PRIMARY KEY,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(resource, action)
);
```

3. **User Roles Table**
```sql
CREATE TABLE user_roles (
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);
```

4. **Role Permissions Table**
```sql
CREATE TABLE role_permissions (
    role_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);
```

### Default Data

The system comes pre-configured with:

1. **Roles:**
   - Admin: Full system access
   - User: Basic access

2. **Permissions:**
   - Tasks: create, read, update, delete
   - Users: create, read, update, delete
   - Roles: assign, revoke

3. **Role-Permission Mappings:**
   - Admin role: All permissions
   - User role: Basic permissions
     - All task operations
     - User read permission

4. **Default Admin User:**
   - Username: admin
   - Email: admin@taskify.com
   - Password: admin123
   - Role: admin

### Running the Migrations

1. Ensure your database is running and environment variables are set.
2. From the `database-migrations` directory, run:
   ```bash
   go run main.go up
   ```

### Verifying the Setup

Check that the tables and data are properly created:
```bash
# Check roles
psql -h localhost -p 5432 -U taskmanager -d taskmanager -c "SELECT * FROM roles;"

# Check permissions
psql -h localhost -p 5432 -U taskmanager -d taskmanager -c "SELECT * FROM permissions;"

# Check role permissions
psql -h localhost -p 5432 -U taskmanager -d taskmanager -c "SELECT r.name as role, p.resource, p.action FROM role_permissions rp JOIN roles r ON r.id = rp.role_id JOIN permissions p ON p.id = rp.permission_id ORDER BY r.name, p.resource, p.action;"

# Check admin user and role
psql -h localhost -p 5432 -U taskmanager -d taskmanager -c "SELECT u.username, u.email, r.name as role FROM users u JOIN user_roles ur ON u.id = ur.user_id JOIN roles r ON r.id = ur.role_id WHERE u.username = 'admin';"
```

## Environment Configuration

Set up the required environment variables:
```bash
# Database configuration
export DB_USER=taskmanager
export DB_PASSWORD=strongpass123
export DB_NAME=taskmanager
export DB_HOST=localhost
export DB_PORT=5432

# JWT configuration
export JWT_SECRET=your-secret-key
```

## Running the Application

1. Clone the repository:
```bash
git clone <repository-url>
cd taskify
```

2. Install dependencies:
```bash
cd backend
go mod download
```

3. Start the server:
```bash
go run main.go
```

The server will start on `http://localhost:8080` with the following endpoints available:
- POST `/api/v1/auth/register` - User registration
- POST `/api/v1/auth/login` - User login
- POST `/api/v1/auth/refresh` - Refresh access token

## API Endpoints

### User Registration
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

Response:
```json
{
  "message": "user created successfully"
}
```

### User Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "550e8400-e29b-41d4-a716-446655440000",
  "expires_in": 3600
}
```

### Refresh Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "660e8400-e29b-41d4-a716-446655440000",
  "expires_in": 3600
}
```

## Postman Collection

Here's a Postman collection to test all endpoints. Import this JSON into Postman:

```json
{
  "info": {
    "name": "Taskify API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Authentication",
      "item": [
        {
          "name": "Register User",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "url": {
              "raw": "http://localhost:8080/api/v1/auth/register",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8080",
              "path": ["api", "v1", "auth", "register"]
            },
            "body": {
              "mode": "raw",
              "raw": "{\n    \"username\": \"testuser\",\n    \"email\": \"test@example.com\",\n    \"password\": \"password123\"\n}"
            }
          }
        },
        {
          "name": "Login User",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "url": {
              "raw": "http://localhost:8080/api/v1/auth/login",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8080",
              "path": ["api", "v1", "auth", "login"]
            },
            "body": {
              "mode": "raw",
              "raw": "{\n    \"username\": \"testuser\",\n    \"password\": \"password123\"\n}"
            }
          }
        },
        {
          "name": "Refresh Token",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "url": {
              "raw": "http://localhost:8080/api/v1/auth/refresh",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8080",
              "path": ["api", "v1", "auth", "refresh"]
            },
            "body": {
              "mode": "raw",
              "raw": "{\n    \"refresh_token\": \"{{refresh_token}}\"\n}"
            }
          }
        }
      ]
    },
    {
      "name": "User Management",
      "item": [
        {
          "name": "Get User Profile",
          "request": {
            "method": "GET",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{access_token}}"
              }
            ],
            "url": {
              "raw": "http://localhost:8080/api/v1/users/profile",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8080",
              "path": ["api", "v1", "users", "profile"]
            }
          }
        },
        {
          "name": "Get User Profile by ID",
          "request": {
            "method": "GET",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{access_token}}"
              }
            ],
            "url": {
              "raw": "http://localhost:8080/api/v1/users/profile/{{user_id}}",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8080",
              "path": ["api", "v1", "users", "profile", "{{user_id}}"]
            }
          }
        },
        {
          "name": "Get All Users",
          "request": {
            "method": "GET",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{access_token}}"
              }
            ],
            "url": {
              "raw": "http://localhost:8080/api/v1/users",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8080",
              "path": ["api", "v1", "users"]
            }
          }
        },
        {
          "name": "Delete User",
          "request": {
            "method": "DELETE",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{access_token}}"
              }
            ],
            "url": {
              "raw": "http://localhost:8080/api/v1/users/{{user_id}}",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8080",
              "path": ["api", "v1", "users", "{{user_id}}"]
            }
          }
        }
      ]
    },
    {
      "name": "Task Management",
      "item": [
        {
          "name": "Create Task",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{access_token}}"
              },
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "url": {
              "raw": "http://localhost:8080/api/v1/tasks",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8080",
              "path": ["api", "v1", "tasks"]
            },
            "body": {
              "mode": "raw",
              "raw": "{\n    \"title\": \"Test Task\",\n    \"description\": \"This is a test task\",\n    \"status\": \"pending\",\n    \"priority\": \"medium\"\n}"
            }
          }
        },
        {
          "name": "Get All Tasks",
          "request": {
            "method": "GET",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{access_token}}"
              }
            ],
            "url": {
              "raw": "http://localhost:8080/api/v1/tasks",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8080",
              "path": ["api", "v1", "tasks"]
            }
          }
        },
        {
          "name": "Get Task by ID",
          "request": {
            "method": "GET",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{access_token}}"
              }
            ],
            "url": {
              "raw": "http://localhost:8080/api/v1/tasks/{{task_id}}",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8080",
              "path": ["api", "v1", "tasks", "{{task_id}}"]
            }
          }
        },
        {
          "name": "Update Task",
          "request": {
            "method": "PUT",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{access_token}}"
              },
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "url": {
              "raw": "http://localhost:8080/api/v1/tasks/{{task_id}}",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8080",
              "path": ["api", "v1", "tasks", "{{task_id}}"]
            },
            "body": {
              "mode": "raw",
              "raw": "{\n    \"title\": \"Updated Task\",\n    \"description\": \"This is an updated task\",\n    \"status\": \"in_progress\",\n    \"priority\": \"high\"\n}"
            }
          }
        },
        {
          "name": "Delete Task",
          "request": {
            "method": "DELETE",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{access_token}}"
              }
            ],
            "url": {
              "raw": "http://localhost:8080/api/v1/tasks/{{task_id}}",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8080",
              "path": ["api", "v1", "tasks", "{{task_id}}"]
            }
          }
        },
        {
          "name": "Get User Tasks",
          "request": {
            "method": "GET",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{access_token}}"
              }
            ],
            "url": {
              "raw": "http://localhost:8080/api/v1/users/{{user_id}}/tasks",
              "protocol": "http",
              "host": ["localhost"],
              "port": "8080",
              "path": ["api", "v1", "users", "{{user_id}}", "tasks"]
            }
          }
        }
      ]
    }
  ],
  "variable": [
    {
      "key": "access_token",
      "value": "your_access_token_here"
    },
    {
      "key": "refresh_token",
      "value": "your_refresh_token_here"
    },
    {
      "key": "user_id",
      "value": "user_uuid_here"
    },
    {
      "key": "task_id",
      "value": "task_uuid_here"
    }
  ]
}
```

### Using the Postman Collection

1. Import the collection into Postman
2. Set up environment variables:
   - `access_token`: JWT access token from login
   - `refresh_token`: Refresh token from login
   - `user_id`: UUID of a user for testing
   - `task_id`: UUID of a task for testing

3. Test Flow:
   1. Register a new user
   2. Login with the user credentials
   3. Copy the access token and refresh token to environment variables
   4. Test protected endpoints using the access token
   5. When the access token expires, use the refresh token endpoint
   6. Update the access token in environment variables

4. Testing Security:
   - Try accessing protected endpoints without a token
   - Try accessing admin endpoints with a regular user token
   - Test SQL injection prevention by using special characters in inputs
   - Verify proper error responses for invalid inputs

## Testing the Implementation

1. Register a new user:
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

2. Verify user creation in database:
```bash
psql -h localhost -p 5432 -U taskmanager -d taskmanager -c "SELECT id, username, email FROM users;"
```

3. Test login with the created user:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

4. Test token refresh:
```bash
# Save the refresh token from the login response
REFRESH_TOKEN="your-refresh-token"

# Use the refresh token to get new tokens
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }"
```

5. Use the access token for authenticated requests:
```bash
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer <access_token>"
```

6. Test Role-Based Access Control:
```bash
# Try accessing admin-only route with regular user token
curl -X GET http://localhost:8080/api/v1/admin/dashboard \
  -H "Authorization: Bearer <access_token>"

# Expected response for non-admin users:
# {"error":"insufficient permissions"}
```

### Role-Based Access Control Testing

The system implements role-based access control through middleware that checks user roles and permissions:

1. **Admin Route Protection**
   - Routes under `/api/v1/admin/*` are protected by both authentication and admin role requirement
   - Example admin route: `/api/v1/admin/dashboard`
   - Only users with the "admin" role can access these routes
   - Regular users receive a 403 Forbidden response with "insufficient permissions" message

2. **Role Verification**
   - The JWT token includes the user's roles and permissions
   - Middleware checks these roles against route requirements
   - Example JWT payload for a regular user:
     ```json
     {
       "user_id": "uuid",
       "username": "testuser",
       "roles": ["user"],
       "permissions": [
         "tasks:create",
         "tasks:read",
         "tasks:update",
         "tasks:delete",
         "users:read"
       ]
     }
     ```

3. **Permission-Based Access**
   - Each role has specific permissions
   - Permissions are checked for protected operations
   - Example: Only users with "tasks:create" permission can create tasks

4. **Testing Different Access Levels**
   - Regular users can access basic endpoints
   - Admin-only routes are protected
   - Permission-specific operations are restricted
   - Invalid or expired tokens are rejected
   - Missing tokens result in 401 Unauthorized

## Security Features

1. **Password Security**
   - Bcrypt hashing with default cost factor
   - Minimum password length requirement (6 characters)
   - Secure password comparison

2. **JWT Authentication**
   - Access tokens with 1-hour expiration
   - Refresh token mechanism with database storage
   - Secure token generation and validation
   - Automatic token rotation on refresh

3. **Role-Based Access Control**
   - Role-based permission system
   - Granular resource and action permissions
   - Default roles with appropriate permissions
   - Secure role assignment and revocation

4. **Refresh Token Security**
   - Refresh tokens stored in database with expiration
   - One-time use refresh tokens (invalidated after use)
   - Automatic cleanup of used tokens
   - UUID-based refresh tokens for uniqueness

5. **Input Validation**
   - Required field validation
   - Email format validation
   - Username and email uniqueness checks

6. **Database Security**
   - UUID for user identification
   - Unique constraints on username and email
   - Soft delete support
   - Proper error handling
   - Foreign key constraints with CASCADE delete

7. **SQL Injection Prevention**
   - All database queries use parameterized queries
   - GORM's query builder for safe SQL generation
   - Input validation before database operations
   - UUID validation for user IDs
   - Proper error handling for database operations
   - Example of secure query:
     ```go
     // Secure implementation using parameterized queries
     result := db.Preload("Roles").Where("id = ?", userID).First(&user)
     
     // Avoid string concatenation for SQL queries
     // ❌ Vulnerable to SQL injection:
     // query := fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", userID)
     // ✅ Use parameterized queries instead:
     // db.Where("id = ?", userID)
     ```

## Troubleshooting

### Common Issues

1. **Database Connection Issues**
   - Error: "password authentication failed for user 'postgres'"
   - Solution: Ensure correct environment variables are set:
     ```bash
     export DB_USER=taskmanager
     export DB_PASSWORD=strongpass123
     export DB_NAME=taskmanager
     export DB_HOST=localhost
     export DB_PORT=5432
     ```

2. **Port Already in Use**
   - Error: "listen tcp :8080: bind: address already in use"
   - Solution: Find and kill the process using port 8080:
     ```bash
     sudo lsof -i :8080 | awk 'NR!=1 {print $2}' | xargs kill -9
     ```

3. **JWT Token Issues**
   - Error: "invalid memory address or nil pointer dereference"
   - Solution: Ensure JWT_SECRET is set:
     ```bash
     export JWT_SECRET=your-secret-key
     ```

4. **Refresh Token Issues**
   - Error: "invalid refresh token"
   - Solution: Ensure you're using a valid refresh token from a recent login
   - Error: "refresh token expired"
   - Solution: Login again to get new tokens

### Debugging Tips

1. Check server logs for detailed error messages
2. Verify database connection using psql
3. Test API endpoints with curl to isolate issues
4. Ensure all environment variables are set correctly
5. Check token expiration times in the database

## Next Steps

1. Implement role-based middleware for route protection
2. Add role management endpoints
3. Implement permission checking in handlers
4. Add user profile management
5. Implement password reset functionality
6. Add rate limiting for authentication endpoints

## Learning Outcomes

Through this implementation, we've learned:
- Secure password handling with bcrypt
- JWT-based authentication
- Refresh token implementation and security
- Role-based access control (RBAC)
- Permission management
- Input validation and sanitization
- Database security best practices
- Error handling and user feedback
- RESTful API design principles
- Environment configuration management
- Docker containerization 