# 🌐 API Gateway (BookingAPI)

> Reverse-proxy API gateway that routes, authenticates, and rate-limits all incoming traffic for the Hotel Reservation Platform.

## Overview

The API Gateway (BookingAPI) is the **single entry point** for all client traffic. It acts as a reverse proxy, routing requests to the appropriate upstream microservice based on URL path prefixes. It provides centralized cross-cutting concerns: JWT authentication, rate limiting, CORS, security headers, request logging, and upstream health monitoring.

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25 |
| Router | [go-chi/chi](https://github.com/go-chi/chi) v5 |
| Proxy | `net/http/httputil` (stdlib reverse proxy) |
| Auth | JWT verification (RSA-256 public key) |
| Container | Docker (multi-stage Alpine build) |

## Architecture

```
cmd/
└── gateway/
    └── main.go           # Application entrypoint
internal/
├── config/               # YAML config loader
├── logging/              # Structured slog logger
└── gateway/
    ├── handlers/         # Health check handlers
    │   └── handler.go
    ├── health/           # Periodic upstream health checker
    │   └── health.go
    ├── middleware/        # Middleware stack
    │   ├── auth.go       # JWT authentication
    │   ├── cors.go       # CORS headers
    │   ├── logging.go    # Request logging
    │   ├── ratelimit.go  # Token bucket rate limiter
    │   ├── requestid.go  # X-Request-ID injection
    │   └── security.go   # Security headers (CSP, HSTS, etc.)
    ├── proxy/            # Reverse proxy engine
    │   └── proxy.go
    └── routing/          # Route configuration
        └── router.go
config.yaml               # Gateway configuration
Dockerfile
go.mod
```

## Routing Table

The gateway routes requests based on **longest path prefix match**:

| Path Prefix | Upstream Service | Auth Required |
|---|---|---|
| `/api/v1/users/*` | `http://user-service:8080` | ❌ (public registration/login) |
| `/api/v1/hotels/*` | `http://hotel-service:8080` | ✅ |
| `/api/v1/rooms/*` | `http://rooms-service:8080` | ✅ |
| `/api/v1/bookings/*` | `http://booking-service:8080` | ✅ |
| `/api/v1/*` (catch-all) | `http://bff-service:8080` | ✅ |
| `/health` | Self (gateway) | ❌ |
| `/ready` | Self (gateway) | ❌ |
| `/live` | Self (gateway) | ❌ |
| `/upstreams` | Self (returns upstream list) | ❌ |

> **Note**: The `/api/v1/users/*` route is **unprotected** to allow registration and login without a JWT.

## Flow Diagram

```mermaid
flowchart TD
    A["Client Request"] --> B["Global Middleware Stack"]
    
    subgraph "Middleware Pipeline"
        B --> B1["Recoverer (panic recovery)"]
        B1 --> B2["RequestID"]
        B2 --> B3["RealIP"]
        B3 --> B4["Timeout"]
        B4 --> B5["Request Logger"]
        B5 --> B6["Rate Limiter"]
        B6 --> B7["CORS"]
        B7 --> B8["Security Headers"]
    end
    
    B8 --> C{"Path Match?"}
    
    C -->|"/health, /ready, /live"| D["Health Handlers"]
    C -->|"/upstreams"| E["Return Upstream List"]
    C -->|"/api/v1/users/*"| F["Reverse Proxy (no auth)"]
    C -->|"/api/v1/*"| G["JWT Auth Middleware"]
    C -->|No Match| H["404 Not Found"]
    
    G --> G1{"Token Valid?"}
    G1 -->|No| G2["401 Unauthorized"]
    G1 -->|Yes| I["Reverse Proxy"]
    
    subgraph "Reverse Proxy"
        I --> I1["Find Upstream by Prefix"]
        I1 --> I2["Strip Path Prefix"]
        I2 --> I3["Forward to Upstream"]
        I3 --> I4["Return Upstream Response"]
    end
    
    F --> I1
```

## Use Case Diagram

```mermaid
graph LR
    subgraph Actors
        Client["🖥️ Frontend Client"]
        Dev["👨‍💻 Developer"]
    end
    
    subgraph "API Gateway"
        UC1["Route to Upstream"]
        UC2["Authenticate Request"]
        UC3["Rate Limit Traffic"]
        UC4["Log Requests"]
        UC5["Health Check"]
        UC6["CORS Preflight"]
        UC7["Monitor Upstreams"]
    end
    
    Client --> UC1
    Client --> UC2
    Client --> UC3
    Client --> UC6
    Dev --> UC5
    Dev --> UC7
```

## State Diagram

```mermaid
stateDiagram-v2
    [*] --> Starting
    Starting --> Ready : All upstreams healthy
    Starting --> Degraded : Some upstreams unhealthy
    Ready --> Degraded : Upstream health check fails
    Degraded --> Ready : Upstream recovers
    Ready --> ShuttingDown : SIGTERM/SIGINT
    Degraded --> ShuttingDown : SIGTERM/SIGINT
    ShuttingDown --> [*] : Graceful shutdown (30s timeout)
    
    state Ready {
        [*] --> Routing
        Routing --> Proxying : Match found
        Proxying --> Routing : Response sent
    }
```

## Package Diagram

```mermaid
graph TB
    subgraph "cmd/gateway"
        Main["main.go"]
    end
    
    subgraph "internal"
        subgraph "config"
            Config["config.go"]
        end
        
        subgraph "logging"
            Logger["logger.go"]
        end
        
        subgraph "gateway"
            subgraph "routing"
                Router["router.go"]
            end
            
            subgraph "proxy"
                RevProxy["proxy.go"]
            end
            
            subgraph "middleware"
                Auth["auth.go"]
                CORSmw["cors.go"]
                Logging["logging.go"]
                RateLimit["ratelimit.go"]
                ReqID["requestid.go"]
                Security["security.go"]
            end
            
            subgraph "handlers"
                Handlers["handler.go"]
            end
            
            subgraph "health"
                HealthCheck["health.go"]
            end
        end
    end
    
    Main --> Config
    Main --> Logger
    Main --> Router
    
    Router --> Handlers
    Router --> RevProxy
    Router --> Auth
    Router --> CORSmw
    Router --> Logging
    Router --> RateLimit
    Router --> ReqID
    Router --> Security
    Router --> HealthCheck
    
    RevProxy --> Config
```

## Middleware Stack (Order)

1. **Recoverer** — Catches panics, returns 500
2. **RequestID** — Injects `X-Request-ID` header
3. **RealIP** — Extracts real client IP from proxy headers
4. **Timeout** — Enforces `read_timeout` from config
5. **Request Logger** — Logs method, path, remote_addr
6. **Rate Limiter** — Token bucket (100 req/s, burst 200)
7. **CORS** — Permissive cross-origin policy
8. **Security Headers** — CSP, HSTS, X-Frame-Options, etc.
9. **JWT Auth** — (only on `/api/v1/*` routes, except users)

## Configuration

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s

rate_limit:
  enabled: true
  requests_per_second: 100
  burst: 200

upstreams:
  - name: "users-service"
    url: "http://user-service:8080"
    path_prefix: "/api/v1/users"
    timeout: 10s
    health_path: "/health"
  # ... more upstreams
```

## Port Mapping

| Context | Port |
|---|---|
| Internal (container) | `8080` |
| External (host) | `8080` |
