# Testing Patterns

**Analysis Date:** 2026-04-23

## Test Framework

**Runner:**
- Go standard testing package
- Config: no dedicated test config file is detected; test behavior is driven by `go test` commands in `.github/workflows/go.yml`

**Assertion Library:**
- Standard library only; no third-party assertion package is detected.

**Run Commands:**
```bash
go test -v -race ./...                                               # Run unit and package tests with race detector
go test -v -race -tags=integration_test ./integration                # Run integration suites for local/remote flows
go test -v -race -tags=integration_test_task ./integration           # Run task-mode integration suite
go test -v -race -tags=integration_test_minio ./integration          # Run MinIO integration suite
go test -v -race -tags=integration_test_sftp ./integration           # Run SFTP integration suite
go test -v -race ./... -coverprofile=coverage.txt -covermode=atomic -timeout=10m  # CI coverage run
go test -bench=. ./internal/toplist                                  # Run benchmark examples such as `internal/toplist/toplist_benchmark_test.go`
```

## Test File Organization

**Location:**
- Unit tests are colocated beside implementation files, for example `action/action_test.go`, `conf/parser_test.go`, `logger/logger_test.go`, and `server/server_test.go`.
- Integration tests live in the dedicated `integration/` package, including `integration/integration_local_test.go`, `integration/integration_remote_test.go`, `integration/integration_task_test.go`, `integration/integration_minio_test.go`, and `integration/integration_sftp_test.go`.
- Benchmarks live beside the code they measure, for example `internal/toplist/toplist_benchmark_test.go` and `internal/clist/clist_benchmark_test.go`.

**Naming:**
- Unit tests follow Go’s default `*_test.go` naming.
- Build-tag-specific suites still use `*_test.go` filenames but add top-of-file build constraints, such as `//go:build integration_test` in `integration/integration_local_test.go` and `//go:build integration_test_minio` in `integration/integration_minio_test.go`.
- Test function names follow `TestXxx` and benchmarks follow `BenchmarkXxx`, for example `TestGenerateAddr` in `server/server_test.go` and `BenchmarkTopList_Add` in `internal/toplist/toplist_benchmark_test.go`.

**Structure:**
```
package/<feature>.go
package/<feature>_test.go
integration/integration_test.go
integration/integration_local_test.go
integration/integration_remote_test.go
integration/integration_task_test.go
integration/integration_minio_test.go
integration/integration_sftp_test.go
```

## Test Structure

**Suite Organization:**
```go
func TestGenerateAddr(t *testing.T) {
	testCases := []struct {
		scheme string
		host   string
		port   int
		expect string
	}{
		{"http", "127.0.0.1", 80, "http://127.0.0.1"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s://%s:%d", tc.scheme, tc.host, tc.port), func(t *testing.T) {
			addr := GenerateAddr(tc.scheme, tc.host, tc.port)
			if addr != tc.expect {
				t.Errorf("expect get %s, but actual get %s", tc.expect, addr)
			}
		})
	}
}
```
Source: `server/server_test.go`

**Patterns:**
- Table-driven tests are the default pattern across the repository, visible in `action/action_test.go`, `server/server_test.go`, `conf/parser_test.go`, and `server/session_test.go`.
- Subtests with `t.Run(...)` are used heavily to isolate cases and label failures.
- Assertions are manual `if` checks followed by `t.Errorf(...)`, not helper libraries.
- Error-path tests use `errors.Is` and `errors.As`, for example `conf/parser_test.go`, `server/session_test.go`, and `encrypt/encrypt_test.go`.

## Mocking

**Framework:**
- None detected.

**Patterns:**
```go
func noServer() result.Result {
	r := result.New()
	r.InitDone()
	r.RegisterNotifyHandler(func(s os.Signal, timeout ...time.Duration) error {
		r.Done()
		return nil
	})
	return r
}
```
Source: `integration/integration_test.go`

- The codebase prefers lightweight fakes, stubs, or real component startup over mock frameworks.
- Logging dependencies are satisfied with real test loggers from `logger.NewTestLogger()` in `logger/logger.go`, used widely in tests such as `encrypt/encrypt_test.go` and `ignore/ignore_test.go`.

**What to Mock:**
- Use small in-process stand-ins for lifecycle or coordination interfaces when a full service is unnecessary, following `noServer()` in `integration/integration_test.go`.
- Prefer real config parsing and concrete package behavior for unit tests, as shown in `conf/parser_test.go` and `action/action_test.go`.

**What NOT to Mock:**
- This repository frequently exercises real file, network, and gRPC behavior in integration tests instead of replacing them with mocks, as shown in `api/api_test.go` and `integration/*_test.go`.

## Fixtures and Factories

**Test Data:**
```go
const (
	jsonConfigPath = "./example/gofs-remote-client.json"
	yamlConfigPath = "./example/gofs-remote-server.yaml"
)
```
Source: `conf/parser_test.go`

```go
const (
	certFile     = "../integration/testdata/cert/cert.pem"
	keyFile      = "../integration/testdata/cert/key.pem"
	taskConfFile = "file://./testdata/tasks.yaml"
)
```
Source: `api/api_test.go`

