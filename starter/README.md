# Project Overview : Taskify : Task Management API
Railtronics is a leading provider of electronic components, employing thousands of professionals who currently coordinate their work through scattered channels like email, messaging, and phone calls. To streamline operations and foster seamless collaboration, we are developing a modern Taskify Application for Railtronics.

This intuitive platform will centralize task management, enabling employees to efficiently plan, track, and prioritize their work. Key features include real-time updates, automated reminders, easy delegation, and robust collaboration tools—empowering teams to stay aligned and accountable. With mobile access and integration with existing productivity tools, employees can manage tasks anytime, anywhere, boosting productivity and reducing communication silos.

By adopting this solution, Railtronics will enhance workflow visibility, improve team coordination, and ensure every project milestone is met—driving operational excellence across the organization.

This project involves building a secure, scalable backend server for task management, enabling CRUD operations via RESTful endpoints with role-based access control (RBAC). Users interact with the system programmatically through HTTP requests, adhering to industry-standard practices for API design and security.

## How will this help my career?
Learning IAM principles like RBAC, ABAC, authentication, and authorization is crucial for software developers because these concepts form the backbone of application security. Mastery of access control models (RBAC/ABAC) and secure authentication/authorization ensures only the right users access sensitive data, helping prevent breaches and comply with regulations. These skills are highly valued, opening doors to advanced roles in security and enterprise development. They also enable you to design scalable, maintainable, and user-friendly systems, making you a more versatile and in-demand developer in today's security-conscious tech landscape.

## Project Introduction
The Taskify application is designed to streamline task management within an organization. Key features include:

- **User Registration & Login:** Secure account creation and authentication for all users.
- **Role-Based Access:** Supports multiple user roles, such as regular employees and administrators.
    - **Administrators:** Have full control over user management, including onboarding new users and offboarding employees who leave the company, managing all tasks across company.
    - **Regular Employees:** Can register, log in, manage their own tasks, and update their personal profiles.
- **Task Management:** Allows users to create, update, track, and complete tasks efficiently.
- **Industry Best Practices:** The application is built following industry standards for security, usability, and performance.

This structure will ensure a secure, user-friendly, and efficient environment for managing tasks and users within Railtronics.

## Project Summary

Taskify application is a secure and user-friendly platform for managing tasks within an organization. It features role-based access for administrators and regular employees, allowing admins to manage users and employees to handle their own tasks and profiles. The application follows industry best practices to ensure reliability, security, and efficiency.

## Final Note

As you progress through this project, don't hesitate to refer back to the course materials—particularly the modules on API design, database integration, and Identity and Access Management. This is your chance to bring together all the concepts you've learned and apply them in a meaningful, real-world context. I'm looking forward to seeing your unique approach to this final challenge and hope you find the experience of building your own task management system both rewarding and enjoyable. Best of luck!


## Running End to End Project with Docker Compose
```shell
docker compose up -d
```

Access Project at `http://host.docker.internal`

# Taskify - Task Management System

