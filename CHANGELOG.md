# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-08-14

### Initial Release

#### Core Features

- Complete Google Secret Manager REST API v1 implementation
- Production parity with Google Cloud Secret Manager error responses
- In-memory and file-based persistent storage options
- Docker support with health checks
- Mock authentication for development
- CORS support for web applications

#### API Implementation

- Secret CRUD operations (create, read, update, delete)
- Secret version management with data access
- Pagination support for list operations
- Resource metadata with labels and replication settings
- Checksum validation for secret data integrity

#### Production Parity

- Exact error response format matching Google Cloud
- Proper HTTP status codes (200, 201, 204, 400, 404, 409, 500)
- Resource path formatting in error messages (e.g., `projects/{project}/secrets/{secret}`)
- Extended error response models supporting Google's AIP-193 standard

#### Testing & Validation

- Comprehensive unit and integration test suites
- Production parity test suite (`tests/integration/production_parity_test.go`)
- Validation script for testing parity (`scripts/validate_parity.sh`)
- Linting configuration with golangci-lint

#### Documentation

- Complete API documentation with usage examples
- Production parity guidelines
- Docker and Kubernetes deployment examples
- Integration guides for Go applications