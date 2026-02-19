# Knowledge Base References for BlytzCloud

## Framework References

### Go (1.26.0)
| Resource | URL |
|----------|-----|
| **Release Notes** | https://go.dev/doc/go1.26 |
| **Go 1.25 Release Notes** | https://go.dev/doc/go1.25 |
| **Go 1.24 Release Notes** | https://go.dev/doc/go1.24 |
| **Standard Library** | https://pkg.go.dev/std |
| **Effective Go** | https://go.dev/doc/effective_go |
| **Go Modules Reference** | https://go.dev/ref/mod |
| **GitHub Repository** | https://github.com/golang/go |

### Gin Web Framework (v1.11.0)
| Resource | URL |
|----------|-----|
| **GitHub Repository** | https://github.com/gin-gonic/gin |
| **Release v1.11.0** | https://github.com/gin-gonic/gin/releases/tag/v1.11.0 |
| **Documentation** | https://gin-gonic.com/docs/ |
| **Go Package Docs** | https://pkg.go.dev/github.com/gin-gonic/gin@v1.11.0 |
| **Examples** | https://github.com/gin-gonic/examples |

### modernc.org/sqlite (v1.46.1)
| Resource | URL |
|----------|-----|
| **GitLab Repository** | https://gitlab.com/cznic/sqlite |
| **Go Package Docs** | https://pkg.go.dev/modernc.org/sqlite@v1.46.1 |
| **SQLite Documentation** | https://www.sqlite.org/docs.html |
| **Virtual Tables (vtab)** | https://pkg.go.dev/modernc.org/sqlite/vtab |

### Stripe Go SDK (v84.3.0)
| Resource | URL |
|----------|-----|
| **GitHub Repository** | https://github.com/stripe/stripe-go |
| **Release v84.3.0** | https://github.com/stripe/stripe-go/releases/tag/v84.3.0 |
| **API Reference** | https://pkg.go.dev/github.com/stripe/stripe-go/v84 |
| **Stripe Docs** | https://stripe.com/docs/api |
| **Migration Guide** | https://github.com/stripe/stripe-go/blob/master/CHANGELOG.md |

### Zap Logging (v1.27.1)
| Resource | URL |
|----------|-----|
| **GitHub Repository** | https://github.com/uber-go/zap |
| **Release v1.27.1** | https://github.com/uber-go/zap/releases/tag/v1.27.1 |
| **Documentation** | https://pkg.go.dev/go.uber.org/zap@v1.27.1 |
| **FAQ & Guide** | https://github.com/uber-go/zap/blob/master/FAQ.md |

### Docker SDK for Go
| Resource | URL |
|----------|-----|
| **GitHub Repository** | https://github.com/docker/docker |
| **Go Package Docs** | https://pkg.go.dev/github.com/docker/docker/api |
| **Docker API Docs** | https://docs.docker.com/engine/api/ |
| **SDK Examples** | https://docs.docker.com/engine/api/sdk/ |

### Caddy (Reverse Proxy)
| Resource | URL |
|----------|-----|
| **Official Website** | https://caddyserver.com/ |
| **Documentation** | https://caddyserver.com/docs/ |
| **GitHub Repository** | https://github.com/caddyserver/caddy |
| **Caddyfile Docs** | https://caddyserver.com/docs/caddyfile |

## Quick Reference Commands

```bash
# Check installed versions
go version
go list -m all

# View package documentation locally
go doc github.com/gin-gonic/gin
go doc modernc.org/sqlite
go doc github.com/stripe/stripe-go/v84

# View specific function/method docs
go doc github.com/gin-gonic/gin.Context.JSON
go doc modernc.org/sqlite.Open
```

## Compatibility Matrix

### Go Version Compatibility (Go 1.26.0)

| Package | Version | Min Go Version | Status | Notes |
|---------|---------|----------------|--------|-------|
| **Gin** | v1.11.0 | Go 1.23+ | ✅ Compatible | Requires Go 1.23 minimum, tested with 1.26 |
| **SQLite** | v1.46.1 | Go 1.20+ | ✅ Compatible | CGO-free, pure Go implementation |
| **Docker SDK** | v29.2.1 | Go 1.22+ | ✅ Compatible | Latest stable, security fixes |
| **Stripe** | v84.3.0 | Go 1.18+ | ✅ Compatible | Latest stable major version |
| **Zap** | v1.27.1 | Go 1.18+ | ✅ Compatible | Uber's logging library |
| **UUID** | v1.6.0 | Go 1.18+ | ✅ Compatible | Latest with Max UUID support |
| **godotenv** | v1.5.1 | Go 1.16+ | ✅ Compatible | Environment variable loader |

### Known Issues & Security Advisories

#### Docker SDK v29.2.1
✅ **All security issues resolved** - This is the latest stable version with all known vulnerabilities patched.

**Key fixes since v25.0.0**:
- Fixed data exfiltration vulnerability (GO-2024-2659)
- Fixed firewalld network isolation issue (GO-2025-3829)
- BuildKit v0.27.1 integration
- Fixed encrypted overlay networks issues
- Various daemon stability improvements

#### SQLite v1.46.1
✅ **No known issues** - This is the latest stable version with SQLite 3.51.2 support.

#### UUID v1.6.0
✅ **Latest stable** - Added Max UUID constant and fixed UUIDv7 monotonicity issues.

**Changes from v1.5.0**:
- Added Max UUID constant support
- Fixed monotonicity issues in UUIDv7 generation
- Documentation fixes

### Inter-Package Compatibility

All packages listed above are **mutually compatible** with Go 1.26.0. No version conflicts detected.

### Testing Compatibility

```bash
# Verify all dependencies work together
go mod tidy
go mod verify
go test ./...

# Check for known vulnerabilities
go list -json -m all | nancy sleuth
# or
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## Migration Resources

- **Go 1.22 → 1.26**: https://go.dev/doc/go1.26
- **Gin v1.9 → v1.11**: Check release notes for HTTP/3 and binding changes
- **Stripe v76 → v84**: Review CHANGELOG for API version updates
- **SQLite v1.28 → v1.46**: Major version jump - check virtual table API changes
