# servicewait

Smart Service Dependency Waiter - waits for services to become available.

## Purpose

Wait for dependent services to be ready before proceeding with application startup. Useful for container orchestration and service dependencies.

## Installation

```bash
go build -o servicewait ./cmd/servicewait
```

## Usage

```bash
servicewait <service1> <service2> ...
```

Format: `name:host[:port][protocol][endpoint]`

### Examples

```bash
# Wait for TCP service
servicewait db:localhost:5432:tcp

# Wait for HTTP service with endpoint
servicewait api:localhost:8080:http:/health

# Wait for multiple services
servicewait db:localhost:5432:tcp redis:localhost:6379:tcp api:localhost:8080:http:/ready
```

## Output

```
=== SERVICE DEPENDENCY WAITER ===

Waiting for db (localhost:5432)...
  db is ready (2s)

Waiting for redis (localhost:6379)...
  redis is ready (1s)

Waiting for api (localhost:8080)...
  api is ready (3s)

Summary: 3 services ready, 0 services unavailable
```

## Protocol Types

- tcp: TCP port connection check
- http/https: HTTP/HTTPS endpoint check
- unix: Unix socket connection

## Default Settings

- Timeout: 5 seconds per check
- Max Retries: 30 (total wait time: ~60 seconds)

## Exit Codes

- 0: All services ready
- 1: One or more services unavailable

## Dependencies

- Go 1.21+
- github.com/fatih/color

## Build and Run

```bash
# Build
go build -o servicewait ./cmd/servicewait

# Run
go run ./cmd/servicewait db:localhost:5432:tcp api:localhost:8080:http:/health
```

## Usage in Docker Compose

```yaml
services:
  app:
    image: myapp
    depends_on:
      - db
      - redis
    command: ["servicewait", "db:db:5432:tcp", "redis:redis:6379:tcp", "start"]
```

## License

MIT