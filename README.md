# Google Secret Manager Emulator

A production-ready HTTP server that emulates the Google Secret Manager API for local development. This emulator provides a complete implementation of the Secret Manager REST API, enabling seamless local development without requiring actual Google Cloud resources.

## Features

- **Complete API Coverage**: Full implementation of Google Secret Manager REST API
- **Production Parity**: Exact error response formats and HTTP status codes matching Google Cloud
- **Local Development**: Run entirely offline with no Google Cloud dependencies
- **Persistent Storage**: Optional JSON file persistence for data across restarts
- **Docker Support**: Production-ready container with health checks
- **Mock Authentication**: Configurable authentication bypass for development
- **CORS Support**: Enable cross-origin requests for web applications
- **Graceful Shutdown**: Proper cleanup and data persistence on shutdown

## Quick Start

### Using Docker Compose (Recommended)

1. Clone the repository:

```bash
git clone https://github.com/charlesgreen/gsm.git
cd gsm
```

1. Start the emulator:

```bash
docker-compose up -d
```

The emulator will be available at `http://localhost:8085`

### Using Go

1. Install dependencies:

```bash
go mod download
```

1. Run the server:

```bash
go run cmd/server/main.go
```

### Using Docker

```bash
docker build -t gsm-emulator .
docker run -p 8085:8085 gsm-emulator
```

## API Documentation

The emulator implements the complete Google Secret Manager REST API v1:

### Health Checks

- `GET /health` - Health check endpoint
- `GET /ready` - Readiness check endpoint

### Secret Management

- `POST /v1/projects/{project}/secrets` - Create a new secret
- `GET /v1/projects/{project}/secrets` - List secrets in a project
- `GET /v1/projects/{project}/secrets/{secret}` - Get secret metadata
- `DELETE /v1/projects/{project}/secrets/{secret}` - Delete a secret

### Secret Versions

- `POST /v1/projects/{project}/secrets/{secret}:addVersion` - Add a new version
- `GET /v1/projects/{project}/secrets/{secret}/versions/{version}:access` - Access secret data
- `GET /v1/projects/{project}/secrets/{secret}/versions` - List versions
- `DELETE /v1/projects/{project}/secrets/{secret}/versions/{version}` - Delete a version

### Example API Usage

#### Create a Secret

```bash
curl -X POST http://localhost:8085/v1/projects/my-project/secrets \
  -H "Content-Type: application/json" \
  -d '{
    "secretId": "my-secret",
    "secret": {
      "labels": {"env": "development"}
    }
  }'
```

#### Add Secret Version

```bash
curl -X POST http://localhost:8085/v1/projects/my-project/secrets/my-secret:addVersion \
  -H "Content-Type: application/json" \
  -d '{
    "payload": {
      "data": "bXktc2VjcmV0LXZhbHVl"
    }
  }'
```

#### Access Secret Value

```bash
curl http://localhost:8085/v1/projects/my-project/secrets/my-secret/versions/latest:access
```

## Production Parity

This emulator is designed to provide **exact production parity** with Google Cloud Secret Manager, ensuring that applications behave identically in development and production environments.

### Error Response Format

All error responses match the exact Google Cloud API error format:

```json
{
  "error": {
    "code": 404,
    "message": "Secret [projects/my-project/secrets/my-secret] not found.",
    "status": "NOT_FOUND"
  }
}
```

### HTTP Status Code Compliance

The emulator returns the same HTTP status codes as production:

- **200 OK**: Successful GET requests
- **201 Created**: Successful resource creation
- **204 No Content**: Successful DELETE requests
- **400 Bad Request (INVALID_ARGUMENT)**: Invalid request format or missing required fields
- **404 Not Found (NOT_FOUND)**: Resource doesn't exist
- **409 Conflict (ALREADY_EXISTS)**: Attempting to create existing resource
- **500 Internal Server Error (INTERNAL)**: Server-side processing errors

### Resource Name Formatting