**Location:**
- Integration fixtures and scenario configs live under `integration/testdata/` and are referenced by helpers in `integration/integration_test.go`.
- Package-local examples and fixture files are used where relevant, such as `conf/example/` referenced by `conf/parser_test.go` and encryption outputs under `encrypt/testdata/` referenced by `encrypt/encrypt_test.go`.

## Coverage

**Requirements:**
- Coverage is collected in CI but no minimum threshold is enforced in visible config.
- Coverage artifacts are produced with `-coverprofile=coverage.txt -covermode=atomic` in `.github/workflows/go.yml`.
- Repository badges in `README.md` indicate Codecov reporting is wired in.

**View Coverage:**
```bash
go test -v -race ./... -coverprofile=coverage.txt -covermode=atomic -timeout=10m
go tool cover -func=coverage.txt
go tool cover -html=coverage.txt
```

## Test Types

**Unit Tests:**
- Unit coverage is broad across packages such as `action`, `auth`, `checksum`, `conf`, `core`, `encrypt`, `flag`, `fs`, `ignore`, `logger`, `progress`, `report`, `result`, `retry`, `server`, and `wait`.
- Unit tests validate parsing, value objects, option constructors, lifecycle helpers, and package-specific logic.

**Integration Tests:**
- Integration testing is a first-class pattern in `integration/`.
- Shared helpers in `integration/integration_test.go` start server/client processes through `cmd.RunWithConfigFile`, wait for readiness with `WaitInit()`, execute scripted file actions through `github.com/no-src/fsctl/command`, then shut processes down cleanly with `Shutdown()` and `Wait()`.
- Feature-specific suites are separated by build tags:
  - `integration/integration_local_test.go` for local-disk scenarios
  - `integration/integration_remote_test.go` for remote server/client flows
  - `integration/integration_task_test.go` for task distribution flows
  - `integration/integration_minio_test.go` for MinIO flows
  - `integration/integration_sftp_test.go` for SFTP flows

**E2E Tests:**
- Not used as a separate framework. The closest equivalent is process-level integration coverage in `integration/*_test.go` and API-level end-to-end behavior in `api/api_test.go`.

## CI Validation

- `.github/workflows/go.yml` runs `go build -v ./...`, race-enabled tests, coverage collection, and tagged integration suites.
- `.github/workflows/go.yml` runs cross-platform on Ubuntu, Windows, and macOS for standard build/test jobs.
- Linux-only preparation steps enable extra integrations:
  - `./scripts/init-env.sh` in `.github/workflows/go.yml`
  - `./scripts/sftp/init-sftp.sh` before SFTP integration tests
  - `./scripts/minio/install-minio.sh` and `./scripts/minio/mount-minio.sh` before MinIO integration tests
- `.github/workflows/govulncheck.yml` runs `./scripts/govulncheck.sh`, which installs and executes `govulncheck ./...`.
- `.github/workflows/codeql.yml` adds CodeQL analysis for Go.
- `.github/workflows/docker.yml` validates Docker build and example container runs.

## Common Patterns

**Async Testing:**
```go
go func() {
	if err := srv.Start(); err != nil {
		t.Errorf("start api server error => %v", err)
	}
}()

for i := 0; i < 3; i++ {
	err = c.Start()
	if err == nil {
		break
	}
	time.Sleep(time.Second * 3)
}
```
Source: `api/api_test.go`

- Async tests commonly use goroutines plus polling or sleeps instead of synchronization helper libraries.
- Integration lifecycle tests rely on `WaitInit()` and `Wait()` from `result.Result` in `result/result.go`.

**Error Testing:**
```go
_, err := NewSessionStore(tc.conn)
if !errors.Is(err, tc.expectErr) {
	t.Errorf("expect to get error [%s], but actual get error [%s]", tc.expectErr, err)
}
```
Source: `server/session_test.go`

## Current Coverage Signals

- Strong signal: many packages include adjacent unit tests, indicating broad baseline coverage.
- Strong signal: critical runtime flows have tagged integration suites in `integration/` and API end-to-end coverage in `api/api_test.go`.
- Strong signal: race detector is enabled in CI for standard and integration test commands via `.github/workflows/go.yml`.
- Moderate signal: benchmark coverage exists for selected internal packages such as `internal/toplist/toplist_benchmark_test.go`.
- Weak signal: no visible per-package coverage thresholds or required minimum percentages are enforced.

## Major Testing Gaps

- No visible repository-local lint gate is coupled to tests; CI quality focuses on build, tests, CodeQL, and vulnerability scanning rather than style enforcement.
- No dedicated fuzz tests are detected in the visible codebase.
- No dedicated golden-file testing helpers are detected; file-heavy behaviors are mostly exercised through imperative integration scripts in `integration/testdata/`.
- Some runtime security behavior remains marked as temporary, such as the TODO around insecure token handling in `api/apiclient/grpc_client.go`, but no focused regression test for the future replacement is visible from the sampled files.
- Session-store tests in `server/session_test.go` actively cover memory mode and invalid Redis parameters, but live Redis-backed behavior is commented out rather than exercised.

---

*Testing analysis: 2026-04-23*
