# E-Commerce-MS

E-Commerce Microservices written in Go

## Getting started

### Environment variables

Create a `.env` file with content:

```env
GOOSE_DBSTRING="host=192.168.1.100 user=postgres password=postgres dbname=ecomm sslmode=disable"
GOOSE_DRIVER="postgres"
GOOSE_MIGRATION_DIR="internal/adapters/postgresql/migrations"
```

### Installing libraries

* Install [SQLC](https://docs.sqlc.dev/en/latest/overview/install.html)

SQLC is a command line tool that generates type-safe code from SQL.

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

* Install [Goose](https://github.com/pressly/goose)

Goose is a database migration tool.

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```