Error messages include complete resource paths matching Google Cloud conventions:

- Secrets: `projects/{project}/secrets/{secret}`
- Versions: `projects/{project}/secrets/{secret}/versions/{version}`

### Testing Production Parity

Run the included validation script to verify production parity:

```bash
# Start the emulator
go run cmd/server/main.go &

# Run parity validation tests
./scripts/validate_parity.sh

# Or run the Go test suite
go test ./tests/integration/production_parity_test.go -v
```

The parity tests validate:

- ✅ Correct HTTP status codes for all scenarios
- ✅ Proper JSON error response structure
- ✅ Resource-specific error messages with full paths
- ✅ Consistent error handling across all endpoints
- ✅ Transaction rollback behavior matching production

### Benefits of Production Parity

1. **Reliable Testing**: Local tests accurately predict production behavior
2. **Debugging Consistency**: Same error responses help identify issues early
3. **Reduced Risk**: Eliminates deployment surprises from behavioral differences
4. **Seamless Integration**: Drop-in replacement for production GSM during development

## Configuration

Configure the emulator using environment variables:

| Variable           | Default   | Description                       |
| ------------------ | --------- | --------------------------------- |
| `GSM_PORT`         | `8085`    | Server port                       |
| `GSM_HOST`         | `0.0.0.0` | Bind address                      |
| `GSM_STORAGE_FILE` | _(none)_  | JSON file for persistence         |
| `GSM_LOG_LEVEL`    | `info`    | Log level (debug/info/warn/error) |
| `GSM_ENABLE_CORS`  | `true`    | Enable CORS headers               |
| `GSM_ENABLE_AUTH`  | `false`   | Enable mock authentication        |

## Integration with Go Applications

### Using the Official Google Cloud Client

The emulator is compatible with the official Google Cloud Secret Manager client library. Simply override the endpoint:

```go
package main

import (
    "context"
    "fmt"
    
    secretmanager "cloud.google.com/go/secretmanager/apiv1"
    "google.golang.org/api/option"
)

func main() {
    ctx := context.Background()
    
    // Use emulator endpoint
    client, err := secretmanager.NewClient(ctx, 
        option.WithEndpoint("http://localhost:8085"),
        option.WithoutAuthentication(),
    )
    if err != nil {
        panic(err)
    }
    defer client.Close()
    
    // Use client normally
    // ... your secret manager operations
}
```

### Environment-Based Configuration

```go
func newSecretManagerClient(ctx context.Context) (*secretmanager.Client, error) {
    if emulatorHost := os.Getenv("SECRET_MANAGER_EMULATOR_HOST"); emulatorHost != "" {
        return secretmanager.NewClient(ctx,
            option.WithEndpoint("http://"+emulatorHost),
            option.WithoutAuthentication(),
        )
    }
    
    // Production: use default authentication
    return secretmanager.NewClient(ctx)
}
```

## Integration with Firebase Emulators

Add to your existing Firebase `docker-compose.yml`:

```yaml
version: '3.8'

services:
  # Your existing Firebase emulators...
  firebase-emulator:
    # ... existing configuration
    
  secret-manager-emulator:
    image: charlesgreen/gsm:latest
    ports:
      - "8085:8085"
    environment:
      - GSM_STORAGE_FILE=/app/data/secrets.json
    volumes:
      - ./emulator-data/secrets:/app/data
    networks:
      - firebase-network

networks:
  firebase-network:
    external: true
```

## Development

### Prerequisites

- Go 1.22 or later
- Docker (optional)

### Building

```bash
# Build binary
go build -o bin/gsm-server cmd/server/main.go

# Build Docker image
docker build -t gsm-emulator .
```

### Testing

```bash
# Run unit tests
go test ./tests/unit/...

# Run integration tests
go test ./tests/integration/...

# Run production parity tests
go test ./tests/integration/production_parity_test.go -v

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Validate production parity with shell script
./scripts/validate_parity.sh
```

