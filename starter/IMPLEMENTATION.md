# Taskify Implementation Guide

This document details the implementation of the Taskify project, focusing on the user registration system and its security features.

## Implementation Steps

### 1. User Registration Implementation

#### Database Setup
First, we set up the PostgreSQL database using Docker:
```bash
docker run -itd --name postgres --restart=always \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=strongpass123 \
  -e POSTGRES_USER=taskmanager \
  -e POSTGRES_DB=taskmanager \
  postgres:17
```

#### Environment Configuration
Set up the required environment variables:
```bash
export DB_USER=taskmanager
export DB_PASSWORD=strongpass123
export DB_NAME=taskmanager
export DB_HOST=localhost
export DB_PORT=5432
```

#### Database Migrations
Run the migrations to create the necessary tables:
```bash
# Create users table
psql -h localhost -p 5432 -U taskmanager -d taskmanager -f database-migrations/migrations/000001_create_users_table.up.sql

# Create tokens table
psql -h localhost -p 5432 -U taskmanager -d taskmanager -f database-migrations/migrations/000002_create_tokens_table.up.sql
```

### 2. Code Implementation

#### User Model
The user model is defined with the following structure:
```go
type User struct {
    gorm.Model
    ID       uuid.UUID `json:"id" gorm:"primaryKey"`
    Username string    `json:"username" gorm:"unique"`
    Email    string    `json:"email" gorm:"unique"`
    Password string    `json:"password"`
}
```

#### Registration Service
The registration service implements secure user creation:
```go
func (s *RegisterServiceImpl) RegisterUser(db *gorm.DB, user models.User) error {
    // Generate UUID
    id, err := uuid.NewV4()
    if err != nil {
        return err
    }
    user.ID = id

    // Hash password using bcrypt
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    user.Password = string(hashedPassword)

    // Create user in database
    result := db.Create(&user)
    return result.Error
}
```

#### Registration Handler
The handler implements input validation and duplicate checking:
```go
func (h *RegisterHandler) Registration(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Check for duplicate username
    var existingUser models.User
    if err := h.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
        return
    }

    // Check for duplicate email
    if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
        return
    }

    // Create new user
    user := models.User{
        Username: req.Username,
        Email:    req.Email,
        Password: req.Password,
    }

    if err := h.registerService.RegisterUser(h.db, user); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "user created successfully"})
}
```

### 3. Testing the Implementation

#### Start the Server
```bash
cd backend
go run main.go
```

#### Test User Registration
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

#### Verify User Creation
```bash
psql -h localhost -p 5432 -U taskmanager -d taskmanager -c "SELECT id, username, email FROM users;"
```

## Security Features Implemented

1. **Password Security**
   - Bcrypt hashing with default cost factor
   - Minimum password length requirement (6 characters)

2. **Input Validation**
   - Required field validation
   - Email format validation
   - Username and email uniqueness checks

3. **Database Security**
   - UUID for user identification
   - Unique constraints on username and email
   - Soft delete support
   - Proper error handling

## Common Issues and Solutions

1. **Database Connection Issues**
   - Ensure correct environment variables are set
   - Verify PostgreSQL is running
   - Check user permissions

2. **Migration Issues**
   - Run migrations in correct order
   - Ensure database exists before running migrations
   - Check user permissions for migration execution

## Next Steps

1. Implement user login functionality
2. Add JWT token generation
3. Implement refresh token mechanism
4. Add role-based access control
5. Implement task management features

## Learning Outcomes

Through this implementation, we've learned:
- Secure password handling with bcrypt
- Input validation and sanitization
- Database security best practices
- Error handling and user feedback
- RESTful API design principles 