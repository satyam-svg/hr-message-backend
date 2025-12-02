# HR Backend API

A professional HR backend system built with Go, Fiber, and Prisma.

## Features

✅ **Authentication System**
- User signup with email and password
- User login with JWT token generation
- Password hashing with bcrypt
- JWT-based authentication middleware
- Concurrent password hashing and token generation

✅ **Clean Architecture**
- Models: Request/Response DTOs
- Services: Business logic with concurrency
- Handlers: HTTP request handlers
- Middleware: JWT authentication
- Utils: Password hashing and JWT utilities
- Routes: API endpoint definitions

## Tech Stack

- **Framework**: Fiber (Go web framework)
- **Database**: PostgreSQL with Prisma ORM
- **Authentication**: JWT tokens
- **Password Hashing**: bcrypt
- **Concurrency**: Go channels and goroutines

## Project Structure

```
hr_backend/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internals/
│   ├── models/
│   │   └── user.go              # User DTOs
│   ├── services/
│   │   └── auth_service.go      # Authentication business logic
│   ├── handlers/
│   │   └── auth_handler.go      # HTTP handlers
│   ├── middleware/
│   │   └── auth_middleware.go   # JWT middleware
│   ├── utils/
│   │   ├── jwt.go               # JWT utilities
│   │   └── password.go          # Password hashing
│   └── routes/
│       └── auth_routes.go       # Route definitions
├── prisma/
│   └── schema.prisma            # Database schema
├── .env                         # Environment variables
└── go.mod                       # Go dependencies
```

## Setup

### 1. Clone and Install Dependencies

```bash
cd hr_backend
go mod download
```

### 2. Configure Environment Variables

Copy `.env.example` to `.env` and update the values:

```bash
cp .env.example .env
```

Update the following variables in `.env`:
- `DATABASE_URL`: Your PostgreSQL connection string
- `JWT_SECRET`: A strong secret key (min 32 characters)
- `PORT`: Server port (default: 3000)

### 3. Generate Prisma Client

```bash
go run github.com/steebchen/prisma-client-go generate
```

### 4. Run Database Migrations

```bash
npx prisma migrate dev --name init
```

### 5. Start the Server

```bash
go run cmd/api/main.go
```

The server will start on `http://localhost:3000`

## API Endpoints

### Authentication

#### 1. Signup
**POST** `/api/auth/signup`

Create a new user account.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123",
  "name": "John Doe"
}
```

**Response (201 Created):**
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "createdAt": "2025-11-28T12:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request body or validation error
- `409 Conflict`: User with email already exists
- `500 Internal Server Error`: Server error

#### 2. Login
**POST** `/api/auth/login`

Authenticate a user and get a JWT token.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123"
}
```

**Response (200 OK):**
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "createdAt": "2025-11-28T12:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request body
- `401 Unauthorized`: Invalid credentials
- `500 Internal Server Error`: Server error

### Health Check

#### GET `/`
Returns API status.

**Response:**
```json
{
  "status": "ok",
  "message": "HR Backend API is running"
}
```

#### GET `/health`
Returns API and database health status.

**Response:**
```json
{
  "status": "ok",
  "database": "connected"
}
```

## Testing with cURL

### Signup
```bash
curl -X POST http://localhost:3000/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test@1234",
    "name": "Test User"
  }'
```

### Login
```bash
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test@1234"
  }'
```

### Protected Route (Example)
```bash
curl -X GET http://localhost:3000/api/protected \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Using JWT Middleware

To protect routes with JWT authentication:

```go
import (
    "github.com/satyam-svg/hr_backend/internals/middleware"
)

// In your route setup
protected := app.Group("/api/protected")
protected.Use(middleware.AuthRequired())

protected.Get("/profile", func(c *fiber.Ctx) error {
    userId := c.Locals("userId").(string)
    userEmail := c.Locals("userEmail").(string)
    
    return c.JSON(fiber.Map{
        "userId": userId,
        "email": userEmail,
    })
})
```

## Concurrency Model

The authentication system uses Go's concurrency features for better performance:

### Password Hashing
```go
// Hash password in goroutine
hashChan := make(chan result)
go func() {
    hash, err := utils.HashPassword(req.Password)
    hashChan <- result{data: hash, err: err}
}()

// Wait for result
hashResult := <-hashChan
```

### Token Generation
```go
// Generate token concurrently
tokenChan := make(chan result)
go func() {
    token, err := utils.GenerateToken(user.ID, user.Email)
    tokenChan <- result{data: token, err: err}
}()

// Wait for result
tokenResult := <-tokenChan
```

## Security Features

- ✅ Password hashing with bcrypt (cost factor: 10)
- ✅ JWT token-based authentication
- ✅ Token expiration (24 hours by default)
- ✅ CORS enabled
- ✅ Input validation
- ✅ Error handling with proper HTTP status codes

## Build for Production

```bash
go build -o hr_backend ./cmd/api
./hr_backend
```

## License

MIT