### Code Quality

The project uses [golangci-lint](https://golangci-lint.run/) for comprehensive code linting and quality checks.

#### Install golangci-lint

```bash
# Install using go
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Or using brew (macOS)
brew install golangci-lint

# Or using curl (Linux/macOS)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
```

#### Run linting

```bash
# Run all configured linters
golangci-lint run

# Run with verbose output
golangci-lint run -v

# Run and fix auto-fixable issues
golangci-lint run --fix

# Run only specific linters
golangci-lint run --enable-only=errcheck,govet
```

#### Linting Configuration

The project includes a `.golangci.yml` configuration file with the following linters enabled:

- **errcheck**: Check for unchecked errors
- **govet**: Vet examines Go source code and reports suspicious constructs
- **staticcheck**: Advanced Go linter
- **revive**: Fast, configurable, extensible, flexible Go linter
- **gocritic**: Highly opinionated Go linter
- **prealloc**: Find slice declarations with non-zero initial length
- **whitespace**: Whitespace linter
- And more...

The linting runs automatically in CI/CD pipelines to ensure code quality.

### Project Structure

```bash
├── cmd/server/          # Main application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/    # HTTP request handlers
│   │   ├── middleware/  # HTTP middleware
│   │   └── routes/      # Route configuration
│   ├── models/          # Data models and structures
│   └── storage/         # Storage implementations
├── pkg/client/          # Example client code
├── tests/               # Test files
├── Dockerfile           # Container configuration
├── docker-compose.yml   # Development setup
└── README.md           # This file
```

## Production Deployment

### Docker

The provided Dockerfile creates a minimal, secure image:

- Uses multi-stage build for small image size
- Runs as non-root user for security
- Includes health checks
- Supports volume mounts for persistence

```bash
# Build and push to registry
docker build -t your-registry/gsm-emulator:latest .
docker push your-registry/gsm-emulator:latest

# Deploy
docker run -d \
  --name secret-manager-emulator \
  -p 8085:8085 \
  -v ./secrets:/app/data \
  -e GSM_STORAGE_FILE=/app/data/secrets.json \
  your-registry/gsm-emulator:latest
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: secret-manager-emulator
spec:
  replicas: 1
  selector:
    matchLabels:
      app: secret-manager-emulator
  template:
    metadata:
      labels:
        app: secret-manager-emulator
    spec:
      containers:
      - name: gsm-emulator
        image: charlesgreen/gsm:latest
        ports:
        - containerPort: 8085
        env:
        - name: GSM_STORAGE_FILE
          value: /app/data/secrets.json
        volumeMounts:
        - name: storage
          mountPath: /app/data
        livenessProbe:
          httpGet:
            path: /health
            port: 8085
        readinessProbe:
          httpGet:
            path: /ready
            port: 8085
      volumes:
      - name: storage
        persistentVolumeClaim:
          claimName: gsm-storage
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Write tests for new functionality
- Follow Go conventions and best practices
- Update documentation for API changes
- Ensure Docker builds successfully
- Test with the official Google Cloud client library

## Troubleshooting

### Common Issues

#### Emulator not starting

- Check port 8085 is available
- Verify Docker daemon is running
- Check logs: `docker-compose logs secret-manager-emulator`

#### Authentication errors

- Ensure `GSM_ENABLE_AUTH=false` for development
- Use `option.WithoutAuthentication()` in client code
- Verify endpoint URL is correct

#### Data persistence issues

- Check volume mount configuration
- Verify write permissions on storage directory
- Ensure `GSM_STORAGE_FILE` path is accessible

#### Connection refused

- Verify emulator is running: `curl http://localhost:8085/health`
- Check firewall/network configuration
- Ensure correct host and port binding

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Google Cloud Secret Manager API documentation
- The Go community for excellent libraries and tools
- Contributors and users of this project

---

For more examples and detailed API documentation, see the [examples](pkg/client/) directory.
