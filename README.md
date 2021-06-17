# booking-app

## Requirements

```bash
❯ go version
go version go1.16 darwin/amd64
```

## How to run?

```bash
❯ go run ./cmd/web/*.go
Starting application on port :8080
```

## Unit test

There are several ways to execute unit test.

```bash
❯ go test

❯ go test -v ./...  # Execute all tests

❯ go test -v  # Show each test cases

❯ go test -cover  # Show coverage rate

❯ go test -coverprofile=coverage.out && go tool cover -html=coverage.out  # Show covered lines with html file on browser
```
