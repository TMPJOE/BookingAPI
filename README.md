# BookingAPI - API Gateway

A high-performance API Gateway built with Go, designed for hotel booking systems.

## Architecture

```
cmd/gateway/main.go          # Application entry point
internal/
├── config/                  # Configuration management
├── logging/                 # Structured logging
└── gateway/
    ├── handlers/            # HTTP request handlers
    ├── middleware/          # Custom middleware (auth, rate limiting, CORS, etc.)
    └── routing/             # Route definitions
config.yaml                  # Configuration file
```

## Features

- **Rate Limiting**: Token bucket algorithm for request throttling
- **Authentication**: Bearer token authentication middleware
- **CORS**: Cross-Origin Resource Sharing support
- **Security Headers**: XSS protection, HSTS, clickjacking prevention
- **Circuit Breaker**: Fault tolerance for upstream services
- **Health Checks**: `/health`, `/ready`, `/live` endpoints
- **Structured Logging**: JSON logging with request tracing

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL (optional, for persistence)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd BookingAPI

# Download dependencies
go mod tidy

# Run the application
go run ./cmd/gateway
```

### Configuration

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Or use `config.yaml` for YAML-based configuration.

### Using Make

```bash
make deps      # Download dependencies
make run       # Run the application
make build     # Build binary
make test      # Run tests
make lint      # Run linter
make clean     # Clean build artifacts
```

## API Endpoints

### Health Checks

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Liveness check |
| `/ready` | GET | Readiness check |
| `/live` | GET | Liveness check |

### API v1 (Requires Authentication)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/bookings` | GET | List all bookings |
| `/api/v1/bookings` | POST | Create a booking |
| `/api/v1/bookings/{id}` | GET | Get a booking |
| `/api/v1/bookings/{id}` | PUT | Update a booking |
| `/api/v1/bookings/{id}` | DELETE | Delete a booking |
| `/api/v1/rooms` | GET | List all rooms |
| `/api/v1/rooms/{id}` | GET | Get a room |
| `/api/v1/guests` | GET | List all guests |
| `/api/v1/guests/{id}` | GET | Get a guest |

### Authentication

Include the Bearer token in the Authorization header:

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/bookings
```

## Configuration Options

### Server

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s
```

### Rate Limiting

```yaml
rate_limit:
  enabled: true
  requests_per_second: 100
  burst: 200
```

### Circuit Breaker

```yaml
circuit_breaker:
  enabled: true
  max_failures: 5
  timeout: 30s
```

## Docker

```bash
# Build
docker build -t booking-api-gateway .

# Run
docker run -p 8080:8080 booking-api-gateway
```

## Testing

```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Project Structure

```
BookingAPI/
├── cmd/
│   └── gateway/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/               # Configuration loading
│   ├── logging/              # Logger setup
│   └── gateway/
│       ├── handlers/         # HTTP handlers
│       ├── middleware/       # Custom middleware
│       └── routing/          # Route definitions
├── config.yaml               # Configuration file
├── Makefile                  # Build automation
├── Dockerfile                # Docker image
├── docker-compose.yml        # Docker Compose
├── .env.example              # Environment template
├── go.mod                    # Go module
└── README.md                 # This file
```

## License

MIT
