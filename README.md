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

## Install `pop` package

https://gobuffalo.io/en/docs/db/getting-started/

> Pop makes it easy to do CRUD operations with basic ORM functionality, run migrations, and build/execute queries.

```bash
❯ go get github.com/gobuffalo/pop/...
```

### Migrations

```bash
❯ soda generate fizz CreateUsersTable
```

Write migration in `...up.fizz` file.

```bash
❯ soda migrate
```

See https://gobuffalo.io/en/docs/db/migrations/ for more details.
