# Auth Service - JWT Authentication with Clean Architecture

A secure authentication service built with Go using Clean Architecture and SOLID principles.

## Features

- ✅ JWT-based authentication
- ✅ Access Token (5 minutes expiry) stored in httpOnly cookie
- ✅ Refresh Token (1 day expiry) with database management
- ✅ Session management with device tracking
- ✅ Clean Architecture structure
- ✅ SOLID principles applied

## Project Structure

```
.
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── domain/                 # Enterprise Business Rules
│   │   ├── entity/             # Business entities
│   │   │   ├── user.go
│   │   │   ├── refresh_token.go
│   │   │   └── token_pair.go
│   │   └── repository/         # Repository interfaces
│   │       ├── user_repository.go
│   │       └── refresh_token_repository.go
│   ├── usecase/                # Application Business Rules
│   │   └── auth/
│   │       ├── dto.go          # Data Transfer Objects
│   │       ├── interface.go    # Use case interfaces
│   │       └── usecase.go      # Business logic implementation
│   ├── delivery/               # Interface Adapters (HTTP)
│   │   └── http/
│   │       ├── handler/        # HTTP handlers
│   │       ├── middleware/     # HTTP middleware
│   │       └── router.go       # Route configuration
│   └── infrastructure/         # Frameworks & Drivers
│       ├── config/             # Configuration
│       ├── database/           # Database implementation
│       └── jwt/                # JWT implementation
└── pkg/
    └── response/               # Shared response utilities
```

## Clean Architecture Layers

1. **Domain Layer** (`internal/domain/`)
   - Contains business entities and repository interfaces
   - No dependencies on external packages
   
2. **Use Case Layer** (`internal/usecase/`)
   - Contains application business rules
   - Depends only on domain layer

3. **Delivery Layer** (`internal/delivery/`)
   - Contains HTTP handlers and middleware
   - Depends on use case layer

4. **Infrastructure Layer** (`internal/infrastructure/`)
   - Contains external implementations (database, JWT, config)
   - Implements interfaces defined in domain/usecase layers

## SOLID Principles Applied

- **S**ingle Responsibility: Each component has one reason to change
- **O**pen/Closed: Classes are open for extension, closed for modification
- **L**iskov Substitution: Implementations can substitute interfaces
- **I**nterface Segregation: Small, focused interfaces
- **D**ependency Inversion: High-level modules depend on abstractions

## API Endpoints

### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register new user |
| POST | `/api/auth/login` | Login and get tokens |
| POST | `/api/auth/refresh` | Refresh access token |
| POST | `/api/auth/logout` | Logout current session |

### Protected Endpoints (Requires Authentication)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/auth/me` | Get current user info |
| GET | `/api/auth/sessions` | Get active sessions |
| POST | `/api/auth/logout-all` | Logout from all devices |

### Health Check

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |

## Request/Response Examples

### Register

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'
```

Response:
```json
{
  "success": true,
  "message": "user registered successfully",
  "data": {
    "user": {
      "id": 1,
      "email": "user@example.com",
      "name": "John Doe"
    }
  }
}
```

### Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

Response:
```json
{
  "success": true,
  "message": "login successful",
  "data": {
    "user": {
      "id": 1,
      "email": "user@example.com",
      "name": "John Doe"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "base64-encoded-random-string"
  }
}
```

### Refresh Token

```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -c cookies.txt
```

### Get Current User (Protected)

```bash
curl -X GET http://localhost:8080/api/auth/me \
  -b cookies.txt
```

### Get Active Sessions (Protected)

```bash
curl -X GET http://localhost:8080/api/auth/sessions \
  -b cookies.txt
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | 8080 | HTTP server port |
| `DATABASE_PATH` | ./auth.db | SQLite database path |
| `JWT_SECRET_KEY` | (change-in-production) | JWT signing secret |
| `JWT_ACCESS_TOKEN_EXPIRY` | 5 (minutes) | Access token expiry |
| `JWT_REFRESH_TOKEN_EXPIRY` | 1440 (minutes = 1 day) | Refresh token expiry |

## Security Features

1. **httpOnly Cookies**: Access and refresh tokens stored in httpOnly cookies
2. **Secure Flag**: Cookies marked as Secure (HTTPS only)
3. **SameSite Strict**: Prevents CSRF attacks
4. **Password Hashing**: bcrypt with default cost
5. **Token Rotation**: Refresh tokens are rotated on each use
6. **Session Tracking**: Device info, IP address, and user agent logged

## Database Schema

### Users Table

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Refresh Tokens Table

```sql
CREATE TABLE refresh_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token TEXT UNIQUE NOT NULL,
    device_info TEXT,
    ip_address TEXT,
    user_agent TEXT,
    expires_at DATETIME NOT NULL,
    revoked_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

## Running the Application

```bash
# Build
go build -o bin/auth-server ./cmd/main.go

# Run
./bin/auth-server

# Or run directly
go run ./cmd/main.go
```

## License

MIT