A secure, scalable task management system built with Go and React, implementing modern architectural patterns and security best practices.

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [Design Decisions](#design-decisions)
3. [Security Implementation](#security-implementation)
4. [Database Design](#database-design)
5. [API Design](#api-design)
6. [Authentication & Authorization](#authentication--authorization)
7. [Error Handling](#error-handling)
8. [Testing Strategy](#testing-strategy)
9. [Deployment](#deployment)
10. [Development Setup](#development-setup)

## Architecture Overview

### Backend Architecture
- **Clean Architecture** pattern with clear separation of concerns:
  - `handlers`: HTTP request handling and response formatting
  - `services`: Business logic implementation
  - `models`: Data structures and database models
  - `middleware`: Request processing and authentication
  - `utils`: Shared utilities and helpers

### Frontend Architecture
- **React** with modern hooks and context API
- **Material-UI** for consistent UI components
- **Context-based** state management
- **Protected Routes** for authenticated access

## Design Decisions

### 1. Database Layer
- **GORM** as ORM for type-safe database operations
- **UUID** for primary keys to ensure global uniqueness
- **Soft Delete** implementation for data retention
- **Indexed Fields** for optimized queries
- **Foreign Key Constraints** for data integrity

### 2. API Design
- **RESTful** principles with resource-based endpoints
- **Versioned API** (`/api/v1/`) for future compatibility
- **Consistent Response Format**:
  ```json
  {
    "data": {},
    "error": null,
    "message": "success"
  }
  ```
- **HTTP Status Codes** for clear response semantics

### 3. Security Measures
- **JWT-based** authentication
- **Role-Based Access Control** (RBAC)
- **Parameterized Queries** to prevent SQL injection
- **Input Validation** at multiple layers
- **Secure Password Hashing** with bcrypt
- **CORS** configuration for API access control

### 4. Error Handling
- **Centralized Error Handling** middleware
- **Structured Error Responses**
- **Logging** for debugging and monitoring
- **Graceful Degradation** for service failures

## Security Implementation

### Authentication Flow
1. User registration with password hashing
2. JWT token generation on login
3. Token refresh mechanism
4. Token validation middleware

### Authorization
- **Role-based** access control
- **Permission-based** operations
- **Middleware** for route protection
- **Resource ownership** validation

### Data Protection
- **Input Sanitization**
- **SQL Injection Prevention**
- **XSS Protection**
- **CSRF Protection**

## Database Design

### Core Tables
1. **Users**
   - UUID primary key
   - Username and email uniqueness
   - Password hashing
   - Role associations

2. **Tasks**
   - UUID primary key
   - User ownership
   - Status tracking
   - Priority levels
   - Due dates

3. **Roles & Permissions**
   - Role definitions
   - Permission mappings
   - User-role associations

## API Design

### Endpoint Structure
```
/api/v1/
├── auth/
│   ├── register
│   ├── login
│   └── refresh
├── users/
│   ├── profile
│   └── {user_id}/tasks
└── tasks/
    ├── {task_id}
    └── {task_id}/status
```

### Response Format
```json
{
  "data": {
    "id": "uuid",
    "attributes": {}
  },
  "error": null,
  "message": "success"
}
```

## Error Handling

### Error Types
1. **Validation Errors**
   - Input validation failures
   - Business rule violations

2. **Authentication Errors**
   - Invalid credentials
   - Expired tokens
   - Missing permissions

3. **System Errors**
   - Database failures
   - Service unavailability
   - Internal errors

### Error Response Format
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": {}
  }
}
```

## Testing Strategy

### Backend Testing
- **Unit Tests** for business logic
- **Integration Tests** for API endpoints
- **Security Tests** for vulnerabilities
- **Performance Tests** for scalability

### Frontend Testing
- **Component Tests**
- **Integration Tests**
- **E2E Tests**
- **Security Tests**

## Deployment

### Environment Configuration
```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=taskmanager
DB_PASSWORD=strongpass123
DB_NAME=taskmanager
DB_SSLMODE=disable

# JWT
JWT_SECRET=your-secret-key
```

### Production Considerations
- **Environment Variables** for configuration
- **Logging** for monitoring
- **Metrics** for performance
- **Backup** strategy
- **Scaling** considerations

## Development Setup

### Prerequisites
- Go 1.23+
- PostgreSQL 17
- Node.js 18+
- Docker (optional)

### Local Development
1. Clone the repository
2. Set up environment variables
3. Start the database
4. Run migrations
5. Start the backend server
6. Start the frontend development server

### Database Setup
```bash
# Using Docker
docker run -itd --name postgres \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=strongpass123 \
  -e POSTGRES_USER=taskmanager \
  -e POSTGRES_DB=taskmanager \
  postgres:17
```

### Running the Application
```bash
# Backend
cd backend
go run main.go

# Frontend
cd frontend
npm install
npm start
```

## Contributing

### Code Style
- Follow Go best practices
- Use meaningful variable names
- Document public APIs
- Write unit tests

### Git Workflow
1. Create feature branch
2. Write tests
3. Implement feature
4. Submit PR
5. Code review
6. Merge to main

## License
MIT License - see LICENSE file for details


