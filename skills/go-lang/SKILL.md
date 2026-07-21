---
name: go-lang
description: Elite Go 1.26+ microservice SDLC. Covers Google OAuth2 ID token verification, HTTP 429 rate limiting, nil pointer safety, paired positive/negative table-driven TDD with testify/mockery, and 85% coverage.
---

# Go Enterprise Service SDLC Skill

This skill prescribes comprehensive standards for production **Go 1.26+** applications, microservices, and gRPC backends, incorporating **Google OAuth2 Authentication**, **HTTP 429 Rate Limiting**, **CWE Security Hardening**, and **TDD**.

## Prerequisites & Pre-Flight Checklist

1. Go 1.26+ installed locally.
2. Google Cloud OAuth2 Client ID configured in `configs/.env.toml`.
3. Bazelisk installed on system PATH.

## Google OAuth2 Provider Architecture for Go Services

1. **Google ID Token Verification Middleware (Gin / gRPC)**:
   - Verify Google OAuth2 Bearer tokens using `google.golang.org/api/idtoken`:
     ```go
     package auth

     import (
     	"context"
     	"net/http"
     	"strings"
     	"github.com/gin-gonic/gin"
     	"google.golang.org/api/idtoken"
     )

     func GoogleOAuth2Middleware(expectedAudience string) gin.HandlerFunc {
     	return func(c *gin.Context) {
     		authHeader := c.GetHeader("Authorization")
     		token := strings.TrimPrefix(authHeader, "Bearer ")
     		if token == "" {
     			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
     			return
     		}

     		payload, err := idtoken.Validate(context.Background(), token, expectedAudience)
     		if err != nil {
     			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Google OAuth2 token"})
     			return
     		}

     		c.Set("user_email", payload.Claims["email"])
     		c.Set("user_id", payload.Subject)
     		c.Next()
     	}
     }
     ```

## HTTP 429 Rate Limiting & Outbound Backoff Invariants

1. **Inbound Rate Limiting Middleware**:
   - Protect Gin/gRPC listeners against quota abuse using `golang.org/x/time/rate` token buckets. Return HTTP 429 Too Many Requests.
2. **Outbound API Resilience (AI & Cloud Services)**:
   - Outbound HTTP calls to Gemini/Vertex AI MUST use `hashicorp/go-retryablehttp` with exponential backoff and randomized jitter.
3. **429 Negative Table-Driven TDD**:
   - Test suites MUST simulate 429 status codes from external API mocks.

## Security Checkpoints & CWE Invariants

- **CWE-250 (Execution with Unnecessary Privileges)**: Final Docker runtime MUST execute inside `scratch` or `distroless` as non-root (`USER nonroot:nonroot`).
- **CWE-89 (SQL Injection)**: Enforce GORM parameterized query syntax across all database repository methods.
- **CWE-798 (Hardcoded Credentials)**: Decrypt XOR secrets in memory at startup via `modenv`; never log sensitive configuration structs.

## Defensive Error Handling & Nil Pointer Safety Invariants

- **Strict Nil Safety**: Pointers, maps, slices, and interfaces MUST be verified against `nil` before dereferencing or indexing.
- **Defensive Error Wrapping**: Always wrap lower-level errors with operational context using `fmt.Errorf("failed to execute operation: %w", err)`.
- **Paired Positive & Negative TDD**: Table tests MUST assert 200 OK, 400 Bad Request, 429 Rate Limited, and 500 Internal Error paths.

## 3-Phase Execution Protocol

### Phase 1: Package Structure & Google OAuth2 Middleware
Establish `/cmd`, `/internal`, `/pkg`, `/api` layout and configure Google OAuth2 validation.

### Phase 2: Implement Paired Table-Driven TDD & Mocks (85% Coverage)
Generate interface mocks using `mockery` and write table-driven unit tests evaluated via `testify/assert`.

### Phase 3: Static Analysis, Coverage Badges & Hermetic Build
```bash
golangci-lint run ./...
go test -v -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
bazel run //:gazelle
bazel build //...
```

## Progressive Disclosure References

- **Go Architecture Guide**: Read [`references/go_architecture.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/go-lang/references/go_architecture.md).
- **Reference Main**: View [`examples/main.go`](file:///Users/rmcguinness/Projects/skill-builder/skills/go-lang/examples/main.go).
- **Reference Server**: View [`examples/server.go`](file:///Users/rmcguinness/Projects/skill-builder/skills/go-lang/examples/server.go).
