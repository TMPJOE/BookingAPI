# BookingAPI - API Gateway

A high-performance API Gateway built with Go, designed for hotel booking systems. The gateway provides routing, authentication, rate limiting, and health monitoring for microservices.

## Architecture

```
cmd/gateway/main.go # Application entry point
internal/
├── config/ # Configuration management
├── logging/ # Structured logging
└── gateway/
    ├── handlers/ # HTTP request handlers
    ├── health/ # Upstream health monitoring
    ├── middleware/ # Custom middleware (auth, rate limiting, CORS, security, logging, request ID)
    ├── proxy/ # Reverse proxy for upstream routing
    └── routing/ # Route definitions using chi router
config.yaml # Configuration file
```

## Features

- **Rate Limiting**: Token bucket algorithm for request throttling
- **Authentication**: Bearer token authentication middleware
- **CORS**: Cross-Origin Resource Sharing support
- **Security Headers**: XSS protection, HSTS, clickjacking prevention, CSP, and more
- **Circuit Breaker**: Fault tolerance for upstream services
- **Health Checks**: `/health`, `/ready`, `/live` endpoints with upstream monitoring
- **Upstream Status**: `/upstreams` endpoint to view all upstream service statuses
- **Structured Logging**: JSON logging with request tracing using slog
- **Reverse Proxy**: Path-based routing to upstream microservices
- **Request ID**: Automatic request ID generation and propagation
- **Real IP**: Client IP extraction from proxy headers

## Quick Start

### Prerequisites

- Go 1.25+
- Upstream microservices (configured in `config.yaml`)

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

Or using Make:

```bash
make deps # Download dependencies
make run  # Run the application
```

### Configuration

Edit `config.yaml` for YAML-based configuration:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s

upstreams:
  - name: "bookings-service"
    url: "http://localhost:8084"
    path_prefix: "/api/v1/bookings"
    timeout: 10s
    health_path: "/health"
```

## Using Make

```bash
make deps         # Download dependencies
make run          # Run the application
make build        # Build binary
make test         # Run tests with race detection
make test-coverage # Run tests with coverage report
make lint         # Run linter
make fmt          # Format code
make tidy         # Tidy dependencies
make clean        # Clean build artifacts
```

## API Endpoints

### Health Checks

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Liveness check |
| `/ready` | GET | Readiness check |
| `/live` | GET | Liveness check |
| `/upstreams` | GET | View upstream service status |

### API v1 (Requires Authentication)

All `/api/v1/*` requests are proxied to upstream services based on path prefix:

| Path Prefix | Upstream Service |
|-------------|------------------|
| `/api/v1/users` | Users Service |
| `/api/v1/hotels` | Hotels Service |
| `/api/v1/rooms` | Rooms Service |
| `/api/v1/bookings` | Bookings Service |
| `/api/v1/payments` | Payments Service |
| `/api/v1/notifications` | Notifications Service |

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

### Logging

```yaml
logging:
  level: "info"    # debug, info, warn, error
  format: "json"   # json, text
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

### Upstreams

```yaml
upstreams:
  - name: "bookings-service"
    url: "http://localhost:8084"
    path_prefix: "/api/v1/bookings"
    timeout: 10s
    health_path: "/health"
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
# Run all tests with race detection
go test -v -race ./...

# Run with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Project Structure

```
BookingAPI/
├── cmd/
│   └── gateway/
│       └── main.go          # Application entry point
├── internal/
│   ├── config/              # Configuration loading
│   │   └── config.go
│   ├── logging/             # Logger setup
│   │   └── logger.go
│   └── gateway/
│       ├── handlers/        # HTTP handlers
│       │   └── handler.go
│       ├── health/          # Upstream health monitoring
│       │   └── health.go
│       ├── middleware/      # Custom middleware
│       │   ├── auth.go
│       │   ├── cors.go
│       │   ├── logging.go
│       │   ├── ratelimit.go
│       │   ├── requestid.go
│       │   └── security.go
│       ├── proxy/           # Reverse proxy
│       │   └── proxy.go
│       └── routing/         # Route definitions
│           └── router.go
├── config.yaml              # Configuration file
├── Makefile                 # Build automation
├── Dockerfile               # Docker image
├── go.mod                   # Go module
└── README.md                # This file
```

## License

MIT
