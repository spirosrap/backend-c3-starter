# Taskify Implementation Guide

This document details the implementation of the Taskify project, focusing on the user registration and authentication system.

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Database Setup](#database-setup)
3. [Tasks Table Migration](#tasks-table-migration)
4. [Environment Configuration](#environment-configuration)
5. [Running the Application](#running-the-application)
6. [API Endpoints](#api-endpoints)
7. [Testing the Implementation](#testing-the-implementation)
8. [Security Features](#security-features)
9. [Troubleshooting](#troubleshooting)

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

3. **Refresh Token Security**
   - Refresh tokens stored in database with expiration
   - One-time use refresh tokens (invalidated after use)
   - Automatic cleanup of used tokens
   - UUID-based refresh tokens for uniqueness

4. **Input Validation**
   - Required field validation
   - Email format validation
   - Username and email uniqueness checks

5. **Database Security**
   - UUID for user identification
   - Unique constraints on username and email
   - Soft delete support
   - Proper error handling

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

1. Add role-based access control
2. Implement task management features
3. Add user profile management
4. Implement password reset functionality
5. Add rate limiting for authentication endpoints

## Learning Outcomes

Through this implementation, we've learned:
- Secure password handling with bcrypt
- JWT-based authentication
- Refresh token implementation and security
- Input validation and sanitization
- Database security best practices
- Error handling and user feedback
- RESTful API design principles
- Environment configuration management
- Docker containerization 