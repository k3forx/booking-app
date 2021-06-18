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

## After all migrations are completed

There are several tables in `booking` database.

```bash
❯ docker exec -it postgres bash
root@b983aecf24c2:/# psql -U root
psql (13.3 (Debian 13.3-1.pgdg100+1))
Type "help" for help.

root=# \c booking;
You are now connected to database "booking" as user "root".
booking=# \d+;
                                       List of relations
 Schema |           Name           |   Type   | Owner | Persistence |    Size    | Description
--------+--------------------------+----------+-------+-------------+------------+-------------
 public | reservations             | table    | root  | permanent   | 8192 bytes |
 public | reservations_id_seq      | sequence | root  | permanent   | 8192 bytes |
 public | restrictions             | table    | root  | permanent   | 0 bytes    |
 public | restrictions_id_seq      | sequence | root  | permanent   | 8192 bytes |
 public | room_restrictions        | table    | root  | permanent   | 0 bytes    |
 public | room_restrictions_id_seq | sequence | root  | permanent   | 8192 bytes |
 public | rooms                    | table    | root  | permanent   | 0 bytes    |
 public | rooms_id_seq             | sequence | root  | permanent   | 8192 bytes |
 public | schema_migration         | table    | root  | permanent   | 8192 bytes |
 public | users                    | table    | root  | permanent   | 8192 bytes |
 public | users_id_seq             | sequence | root  | permanent   | 8192 bytes |
(11 rows)

booking=# exit
root@b983aecf24c2:/# exit
exit
```

### Re-create all tables

```bash
❯ soda reset
```
